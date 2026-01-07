package gzauth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/soryetong/gooze-starter/gzconsole"
	"github.com/soryetong/gooze-starter/pkg/gzutil"
	"github.com/spf13/viper"
)

var secretKey []byte

func getSecretKey() []byte {
	if len(secretKey) == 0 {
		secretKey = loadSecretKey()
	}

	return secretKey
}

func loadSecretKey() []byte {
	key := viper.GetString("Jwt.SecretKey")
	if key != "" {
		return []byte(key)
	}

	randomBytes := []byte("1234567890")
	gzconsole.Echo.Warnf("⚠️ 警告: Jwt.SecretKey 为空，使用固定密钥: %s\n", string(randomBytes))

	return randomBytes
}

func GenerateJwtToken(claimsMap jwt.MapClaims) (string, error) {
	if claimsMap == nil {
		claimsMap = make(jwt.MapClaims)
	}
	claimsMap["iat"] = time.Now().Unix()
	claimsMap["nbf"] = time.Now().Unix()
	if _, ok := claimsMap["exp"]; !ok {
		expire := viper.GetInt64("Jwt.Expire")
		if expire > 0 {
			claimsMap["exp"] = time.Now().Add(time.Duration(expire) * time.Second).Unix()
		} else {
			gzconsole.Echo.Warnf("⚠️ 警告: Jwt.Expire 为空，使用默认值: 20 分钟\n")
			claimsMap["exp"] = time.Now().Add(time.Minute * 20).Unix()
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsMap)
	tokenString, err := token.SignedString(getSecretKey())
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseJwtToken(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return getSecretKey(), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("无效的 Token")
}

func GetTokenValue[T gzutil.MapSupportedTypes](ctx context.Context, key string) T {
	claimsMap, ok := ctx.Value("claims").(map[string]interface{})
	if !ok {
		var zero T
		return zero
	}

	return gzutil.GetMapSpecificValue[T](claimsMap, key)
}
