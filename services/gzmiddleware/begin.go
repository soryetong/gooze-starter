package gzmiddleware

import (
	"github.com/gin-gonic/gin"
	"github.com/soryetong/gooze-starter/pkg/gzutil"
)

func Begin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("trace_id", gzutil.GenerateUuid())
		ctx.Set("source", "HttpRequest")
		ctx.Next()
	}
}
