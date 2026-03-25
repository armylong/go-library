package middlewares

import (
	"github.com/gin-gonic/gin"
	"sync"
)

func SafeQuit() (safeQuit gin.HandlerFunc, wg *sync.WaitGroup) {
	wg = &sync.WaitGroup{}
	return func(context *gin.Context) {
		wg.Add(1)
		defer wg.Done()
		context.Next()
	}, wg
}
