package longgin

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

// RegisterJsonController 将一个按规范书写的结构体注册到路由中
func RegisterJsonController(router gin.IRouter, controller any) {
	basePath := ``
	if g, ok := router.(*gin.RouterGroup); ok {
		basePath = g.BasePath()
	} else if g, ok := router.(*gin.Engine); ok {
		basePath = g.BasePath()
	}

	rv := reflect.ValueOf(controller)
	if rv.Kind() != reflect.Ptr {
		log.Panic(nil, basePath, `controller 必须为指针`)
	} else if rv.Elem().Kind() != reflect.Struct {
		fmt.Print(basePath, "controller 必须是结构体的指针")
	}

	rt := rv.Type()
	i := 0
	nm := rt.NumMethod()
	for ; i < nm; i++ {
		func() {
			methodName := rt.Method(i).Name
			relativePath := getPathByMethodName(methodName)
			if relativePath == "" {
				return
			}
			rm := rv.Method(i)
			fullPath := basePath + `/` + relativePath
			defer func() {
				err := recover()
				if err != nil {
					fmt.Println(`注册路由:失败:`, fullPath, rt.String(), methodName, err)
				}
			}()
			actionHandlers := make([]gin.HandlerFunc, 0, 1)
			actionHandler := NewJsonActionHandler(rm)
			actionHandlers = append(actionHandlers, actionHandler)
			router.Any(relativePath, actionHandlers...)
			fmt.Printf("注册路由: %s -> %s.%s()\n", fullPath, rt.String(), methodName)
			if relativePath == `index` {
				router.Any(`/`, actionHandlers...)
				fmt.Printf("注册路由: %s/ -> %s.%s()\n", basePath, rt.String(), methodName)
			}
		}()
	}
}

// NewJsonActionHandler 创建一个actionHandler包装
func NewJsonActionHandler(handler any) gin.HandlerFunc {
	rm, ok := handler.(reflect.Value)
	if !ok {
		rm = reflect.ValueOf(handler)
	}
	if rm.Kind() != reflect.Func {
		panic(errors.New(`ActionHandler必须为Func`))
	}
	rmt := rm.Type()
	ctxIndex, bindIndex, bindType, err := readInIndexes(rmt)
	if err != nil {
		panic(err)
	}
	errIndex, retIndex, err := readOutIndexes(rmt)
	if err != nil {
		panic(err)
	}
	argsCreator := newArgsCreator(bindIndex, ctxIndex, bindType)
	actionHandler := newActionHandler(rm, argsCreator, errIndex, retIndex)
	return actionHandler
}

// getPathByMethodName 根据方法名获取相对路径
func getPathByMethodName(name string) string {
	prefix := `Action`
	prefixLen := len(prefix)
	if len(name) <= prefixLen {
		return ``
	}
	if !strings.HasPrefix(name, prefix) {
		return ``
	}
	path := name[prefixLen:]
	if path[0] >= 'A' && path[0] <= 'Z' {
		path = fmt.Sprintf(`%c%s`, path[0]+(byte('a')-byte('A')), path[1:])
	}
	return path
}

// readInIndexes 读取ctx和form在参数中的位置
func readInIndexes(rmt reflect.Type) (ctxIndex int, bindIndex int, bindType reflect.Type, err error) {
	numIn := rmt.NumIn()
	if numIn > 2 {
		return -1, -1, nil, errors.New(`入参个数不能大于2个`)
	}
	ctxIndex = -1
	bindIndex = -1
	ginCtxRt := reflect.TypeOf((*gin.Context)(nil))
	stdCtxRt := reflect.TypeOf((*context.Context)(nil)).Elem()
	for j := 0; j < numIn; j++ {
		in := rmt.In(j)
		if in == ginCtxRt || in == stdCtxRt {
			ctxIndex = j
		} else {

			inKind := in.Kind()
			if inKind == reflect.Ptr {
				inKind = in.Elem().Kind()
			}
			if inKind != reflect.Struct {
				return -1, -1, nil, errors.New(`bind参数必须为struct或*struct`)
			}
			bindType = in
			bindIndex = j
		}
	}
	if numIn > 1 && (ctxIndex == -1 || bindIndex == -1) {
		return -1, -1, nil, errors.New(`入参为2个时，必须一个为*gin.Context另一个非*gin.Context`)
	}
	return ctxIndex, bindIndex, bindType, nil
}

// readOutIndexes 读取resp和err在返回值中的位置
func readOutIndexes(rmt reflect.Type) (errIndex, retIndex int, err error) {
	errIndex = -1
	retIndex = -1
	numOut := rmt.NumOut()
	if numOut > 2 {

		return -1, -1, errors.New(`出参个数不能大于2个`)
	}
	if numOut > 0 {
		errorInterface := reflect.TypeOf((*error)(nil)).Elem()
		for j := 0; j < numOut; j++ {
			func() {
				out := rmt.Out(j)
				switch out.Kind() {
				case reflect.Struct:
					fallthrough
				case reflect.Interface:
					fallthrough
				case reflect.Ptr:
					if out.Implements(errorInterface) {
						errIndex = j
						return
					}
				}
				retIndex = j
			}()
		}
	}
	if numOut > 1 && (errIndex == -1 || retIndex == -1) {
		return -1, -1, errors.New(`出参为2个时，必须一个为error另一个非error`)
	}
	return errIndex, retIndex, nil
}

// newActionHandler 创建一个action处理器
func newActionHandler(rm reflect.Value, createArgs actionArgsCreator, errIndex int, retIndex int) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		inArgs := createArgs(ctx)
		if len(ctx.Errors) != 0 {
			ctx.JSON(http.StatusInternalServerError, ErrorWithContext(ctx, ctx.Errors.Last(), ErrorUnknown))
			return
		}
		outValues := rm.Call(inArgs)
		if ctx.IsAborted() {
			return
		}
		if errIndex >= 0 {
			refErr := outValues[errIndex]
			if !refErr.IsNil() {
				err := refErr.Interface().(error)
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, ErrorWithContext(ctx, err, ErrorUnknown))
					return
				}
			}
		}
		if retIndex >= 0 {
			refVal := outValues[retIndex]
			if refVal.Kind() != reflect.Ptr || !refVal.IsNil() {
				ret := refVal.Interface()
				if _, ok := ret.(RawResponse); ok {
					ctx.JSON(http.StatusOK, ret)
					return
				}
				ctx.JSON(http.StatusOK, Success(ret))
				return
			}
		}
		ctx.JSON(http.StatusOK, Success(nil))
	}
}

type actionArgsCreator func(ctx *gin.Context) []reflect.Value

// newArgsCreator 创建一个入参生成器
func newArgsCreator(bindIndex int, ctxIndex int, bindType reflect.Type) actionArgsCreator {
	inNum := 0
	if bindIndex >= 0 {
		inNum++
	}
	if ctxIndex >= 0 {
		inNum++
	}
	bindIsPtr := bindType != nil && bindType.Kind() == reflect.Ptr
	createArgs := func(ctx *gin.Context) []reflect.Value {
		args := make([]reflect.Value, inNum)
		if inNum == 0 {
			return args
		}
		ctx.Set("gin.Context", ctx)
		if ctxIndex != -1 {
			args[ctxIndex] = reflect.ValueOf(ctx)
		}
		if bindIndex != -1 {
			if bindIsPtr {
				bindValue := reflect.New(bindType.Elem())
				v := bindValue.Interface()
				err := Bind(ctx, v)
				if err != nil {
					fmt.Print("前端参数解析异常", err.Error())
				}
				args[bindIndex] = bindValue
			} else {
				bindValue := reflect.New(bindType)
				v := bindValue.Interface()
				err := Bind(ctx, v)
				if err != nil {
					fmt.Print("前端参数解析异常", err.Error())
				}
				args[bindIndex] = bindValue.Elem()
			}
		}
		return args
	}
	return createArgs
}
