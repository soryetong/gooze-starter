package gzmiddleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/soryetong/gooze-starter/gooze"
	"github.com/soryetong/gooze-starter/gzconsole"
	"github.com/soryetong/gooze-starter/pkg/gzerror"
	"github.com/soryetong/gooze-starter/pkg/gzutil"
	"github.com/soryetong/gooze-starter/services/gzauth"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

func Casbin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if strings.ToLower(ctx.Request.Method) == "OPTIONS" {
			ctx.Next()
			return
		}

		roleId := gzauth.GetTokenValue[int64](ctx, "role_id")
		if roleId == 0 {
			gzconsole.Echo.Info("ℹ️ 提示: 无法使用 `Casbin` 权限校验, 请确保 `Token` 中包含了字段 `role_id`")
			ctx.Next()
			return
		}

		path := gzutil.ConvertToRestfulURL(strings.TrimPrefix(ctx.Request.URL.Path, viper.GetString("App.RouterPrefix")))
		success, _ := gooze.Casbin.Enforce(cast.ToString(roleId), path, ctx.Request.Method)
		if !success {
			gooze.Fail(ctx, gzerror.NoAuth)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
