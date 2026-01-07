package gzmiddleware

import (
	"github.com/gin-gonic/gin"
	"github.com/soryetong/gooze-starter/gooze"
	"github.com/soryetong/gooze-starter/pkg/gzerror"
	"github.com/soryetong/gooze-starter/services/gzlimiter"
)

func Limiter(limiterStore *gzlimiter.LimiterStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !limiterStore.Allow(ctx) {
			gooze.Fail(ctx, gzerror.RequestLimit)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
