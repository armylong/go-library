package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func Bind(ginCtx *gin.Context, data any) (err error) {

	b := binding.Default(ginCtx.Request.Method, ginCtx.ContentType())
	if bb, ok := b.(binding.BindingBody); ok {
		if bb == binding.JSON {
			bb = json
		}
		return mustBindBodyWith(ginCtx, data, bb)
	} else {
		return ginCtx.MustBindWith(data, b)
	}
}

func mustBindBodyWith(ginCtx *gin.Context, data any, bb binding.BindingBody) (err error) {
	if err = ginCtx.ShouldBindBodyWith(data, bb); err != nil {
		ginCtx.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind)
		return err
	}
	return nil
}
