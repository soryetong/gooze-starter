package gzcache

import (
	"hash/maphash"
	"sync"
	"sync/atomic"
	"time"

	"github.com/soryetong/gooze-starter/pkg/gzutil"
	"github.com/spf13/cast"
)

// listNode 是 LRU 链表的节点
type listNode struct {
	key       string
	value     any
	expiresAt time.Time
	prev      *listNode
	next      *listNode
	ttl       time.Duration
}

// shard 是缓存的一个分片，包含局部锁、数据、双向链表和原子计数器
type shard struct {
	mu       sync.Mutex
	capacity int
	items    map[string]*listNode
	head     *listNode
	tail     *listNode
	count    atomic.Int64
	parent   *CacheNode
}

// CacheNode 是完整缓存结构，支持 TTL、LRU 和并发分片
type CacheNode struct {
	shards         []*shard
	shardMask      uint64
	seed           maphash.Seed
	onEvict        func(key string, value any) // 被动淘汰/过期时的回调
	cleanerStop    chan struct{}
	cleanerRunning atomic.Bool
	cleanInterval  time.Duration
}

// defaultShardCount 是默认分片数量（必须为 2 的幂）
const defaultShardCount = 256

// newCache 创建一个新的缓存实例
// capacity: 缓存总容量。如果 capacity <= 0，则容量无限。
// shardCount: 分片数量，必须是 2 的幂。如果不是，将使用 defaultShardCount。
// cleanInterval: 启用定期清理，0 表示不启用。如果当前业务不是需要严格控制内存占用的，不建议开启。
func New(capacity int, shardCount int, cleanInterval time.Duration) *CacheNode {
	count := defaultShardCount
	if isPowerOfTwo(shardCount) {
		count = shardCount
	}

	var shardCap int
	if capacity > 0 {
		shardCap = (capacity + count - 1) / count
	} else {
		shardCap = 0
	}

	c := &CacheNode{
		shards:        make([]*shard, count),
		shardMask:     uint64(count - 1),
		seed:          maphash.MakeSeed(),
		cleanInterval: cleanInterval,
	}

	for i := 0; i < count; i++ {
		c.shards[i] = &shard{
			capacity: shardCap,
			items:    make(map[string]*listNode),
			parent:   c,
			// head 和 tail 初始化为 nil
		}
	}

	if cleanInterval > 0 {
		c.startCleaner()
	}

	return c
}

// isPowerOfTwo 判断一个整数是否为 2 的幂
func isPowerOfTwo(n int) bool {
	return n > 0 && (n&(n-1)) == 0
}

// getShard 根据 key 获取对应的分片
func (c *CacheNode) getShard(key string) *shard {
	var h maphash.Hash
	h.SetSeed(c.seed)
	_, _ = h.WriteString(key)
	return c.shards[h.Sum64()&c.shardMask]
}

// Set 插入或更新键值，带 TTL 过期时间，小于等于 0，则永不过期。
func (c *CacheNode) Set(key string, value any, ttl time.Duration) {
	s := c.getShard(key)
	s.mu.Lock()
	defer s.mu.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	if node, exists := s.items[key]; exists {
		node.value = value
		node.ttl = ttl
		node.expiresAt = expiresAt
		s.moveToHead(node)
	} else {
		// 新节点添加到 map 和链表头部
		newNode := &listNode{
			key:       key,
			value:     value,
			expiresAt: expiresAt,
			ttl:       ttl,
		}
		s.items[key] = newNode
		s.addNode(newNode)
		s.count.Add(1)

		// 检查容量，如果超出则淘汰末尾节点 (LRU)，只有在 capacity > 0 时才进行淘汰
		if s.capacity > 0 && s.count.Load() > int64(s.capacity) {
			s.removeLRU()
		}
	}
}

// GetString获取一个字符串类型的缓存项）
func (c *CacheNode) GetString(key string) string {
	value, exists := c.Get(key)

	return gzutil.Ternary(!exists, "", cast.ToString(value))
}

// 获取一个整数类型的缓存项
func (c *CacheNode) GetInt64(key string) int64 {
	value, exists := c.Get(key)

	return gzutil.Ternary(!exists, 0, cast.ToInt64(value))
}

// Get 获取一个缓存项，自动清理过期项
// 注：本实现采用 "访问即续命（sliding expiration）" 行为：每次访问如果该条目有 ttl (>0)，会延长其过期时间。
func (c *CacheNode) Get(key string) (any, bool) {
	s := c.getShard(key)
	s.mu.Lock()
	defer s.mu.Unlock()

	node, exists := s.items[key]
	if !exists {
		return nil, false
	}

	// 判断是否过期
	if !node.expiresAt.IsZero() && time.Now().After(node.expiresAt) {
		s.removeNode(node)
		delete(s.items, key)
		s.count.Add(-1)
		return nil, false
	}

	if node.ttl > 0 {
		node.expiresAt = time.Now().Add(node.ttl)
	}
	s.moveToHead(node)
	return node.value, true
}

// Delete 删除一个缓存项（会触发 onEvict）
func (c *CacheNode) Delete(key string) {
	s := c.getShard(key)
	s.mu.Lock()
	defer s.mu.Unlock()

	if node, exists := s.items[key]; exists {
		s.removeNode(node)
		delete(s.items, key)
		s.count.Add(-1)
	}
}

// Len 返回缓存中当前的项数 (所有分片总和)
func (c *CacheNode) Len() int {
	var total int64
	for _, s := range c.shards {
		total += s.count.Load()
	}
	if total < 0 {
		return 0
	}
	return int(total)
}

// SetOnEvict 设置删除/淘汰时的回调函数
func (c *CacheNode) SetOnEvict(cb func(string, any)) {
	c.onEvict = cb
}

// Keys 返回当前所有 key 的列表（注意会锁全部 shard）
func (c *CacheNode) Keys() []string {
	var keys []string
	for _, s := range c.shards {
		s.mu.Lock()
		for k := range s.items {
			keys = append(keys, k)
		}
		s.mu.Unlock()
	}

	return keys
}

// Purge 清空整个缓存（触发 onEvict 回调以通知外部每个被清理的项）
func (c *CacheNode) Purge() {
	for _, s := range c.shards {
		s.mu.Lock()
		if c.onEvict != nil {
			for _, node := range s.items {
				c.onEvict(node.key, node.value)
			}
		}
		s.items = make(map[string]*listNode)
		s.head = nil
		s.tail = nil
		s.count.Store(0)
		s.mu.Unlock()
	}
}

// Close 停止定期清理协程（如果有）
func (c *CacheNode) Close() {
	if c.cleanerRunning.CompareAndSwap(true, false) {
		if c.cleanerStop != nil {
			close(c.cleanerStop)
		}
	}
}

// --- shard 内部 LRU 链表操作 ---
// addNode 将新节点添加到链表头部（内部操作，不触发 onEvict）
func (s *shard) addNode(node *listNode) {
	node.prev = nil
	node.next = s.head
	if s.head != nil {
		s.head.prev = node
	}
	s.head = node
	if s.tail == nil {
		s.tail = node
	}
}

// unlink 从链表中解绑一个节点（不触发 onEvict）
func (s *shard) unlink(node *listNode) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		s.head = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	} else {
		s.tail = node.prev
	}
	node.prev, node.next = nil, nil
}

// removeNode 从链表中移除一个节点并触发 onEvict（如果设置了回调）
func (s *shard) removeNode(node *listNode) {
	s.unlink(node)

	if s.parent.onEvict != nil {
		s.parent.onEvict(node.key, node.value)
	}
}

// moveToHead 将节点移动到链表头部，表示最近使用
// 使用 unlink 而不是 removeNode，以避免触发 onEvict
func (s *shard) moveToHead(node *listNode) {
	if node == s.head { // 已经是头部，无需移动
		return
	}
	s.unlink(node)
	s.addNode(node)
}

// removeLRU 移除链表末尾的节点，表示最近最少使用（并触发 onEvict）
func (s *shard) removeLRU() {
	if s.tail == nil {
		return
	}
	node := s.tail
	s.removeNode(node)
	delete(s.items, node.key)
	s.count.Add(-1)
}

// 启动后台协程定期清理所有过期项
func (c *CacheNode) startCleaner() {
	if !c.cleanerRunning.CompareAndSwap(false, true) {
		return
	}

	c.cleanerStop = make(chan struct{})

	gzutil.SafeGo(func() {
		ticker := time.NewTicker(c.cleanInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.cleanExpired()
			case <-c.cleanerStop:
				c.cleanerRunning.Store(false)
				return
			}
		}
	})
}

// 清理所有分片中过期的键
func (c *CacheNode) cleanExpired() {
	now := time.Now()
	for _, s := range c.shards {
		s.mu.Lock()
		for key, node := range s.items {
			if !node.expiresAt.IsZero() && now.After(node.expiresAt) {
				s.removeNode(node)
				delete(s.items, key)
				s.count.Add(-1)
			}
		}
		s.mu.Unlock()
	}
}
