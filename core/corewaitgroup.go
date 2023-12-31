package core

import (
	"reflect"
	"sync"

	"github.com/shinanca/gonec/names"
)

// VMWaitGroup - группа ожидания исполнения горутин
type VMWaitGroup struct {
	wg sync.WaitGroup
}

var ReflectVMWaitGroup = reflect.TypeOf(VMWaitGroup{})

func (x *VMWaitGroup) VMTypeString() string { return "ГруппаОжидания" }

func (x *VMWaitGroup) Interface() interface{} {
	return x
}

func (x *VMWaitGroup) String() string {
	return "Группа ожидания"
}

func (x *VMWaitGroup) Add(delta int) {
	x.wg.Add(delta)
}

func (x *VMWaitGroup) Done() {
	x.wg.Done()
}

func (x *VMWaitGroup) Wait() {
	x.wg.Wait()
}

func (x *VMWaitGroup) MethodMember(name int) (VMFunc, bool) {
	// только эти методы будут доступны из кода на языке Гонец!
	switch names.UniqueNames.GetLowerCase(name) {
	case "добавить":
		return VMFuncOneParam(x.Добавить), true
	case "завершить":
		return VMFuncZeroParams(x.Завершить), true
	case "ожидать":
		return VMFuncZeroParams(x.Ожидать), true
	}
	return nil, false
}

func (x *VMWaitGroup) Добавить(n VMInt, rets *VMSlice) error {
	x.Add(int(n))
	return nil
}

func (x *VMWaitGroup) Завершить(rets *VMSlice) error {
	x.Done()
	return nil
}

func (x *VMWaitGroup) Ожидать(rets *VMSlice) error {
	x.Wait()
	return nil
}
