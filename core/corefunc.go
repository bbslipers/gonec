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

type argValidator func(VMValue) bool

func typeValidator[V VMValue](arg VMValue) bool {
	_, ok := arg.(V)
	return ok
}

// funcValidateArgs validates params using the passed paramCheckHelper instances
// length of args must be validated prior to this call and must be equal to length of errs
func funcValidateArgs(validators []argValidator, errs []error, args VMSlice) error {
	for i, v := range validators {
		if !v(args[i]) {
			return errs[i]
		}
	}
	return nil
}

// vmFuncValidatedArgs wraps a VMFunc with validation using given validators and errors,
// as well as VMFuncNParamsOptionals, which validates the number of args.
// It is expected that len(validators) == len(errs) == nreq
func vmFuncValidatedArgs(nreq, nopt int, validators []argValidator, errs []error,
	f VMFunc,
) VMFunc {
	return VMFuncNParamsOptionals(nreq, nopt, func(args VMSlice, rets *VMSlice) error {
		if err := funcValidateArgs(validators, errs, args); err != nil {
			return err
		}
		return f(args, rets)
	})
}

func VMFuncZeroParams(f func(rets *VMSlice) error) VMFunc {
	return VMFunc(func(args VMSlice, rets *VMSlice) error {
		if len(args) != 0 {
			return VMErrorNoNeedArgs
		}
		return f(rets)
	})
}

func VMFuncOneParam[V1 VMValue](f func(arg1 V1, rets *VMSlice) error) VMFunc {
	return VMFuncOneParamOptionals(0, func(arg1 V1, rest VMSlice, rets *VMSlice) error {
		return f(arg1, rets)
	})
}

func VMFuncOneParamOptionals[V1 VMValue](n int, f func(
	arg1 V1, rest VMSlice, rets *VMSlice) error,
) VMFunc {
	errs := []error{VMErrorNeedArgType[V1](0, 1)}
	validators := []argValidator{typeValidator[V1]}
	return vmFuncValidatedArgs(1, n, validators, errs, func(args VMSlice, rets *VMSlice) error {
		return f(args[0].(V1), args[1:], rets)
	})
}

func VMFuncTwoParams[V1, V2 VMValue](f func(
	arg1 V1, arg2 V2, rets *VMSlice) error,
) VMFunc {
	return VMFuncTwoParamsOptionals(0, func(arg1 V1, arg2 V2,
		rest VMSlice, rets *VMSlice,
	) error {
		return f(arg1, arg2, rets)
	})
}

func VMFuncTwoParamsOptionals[V1, V2 VMValue](n int, f func(
	arg1 V1, arg2 V2, rest VMSlice, rets *VMSlice) error,
) VMFunc {
	errs := []error{
		VMErrorNeedArgType[V1](0, 2),
		VMErrorNeedArgType[V2](1, 2),
	}
	validators := []argValidator{typeValidator[V1], typeValidator[V2]}
	return vmFuncValidatedArgs(2, n, validators, errs, func(args VMSlice, rets *VMSlice) error {
		return f(args[0].(V1), args[1].(V2), args[2:], rets)
	})
}

func VMFuncThreeParams[V1, V2, V3 VMValue](f func(
	arg1 V1, arg2 V2, arg3 V3, rets *VMSlice) error,
) VMFunc {
	return VMFuncThreeParamsOptionals(0, func(arg1 V1, arg2 V2, arg3 V3,
		rest VMSlice, rets *VMSlice,
	) error {
		return f(arg1, arg2, arg3, rets)
	})
}

func VMFuncThreeParamsOptionals[V1, V2, V3 VMValue](n int, f func(
	arg1 V1, arg2 V2, arg3 V3, rest VMSlice, rets *VMSlice) error,
) VMFunc {
	errs := []error{
		VMErrorNeedArgType[V1](0, 3),
		VMErrorNeedArgType[V2](1, 3),
		VMErrorNeedArgType[V3](2, 3),
	}
	validators := []argValidator{typeValidator[V1], typeValidator[V2], typeValidator[V3]}
	return vmFuncValidatedArgs(3, n, validators, errs, func(args VMSlice, rets *VMSlice) error {
		return f(args[0].(V1), args[1].(V2), args[2].(V3), args[3:], rets)
	})
}

// VMFuncNParams оборачивает функцию во враппер, который проверяет,
// что передано ровно n аргументов
func VMFuncNParams(n int, f VMMethod) VMFunc {
	return VMFuncNParamsOptionals(n, 0, f)
}

// VMFuncNParamsOptionals оборачивает функцию во враппер, который проверяет,
// что передано как минимум nreq аргументов и максимум nopt дополнительных
func VMFuncNParamsOptionals(nreq, nopt int, f VMMethod) VMFunc {
	if nreq+nopt == 0 {
		return VMFuncZeroParams(func(rets *VMSlice) error {
			return f(VMSlice{}, rets)
		})
	}

	needArgsErr := VMErrorNeedArgs(nreq)
	maxArgsErr := VMErrorMaxArgs(nreq + nopt)
	return VMFunc(func(args VMSlice, rets *VMSlice) error {
		if len(args) < nreq {
			return needArgsErr
		} else if len(args) <= nreq+nopt {
			return f(args, rets)
		} else {
			return maxArgsErr
		}
	})
}
