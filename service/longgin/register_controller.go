package longgin

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

// RegisterController 将一个controller中的所有actionXXX(*gin.Context)注册到路由中。
// XXX的首字母会被转为小写，如func actionHello(*gin.Context)会被注册为/hello
// 需要注意前端访问时url会区分大小写
func RegisterController(router gin.IRouter, controller any) {

	basePath := ``
	if g, ok := router.(*gin.RouterGroup); ok {
		basePath = g.BasePath()
	} else if g, ok := router.(*gin.Engine); ok {
		basePath = g.BasePath()
	}

	rv := reflect.ValueOf(controller)
	if rv.Kind() != reflect.Ptr {
		fmt.Printf(`controller 必须为指针`)
		return
	} else if rv.Elem().Kind() != reflect.Struct {
		fmt.Printf(`controller 必须是结构体的指针`)
		return
	}

	rt := rv.Type()
	i := 0
	nm := rt.NumMethod()
	for ; i < nm; i++ {
		name := rt.Method(i).Name
		if !strings.HasPrefix(name, `Action`) {
			continue
		}
		path := name[6:]
		if path == `` {
			continue
		}
		rm := rv.Method(i)
		if path[0] >= 'A' && path[0] <= 'Z' {
			path = fmt.Sprintf(`%c%s`, path[0]+(byte('a')-byte('A')), path[1:])
		}
		fullPath := basePath + `/` + path
		if rm.Type().NumIn() != 1 {
			fmt.Printf(`%s 入参个数不为1(%d)`, fullPath, rm.Type().NumIn())
			continue
		}
		if rm.Type().NumOut() != 0 {
			fmt.Printf(`%s 出参个数不为0`, fullPath)
			continue
		}
		actionHandler, ok := rm.Interface().(func(*gin.Context))
		if !ok {
			fmt.Printf(`%s 必须为 %v`, fullPath, reflect.TypeOf(func(*gin.Context) {}))
			continue
		}
		router.Any(path, actionHandler)
		fmt.Printf(`注册路由：%s -> %s.%s()`, basePath+`/`+path, rt.String(), name)
		if path == `index` {
			router.Any(`/`, actionHandler)
			fmt.Printf(`注册路由：%s/ -> %s.%s()`, basePath, rt.String(), name)
		}
	}

}
