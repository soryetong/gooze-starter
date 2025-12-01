package gooze

import "github.com/gin-gonic/gin"

type HandlerFunc func(*gin.Context) error

func Adapter(h HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := h(ctx); err != nil {
			_ = ctx.Error(err)
			return
		}
	}
}
