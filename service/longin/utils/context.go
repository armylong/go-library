package utils

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
)

func GetGinContext(ctx context.Context) (*gin.Context, error) {
	ginCtx, ok := ctx.(*gin.Context)
	if !ok {
		tmp := ctx.Value("gin.Context")
		ginCtx, ok = tmp.(*gin.Context)
		if !ok {
			err := errors.New("ctx is not gin.Context")
			return nil, err
		}
	}
	return ginCtx, nil
}
