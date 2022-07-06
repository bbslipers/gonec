package core

import (
	"github.com/shinanca/gonec/names"
)

// VMChan - канал для передачи любого типа вирт. машины
type VMChan chan VMValue

func (x VMChan) VMTypeString() string { return "Канал" }

func (x VMChan) Interface() interface{} {
	return x
}

func (x VMChan) Send(v VMValue) {
	x <- v
}

func (x VMChan) Recv() (VMValue, bool) {
	rv, ok := <-x
	return rv, ok
}

func (x VMChan) TrySend(v VMValue) (ok bool) {
	select {
	case x <- v:
		ok = true
	default:
		ok = false
	}
	return
}

func (x VMChan) TryRecv() (v VMValue, ok bool, notready bool) {
	select {
	case v, ok = <-x:
		notready = false
	default:
		ok = false
		notready = true
	}
	return
}

func (x VMChan) Close() { close(x) }

func (x VMChan) Size() int { return cap(x) }

func (x VMChan) MethodMember(name int) (VMFunc, bool) {
	// только эти методы будут доступны из кода на языке Гонец!
	switch names.UniqueNames.GetLowerCase(name) {
	case "закрыть":
		return VMFuncZeroParams(x.Закрыть), true
	case "размер":
		return VMFuncZeroParams(x.Размер), true
		// TODO: подключить соединение
	}
	return nil, false
}

func (x VMChan) Закрыть(args VMSlice, rets *VMSlice) error {
	x.Close()
	return nil
}

func (x VMChan) Размер(args VMSlice, rets *VMSlice) error {
	rets.Append(VMInt(x.Size()))
	return nil
}
