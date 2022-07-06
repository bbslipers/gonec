package core

import (
	"fmt"
	"reflect"
)

// VMFunc вызывается как обертка метода объекта метаданных или обертка функции библиотеки
// возвращаемое из обертки значение должно быть приведено к типу вирт. машины
// функции такого типа создаются на языке Гонец,
// их можно использовать в стандартной библиотеке, проверив на этот тип
// в args передаются входные параметры, в rets передается ссылка на слайс возвращаемых значений - он заполняется в функции
type VMFunc func(args VMSlice, rets *VMSlice) error

var ReflectVMFunc = reflect.TypeOf(VMFunc(nil))

func (f VMFunc) VMTypeString() string { return "Функция" }

func (f VMFunc) Interface() interface{} {
	return f
}

func (f VMFunc) String() string {
	return fmt.Sprintf("[Функция: %p]", f)
}

func (f VMFunc) Func() VMFunc {
	return f
}

type (
	VMMethod      = VMFunc
	VMConstructor = func(VMSlice) error
)

func VMFuncMustParamType[V VMValue](args VMSlice, i, n int) error {
	if _, ok := args[i].(V); !ok {
		return VMErrorNeedArgType[V](i, n)
	}
	return nil
}

func paramCheckHelper[V VMValue](args VMSlice, i, n int, errs []error) error {
	if _, ok := args[i].(V); !ok {
		return errs[i]
	}
	return nil
}

func VMFuncZeroParams(f VMMethod) VMFunc {
	return VMFunc(func(args VMSlice, rets *VMSlice) error {
		if len(args) != 0 {
			return VMErrorNoNeedArgs
		}
		return f(args, rets)
	})
}

func VMFuncOneParam[V1 VMValue](f VMMethod) VMFunc {
	errs := []error{VMErrorNeedArgType[V1](0, 1)}
	return VMFuncNParams(1, func(args VMSlice, rets *VMSlice) error {
		if err := paramCheckHelper[V1](args, 0, 1, errs); err != nil {
			return err
		}
		return f(args, rets)
	})
}

func VMFuncTwoParams[V1, V2 VMValue](f VMMethod) VMFunc {
	errs := []error{
		VMErrorNeedArgType[V1](0, 2),
		VMErrorNeedArgType[V2](1, 2),
	}
	return VMFuncNParams(2, func(args VMSlice, rets *VMSlice) error {
		if err := paramCheckHelper[V1](args, 0, 2, errs); err != nil {
			return err
		} else if err := paramCheckHelper[V2](args, 1, 2, errs); err != nil {
			return err
		}
		return f(args, rets)
	})
}

func VMFuncThreeParams[V1, V2, V3 VMValue](f VMMethod) VMFunc {
	errs := []error{
		VMErrorNeedArgType[V1](0, 3),
		VMErrorNeedArgType[V2](1, 3),
		VMErrorNeedArgType[V3](2, 3),
	}
	return VMFuncNParams(3, func(args VMSlice, rets *VMSlice) error {
		if err := paramCheckHelper[V1](args, 0, 3, errs); err != nil {
			return err
		} else if err := paramCheckHelper[V2](args, 1, 3, errs); err != nil {
			return err
		} else if err := paramCheckHelper[V3](args, 2, 3, errs); err != nil {
			return err
		}
		return f(args, rets)
	})
}

func VMFuncNParams(n int, f VMMethod) VMFunc {
	needArgsErr := VMErrorNeedArgs(n)
	return VMFunc(func(args VMSlice, rets *VMSlice) error {
		if len(args) != n {
			switch n {
			case 0:
				return VMErrorNoNeedArgs
			default:
				return needArgsErr
			}
		}
		return f(args, rets)
	})
}
