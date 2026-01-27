package gzutil

import (
	"math/rand/v2"
	"strings"
	"unsafe"

	"github.com/google/uuid"
)

// 生成 uuid
func GenerateUuid() string {
	v7, _ := uuid.NewV7()
	return v7.String()
}

// 生成不带横杠的32位uuid
func GenerateNoWhippletreeUuid() string {
	uuidStr := GenerateUuid()
	uuidStr = strings.ReplaceAll(uuidStr, "-", "")

	return uuidStr
}

const (
	letters      = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdBits = 6
	letterIdMask = 1<<letterIdBits - 1 // 0b111111
	letterIdMax  = 63 / letterIdBits
)

func RandString(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, rand.Int64(), letterIdMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int64(), letterIdMax
		}

		if idx := int(cache & letterIdMask); idx < len(letters) {
			b[i] = letters[idx]
			i--
		}

		cache >>= letterIdBits
		remain--
	}

	return unsafe.String(&b[0], len(b))
}

func Interval64(min, max int64) int64 {
	if min == max {
		return min
	}

	if min < 0 {
		min = 0
	}

	if min > max {
		min, max = max, min
	}

	return rand.Int64N(max-min) + min
}
