package longgin

import (
	"github.com/armylong/go-library/service/longin/utils"
	"github.com/gin-gonic/gin"
)

func Bind(ginCtx *gin.Context, data any) (err error) {
	return utils.Bind(ginCtx, data)
}
