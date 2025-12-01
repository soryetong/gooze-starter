package gzutil

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os/exec"
	"reflect"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/spf13/cast"
)

var (
	ServerAddr  string
	ServerIsTLS bool
)

func GetServerAddr() string {
	prefix := "http"
	if ServerIsTLS {
		prefix = "https"
	}

	addr := ServerAddr
	addrArr := strings.Split(ServerAddr, ":")
	if addrArr[0] == "" {
		addrArr[0] = GetLocalIP()
		addr = strings.Join(addrArr, ":")
	}

	return fmt.Sprintf("%s://%s", prefix, addr)
}

func SafeGo(fn func()) {
	go RunSafe(fn)
}

func RunSafe(fn func()) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("[run func] panic: %v, funcName: %s，stack=%s \n",
				err, GetCallerName(fn), string(debug.Stack()))
		}
	}()

	fn()
}

// ValidatePasswd 校验密码是否一致
func ValidatePasswd(pwd, salt, passwd string) bool {
	return Md5Encode(pwd+salt) == passwd
}

// MakePasswd 生成密码
func MakePasswd(pwd, salt string) string {
	return Md5Encode(pwd + salt)
}

// Md5Encode md5处理
func Md5Encode(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

type Number interface {
	~int | ~int32 | ~int64 | ~float64 | ~float32 | ~string
}

// IsValidNumber 判断是否是有效数字
func IsValidNumber[T Number](value T) bool {
	switch v := any(value).(type) {
	case int:
		return v > 0
	case int32:
		return v > 0
	case int64:
		return v > 0
	case float64:
		return v > 0
	case float32:
		return v > 0
	case string:
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			return num > 0
		}
	}
	return false
}

// GetCallerName 获取调用者的名称
func GetCallerName(caller interface{}) string {
	typ := reflect.TypeOf(caller)

	switch typ.Kind() {
	case reflect.Ptr: // 指针类型
		if typ.Elem().Kind() == reflect.Struct {
			return typ.Elem().Name()
		}
		return fmt.Sprintf("%v", reflect.ValueOf(caller).Elem())
	case reflect.Struct: // 结构体类型
		return typ.Name()
	default: // 其他类型
		return fmt.Sprintf("%v", caller)
	}
}

// GetModuleName retrieves the current project's module name using `go list`.
func GetModuleName() (string, error) {
	cmd := exec.Command("go", "list", "-m")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get module name: %w", err)
	}
	return strings.TrimSpace(out.String()), nil
}

// SeparateCamel 按照自定符号分隔驼峰
func SeparateCamel(name, separator string) string {
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteString(separator)
		}

		result.WriteRune(r | ' ')
	}
	return result.String()
}

// UcFirst 字符串首字母大写
func UcFirst(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// LcFirst 字符串首字母小写
func LcFirst(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// GetRequestPath 获取请求路径
func GetRequestPath(path, prefix string) (uri string, id int64) {
	uri = strings.TrimPrefix(path, prefix)
	re := regexp.MustCompile(`^(.*)/(\d+)$`)
	matches := re.FindStringSubmatch(uri)
	if len(matches) == 3 {
		uri = matches[1]
		id = cast.ToInt64(matches[2])
	}

	return
}

// ConvertToRestfulURL 将URI转换为REST ful URL
func ConvertToRestfulURL(url string) string {
	re := regexp.MustCompile(`(^.+?/[^/]+)/\d+$`)
	return re.ReplaceAllString(url, `$1/:id`)
}

// ConvertRestfulURLToUri 将REST ful URL转换为URI
func ConvertRestfulURLToUri(url string) (string, string) {
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return url, ""
	}
	last := parts[len(parts)-1]
	if strings.HasPrefix(last, ":") {
		path := strings.Join(parts[:len(parts)-1], "/")
		param := strings.TrimPrefix(last, ":")
		return path, param
	}

	return url, ""
}

// RemoveDomain 移除URL中的域名
func RemoveDomain(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	result := parsedURL.EscapedPath()
	if parsedURL.RawQuery != "" {
		result += "?" + parsedURL.RawQuery
	}

	if parsedURL.Fragment != "" {
		result += "#" + parsedURL.Fragment
	}

	return result
}

// JoinDomain 将域名和 URL 路径拼接为完整 URL
func JoinDomain(domain string, path string) string {
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
		domain = "http://" + domain
	}

	parsedDomain, err := url.Parse(domain)
	if err != nil {
		return ""
	}

	parsedPath, err := url.Parse(path)
	if err != nil {
		return ""
	}

	return parsedDomain.ResolveReference(parsedPath).String()
}

// GetMapValue 获取map的值
func GetMapValue[T any](m map[string]interface{}, key string) T {
	var zero T

	value, exists := m[key]
	if exists {
		v := reflect.ValueOf(value)
		if v.Type().ConvertibleTo(reflect.TypeOf(zero)) {
			return v.Convert(reflect.TypeOf(zero)).Interface().(T)
		}
	}

	return zero
}

type MapSupportedTypes interface {
	string | int64 | float64 | bool
}

// GetMapSpecificValue 获取map的特定类型的值, 相较于GetMapValue, 不用每次反射获取值
func GetMapSpecificValue[T MapSupportedTypes](m map[string]interface{}, key string) T {
	var zero T

	value, exists := m[key]
	if exists {
		if v, ok := value.(T); ok {
			return v
		}
	}

	var result any
	switch v := value.(type) {
	case float64:
		if _, ok := any(zero).(int64); ok {
			result = int64(v)
		} else if _, ok := any(zero).(bool); ok {
			result = v != 0
		} else {
			return zero
		}
	case int64:
		if _, ok := any(zero).(float64); ok {
			result = float64(v)
		} else if _, ok := any(zero).(bool); ok {
			result = v != 0
		} else {
			return zero
		}
	case string:
		if _, ok := any(zero).(bool); ok {
			lowerVal := strings.ToLower(v)
			if lowerVal == "true" || lowerVal == "1" {
				result = true
			} else if lowerVal == "false" || lowerVal == "0" {
				result = false
			} else {
				return zero
			}
		} else {
			result = v
		}
	case bool:
		if _, ok := any(zero).(string); ok {
			result = fmt.Sprintf("%v", v) // 转成 "true" / "false"
		} else {
			result = v
		}
	default:
		return zero
	}

	if finalValue, ok := result.(T); ok {
		return finalValue
	}

	return zero
}

func InArray[T comparable](val T, slice []T) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// 获取操作系统
func GetPlatform(userAgent string) string {
	ua := strings.ToLower(userAgent)

	// 移动端
	if strings.Contains(ua, "android") {
		return "Android"
	} else if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") || strings.Contains(ua, "ipod") {
		return "iOS"
	}

	// 桌面端
	if strings.Contains(ua, "windows") {
		return "Windows"
	} else if strings.Contains(ua, "macintosh") || strings.Contains(ua, "mac os") {
		return "MacOS"
	} else if strings.Contains(ua, "linux") {
		return "Linux"
	}

	return "Unknown"
}

// 获取浏览器类型
func GetBrowser(userAgent string) string {
	ua := strings.ToLower(userAgent)

	if strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg") {
		return "Google Chrome"
	} else if strings.Contains(ua, "edg") {
		return "Microsoft Edge"
	} else if strings.Contains(ua, "firefox") {
		return "Mozilla Firefox"
	} else if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		return "Apple Safari"
	} else if strings.Contains(ua, "opr") || strings.Contains(ua, "opera") {
		return "Opera"
	} else if strings.Contains(ua, "msie") || strings.Contains(ua, "trident") {
		return "Internet Explorer"
	}

	return "Unknown"
}

// 三元运算符
func Ternary[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

// 合并切片
func MergeSlice[T any](a, b []T) []T {
	la, lb := len(a), len(b)
	if lb == 0 {
		return a
	}
	if la == 0 {
		return b
	}
	out := make([]T, la+lb)
	copy(out, a)
	copy(out[la:], b)
	return out
}

// 移除切片中的元素, 第一个参数基值, 第二个参数待移除的元素
func RemoveSliceElements[T comparable](a, b []T) []T {
	toRemove := make(map[T]struct{}, len(b))
	for _, v := range b {
		toRemove[v] = struct{}{}
	}

	out := make([]T, 0, len(a))
	for _, v := range a {
		if _, found := toRemove[v]; !found {
			out = append(out, v)
		}
	}
	return out
}

// 去重
func Unique[T comparable](items []T) []T {
	seen := make(map[T]struct{})
	result := make([]T, 0, len(items))

	for _, item := range items {
		if _, ok := seen[item]; !ok {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

func RandOne[T any](list []T) T {
	return list[rand.Intn(len(list))]
}
