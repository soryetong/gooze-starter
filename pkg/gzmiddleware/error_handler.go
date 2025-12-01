package gzmiddleware

import (
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/soryetong/gooze-starter/gooze"
	"github.com/soryetong/gooze-starter/pkg/gzerror"
	"go.uber.org/zap"
)

func ErrorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				gooze.Log.Error("捕获到异常错误",
					zap.Any("err", r),
					zap.String("stack", string(debug.Stack())),
				)

				gooze.Fail(ctx, gzerror.Error, "系统异常")
				ctx.Abort()
			}
		}()

		ctx.Next()

		if len(ctx.Errors) > 0 {
			err := ctx.Errors.Last().Err

			// 业务错误
			if ge, ok := err.(*gzerror.BizError); ok {
				gooze.Fail(ctx, ge.Code, ge.Msg)
				return
			}

			gooze.Fail(ctx, gzerror.Error, err.Error())
		}
	}
}
