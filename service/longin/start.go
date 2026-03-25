package longgin

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	library "github.com/armylong/go-library"
	"github.com/armylong/go-library/service/application"
	"github.com/armylong/go-library/service/conf"
	"github.com/armylong/go-library/service/longin/middlewares"
	"github.com/gin-gonic/gin"
)

func Start(handler func(engine *gin.Engine)) error {

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.ContextWithFallback = true

	// 注册拦截器
	safeQuit, safeQuitWaitGroup := middlewares.SafeQuit()
	engine.Use(safeQuit)

	//业务注册路由
	handler(engine)

	//监听端口
	ip := `0.0.0.0`
	port := conf.GetHttpPort()
	listener, err := net.Listen("tcp", fmt.Sprintf(`%s:%d`, ip, port))
	if err != nil {
		return err
	}
	//goland:noinspection HttpUrlsUsage
	parse, _ := url.Parse(`http://` + listener.Addr().String())
	port, _ = strconv.Atoi(parse.Port())
	fmt.Printf(`listen: %s://%s:%d (gin: %s, longgin: %s, library: %s)`, parse.Scheme, ip, port, gin.Version, library.Version())

	if err != nil {
		return err
	}
	err = engine.RunListener(listener)
	if err != nil {
		return err
	}

	application.OnExit(func() {
		//等待安全退出
		safeQuitWaitGroup.Wait()
		//延迟关闭端口
		time.Sleep(time.Second)
		_ = listener.Close()
	})
	application.WaitExit()
	return err
	//return nil
}
