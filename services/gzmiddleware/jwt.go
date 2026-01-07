package gzmiddleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/soryetong/gooze-starter/gooze"
	"github.com/soryetong/gooze-starter/pkg/gzerror"
	"github.com/soryetong/gooze-starter/services/gzauth"
)

func Jwt() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader("Authorization")
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			gooze.Fail(ctx, gzerror.NeedLogin)
			ctx.Abort()
			return
		}

		claims, err := gzauth.ParseJwtToken(tokenString[7:])
		if err != nil {
			gooze.Fail(ctx, gzerror.NeedLogin)
			ctx.Abort()
			return
		}

		ctx.Set("claims", claims)
		ctx.Next()
	}
}
