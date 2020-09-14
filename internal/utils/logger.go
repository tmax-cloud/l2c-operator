package utils

import (
	"github.com/go-logr/logr"
	"reflect"
	"runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)

type TupLogger struct {
	subresource           interface{}
	resNamespace, resName string
}

func funcName() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(4, pc) //Skip: 3 (Callers, getFuncName, GetRegistryLogger, get)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	return frame.Function
}

func GetTupLogger(subresource interface{}, resNamespace, resName string) logr.Logger {
	typeName := reflect.TypeOf(subresource).Name()
	funcName := funcName()
	path := strings.Split(funcName, ".")
	funcName = path[len(path)-1]

	return log.Log.WithValues(typeName+".Namespace", resNamespace, typeName+".Name", resName, typeName+".Api", funcName)
}

func NewTupLogger(subresource interface{}, resNamespace, resName string) *TupLogger {
	logger := &TupLogger{}
	logger.subresource = subresource
	logger.resNamespace = resNamespace
	logger.resName = resName

	return logger
}

func (r *TupLogger) Info(msg string, keysAndValues ...interface{}) {
	log := GetTupLogger(r.subresource, r.resNamespace, r.resName)
	if len(keysAndValues) > 0 {
		log.Info(msg, keysAndValues...)
	} else {
		log.Info(msg)
	}
}

func (r *TupLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	log := GetTupLogger(r.subresource, r.resNamespace, r.resName)
	if len(keysAndValues) > 0 {
		log.Error(err, msg, keysAndValues...)
	} else {
		log.Error(err, msg)
	}
}
