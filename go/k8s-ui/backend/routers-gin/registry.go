package routers_gin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"reflect"
)

type Registry struct {
	engine *gin.Engine
}

func New(e *gin.Engine) *Registry {
	return &Registry{
		engine: e,
	}
}

type ControllerInterface interface {
	Init()
}

func callReflect(any interface{}, name string, args ...interface{}) []reflect.Value {
	inputs := make([]reflect.Value, len(args))
	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}

	if v := reflect.ValueOf(any).MethodByName(name); v.String() == "<invalid Value>" {
		return nil
	} else {
		return v.Call(inputs)
	}
}

func (r *Registry) AddRouter(method string, path string, controller ControllerInterface, action string) {
	// init
	callReflect(controller, "Init")
	fun := callReflect(controller, action)
	if fun == nil || len(fun) == 0 {
		panic(fmt.Sprintf("%T.%s() does not exists!  ", controller, action))
	}

	handlerFunc := fun[0].Interface().(gin.HandlerFunc)
	// add path
	ret := callReflect(r.engine, method, path, handlerFunc)
	if ret == nil {
		panic(fmt.Sprintf("%T.%s() does not exists!  ", r.engine, method))
	}
}
