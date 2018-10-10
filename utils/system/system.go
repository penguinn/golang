package system

import (
	"runtime"
	"strings"
	"strconv"
	"fmt"
	"reflect"
	"time"
	"path"
	"io"
)

//返回goroutineId
func GetGoId() int64 {
	var (
		buf [64]byte
		n = runtime.Stack(buf[:], false)
		stk = strings.TrimSpace(string(buf[:n]))
	)

	idField := strings.Fields(stk)[1]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Errorf("can not get goroutine id: %v", err))
	}

	return int64(id)
}

// 返回方法的名称
func NameOfFunction(f interface{}) string {
	r := reflect.ValueOf(f)
	if r.Type().Kind() != reflect.Func {
		return ""
	}
	return runtime.FuncForPC(r.Pointer()).Name()
}

// 调用某个函数之前skip函数名称
func CallerFunc(skip int) string {
	p, _, _, _ := runtime.Caller(skip + 1)
	f := runtime.FuncForPC(p)
	return f.Name()
}

// 调用时间
func CallTimeUsage(w io.Writer) func() {
	pc, _, _, _ := runtime.Caller(1)
	start := time.Now()

	return func() {
		w.Write([]byte(fmt.Sprintf("Call Function[%s] Use Time %v", path.Base(runtime.FuncForPC(pc).Name()), time.Now().Sub(start))))
	}
}