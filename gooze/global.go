package gooze

import (
	"sync"

	casbinV2 "github.com/casbin/casbin/v2"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/soryetong/gooze-starter/gzconsole"
	"github.com/soryetong/gooze-starter/pkg/gzcache"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"

	_ "github.com/soryetong/gooze-starter/cmdgen"
)

var (
	dbMap sync.Map

	Config *BaseConfig
	Log    *iLog
	Cache  *gzcache.CacheNode
	Casbin *casbinV2.SyncedEnforcer
	Rdb    redis.Cmdable
	Mdb    *mongo.Client
)

func Run() {
	gzconsole.Register(-1, serviceMgrCmd)
	if err := gzconsole.RootCmd.Execute(); err != nil {
		gzconsole.Echo.Fatalf("❌  服务启动失败: [%s] \n", err)
	}
}

func RunTest(config, env string, show bool) {
	if err := initialize(config, env, show); err != nil {
		gzconsole.Echo.Fatalf("初始化失败：[%v]\n", err)
	}
	if err := gzconsole.RunStartupTasks(); err != nil {
		gzconsole.Echo.Fatalf("初始化失败: [%v] \n", err)
	}
}

type instance struct {
	Name string
	GORM *gorm.DB
	SQLX *sqlx.DB
}

func SetDb(name string, gdb *gorm.DB, sdb *sqlx.DB) {
	if gdb != nil {
		dbMap.Store(name, &instance{Name: name, GORM: gdb})
	} else if sdb != nil {
		dbMap.Store(name, &instance{Name: name, SQLX: sdb})
	}
}

func Gorm(name ...string) *gorm.DB {
	if len(name) == 0 {
		name = []string{"default"}
	}
	if v, ok := dbMap.Load(name[0]); ok {
		return v.(*instance).GORM
	}

	return nil
}

func Sqlx(name ...string) *sqlx.DB {
	if len(name) == 0 {
		name = []string{"default"}
	}
	if v, ok := dbMap.Load(name[0]); ok {
		return v.(*instance).SQLX
	}

	return nil
}
