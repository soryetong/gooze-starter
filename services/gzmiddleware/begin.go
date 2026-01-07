package gzmiddleware

import (
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

func Begin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("trace_id", uuid.NewV4().String())
		ctx.Set("source", "HttpRequest")
		ctx.Next()
	}
}
