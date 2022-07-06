// Package core implements core interface for gonec script.
package core

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/covrom/decnum"

	uuid "github.com/satori/go.uuid"
	"github.com/shinanca/gonec/names"
)

// LoadAllBuiltins is a convenience function that loads all defineSd builtins.
func LoadAllBuiltins(env *Env) {
	Import(env)

	pkgs := map[string]func(env *Env) *Env{
		// "sort":          gonec_sort.Import,
		// "strings":       gonec_strings.Import,
	}

	env.DefineS("импорт", VMFuncOneParam[VMString](func(
		args VMSlice, rets *VMSlice, envout *(*Env),
	) error {
		*envout = env
		s := args[0].(VMString)
		if loader, ok := pkgs[strings.ToLower(string(s))]; ok {
			rets.Append(loader(env)) // возвращает окружение, инициализированное пакетом
			return nil
		}
		return fmt.Errorf("Пакет '%s' не найден", s)
	}))

	// успешно загружен глобальный контекст
	env.SetBuiltsIsLoaded()
}

// Import общая стандартная бибилиотека
func Import(env *Env) *Env {
	env.DefineS("длина", VMFuncOneParam[VMIndexer](func(args VMSlice, rets *VMSlice, envout *(*Env)) error {
		*envout = env
		rets.Append(args[0].(VMIndexer).Length())
		return nil
	}))

	env.DefineS("диапазон", VMFunc(func(args VMSlice, rets *VMSlice, envout *(*Env)) error {
		*envout = env
		if len(args) < 1 {
			return VMErrorNoArgs
		}
		if len(args) > 2 {
			return VMErrorNeedLengthOrBoundary
		}
		var min, max int64
		var arr VMSlice
		if len(args) == 1 {
			min = 0
			maxvm, ok := args[0].(VMInt)
			if !ok {
				return VMErrorNeedInt
			}

			max = maxvm.Int()
			if max == 0 {
				rets.Append(make(VMSlice, 0))
				return nil
			} else if max < 0 {
				return errors.New("Диапазон не может быть до отрицательного числа")
			}
			max--
		} else {
			minvm, ok := args[0].(VMInt)
			if !ok {
				return VMErrorNeedInt
			}
			min = minvm.Int()
			maxvm, ok := args[1].(VMInt)
			if !ok {
				return VMErrorNeedInt
			}
			max = maxvm.Int()
		}
		if min > max {
			return VMErrorNeedLess
		}
		arr = make(VMSlice, max-min+1)

		for i := min; i <= max; i++ {
			arr[i-min] = VMInt(i)
		}
		rets.Append(arr)
		return nil
	}))

	env.DefineS("текущаядата", VMFuncZeroParams(func(
		args VMSlice, rets *VMSlice, envout *(*Env),
	) error {
		*envout = env
		rets.Append(Now())
		return nil
	}))

	env.DefineS("прошловременис", VMFuncOneParam[VMDateTimer](func(
		args VMSlice, rets *VMSlice, envout *(*Env),
	) error {
		*envout = env
		rets.Append(Now().Sub(args[0].(VMDateTimer).Time()))
		return nil
	}))

	env.DefineS("пауза", VMFuncOneParam[VMNumberer](func(
		args VMSlice, rets *VMSlice, envout *(*Env),
	) error {
		*envout = env
		sec1 := NewVMDecNumFromInt64(int64(VMSecond))
		time.Sleep(time.Duration(args[0].(VMNumberer).DecNum().Mul(sec1).Int()))
		return nil
	}))

	env.DefineS("длительностьнаносекунды", VMNanosecond)
	env.DefineS("длительностьмикросекунды", VMMicrosecond)
	env.DefineS("длительностьмиллисекунды", VMMillisecond)
	env.DefineS("длительностьсекунды", VMSecond)
	env.DefineS("длительностьминуты", VMMinute)
	env.DefineS("длительностьчаса", VMHour)
	env.DefineS("длительностьдня", VMDay)

	env.DefineS("хэш", VMFuncOneParam[VMHasher](func(
		args VMSlice, rets *VMSlice, envout *(*Env),
	) error {
		*envout = env
		rets.Append(args[0].(VMHasher).Hash())
		return nil
	}))

	env.DefineS("уникальныйидентификатор", VMFuncZeroParams(func(
		args VMSlice, rets *VMSlice, envout *(*Env),
	) error {
		*envout = env
		rets.Append(VMString(uuid.NewV1().String()))
		return nil
	}))

	env.DefineS("получитьмассивизпула", VMFuncZeroParams(func(
		args VMSlice, rets *VMSlice, envout *(*Env),
	) error {
		*envout = env
		rets.Append(GetGlobalVMSlice())
		return nil
	}))

	env.DefineS("вернутьмассиввпул", VMFuncOneParam[VMSlice](func(
		args VMSlice, rets *VMSlice, envout *(*Env),
	) error {
		*envout = env
		PutGlobalVMSlice(args[0].(VMSlice))
		return nil
	}))

	env.DefineS("случайнаястрока", VMFuncOneParam[VMInt](func(
		args VMSlice, rets *VMSlice, envout *(*Env),
	) error {
		*envout = env
		rets.Append(VMString(MustGenerateRandomString(int(args[0].(VMInt)))))
		return nil
	}))

	env.DefineS("нрег", VMFuncOneParam[VMStringer](func(
		args VMSlice, rets *VMSlice, envout *(*Env),
	) error {
		*envout = env
		rets.Append(VMString(strings.ToLower(args[0].(VMStringer).String())))
		return nil
	}))

	env.DefineS("врег", VMFuncOneParam[VMStringer](func(
		args VMSlice, rets *VMSlice, envout *(*Env),
	) error {
		*envout = env
		rets.Append(VMString(strings.ToUpper(args[0].(VMStringer).String())))
		return nil
	}))

	env.DefineS("стрсодержит", VMFuncTwoParams[VMStringer, VMStringer](func(
		args VMSlice, rets *VMSlice, envout *(*Env),
	) error {
		*envout = env
		rets.Append(VMBool(strings.Contains(
			args[0].(VMStringer).String(),
			args[1].(VMStringer).String())))
		return nil
	}))

	env.DefineS("стрсодержитлюбой", VMFuncTwoParams[VMStringer, VMStringer](func(
		args VMSlice, rets *VMSlice, envout *(*Env),
	) error {
		*envout = env
		rets.Append(VMBool(strings.ContainsAny(
			args[0].(VMStringer).String(),
			args[1].(VMStringer).String())))
		return nil
	}))

	env.DefineS("стрколичество", VMFuncTwoParams[VMStringer, VMStringer](func(
		args VMSlice, rets *VMSlice, envout *(*Env),
	) error {
		*envout = env
		rets.Append(VMInt(strings.Count(
			args[0].(VMStringer).String(),
			args[1].(VMStringer).String())))
		return nil
	}))

	env.DefineS("стрнайти", VMFuncTwoParams[VMStringer, VMStringer](func(
		args VMSlice, rets *VMSlice, envout *(*Env),
	) error {
		*envout = env
		rets.Append(VMInt(strings.Index(
			args[0].(VMStringer).String(),
			args[1].(VMStringer).String())))
		return nil
	}))

	env.DefineS("стрнайтилюбой", VMFuncTwoParams[VMStringer, VMStringer](func(
		args VMSlice, rets *VMSlice, envout *(*Env),
	) error {
		*envout = env
		rets.Append(VMInt(strings.IndexAny(
			args[0].(VMStringer).String(),
			args[1].(VMStringer).String())))
		return nil
	}))

	env.DefineS("стрнайтипоследний", VMFuncTwoParams[VMStringer, VMStringer](func(
		args VMSlice, rets *VMSlice, envout *(*Env),
	) error {
		*envout = env
		rets.Append(VMInt(strings.LastIndex(
			args[0].(VMStringer).String(),
			args[1].(VMStringer).String())))
		return nil
	}))

	env.DefineS("стрзаменить", VMFuncThreeParams[VMStringer, VMStringer, VMStringer](func(
		args VMSlice, rets *VMSlice, envout *(*Env),
	) error {
		*envout = env
		rets.Append(VMString(strings.Replace(
			args[0].(VMStringer).String(),
			args[1].(VMStringer).String(),
			args[2].(VMStringer).String(), -1)))
		return nil
	}))

	env.DefineS("окр", VMFuncTwoParams[VMDecNum, VMInt](func(args VMSlice, rets *VMSlice, envout *(*Env)) error {
		*envout = env
		rets.Append(VMDecNum{num: args[0].(VMDecNum).num.RoundWithMode(
			int32(args[1].(VMInt)), decnum.RoundHalfUp)})
		return nil
	}))

	env.DefineS("формат", VMFunc(func(args VMSlice, rets *VMSlice, envout *(*Env)) error {
		*envout = env
		if len(args) < 2 {
			return VMErrorNeedFormatAndArgs
		}
		if v, ok := args[0].(VMString); ok {
			as := VMSlice(args[1:]).Args()
			rets.Append(VMString(env.Sprintf(string(v), as...)))
			return nil
		}
		return VMErrorNeedString
	}))

	env.DefineS("кодсимвола", VMFuncOneParam[VMStringer](func(
		args VMSlice, rets *VMSlice, envout *(*Env),
	) error {
		*envout = env
		s := args[0].(VMStringer).String()
		if len(s) == 0 {
			rets.Append(VMInt(0))
		} else {
			rets.Append(VMInt([]rune(s)[0]))
		}
		return nil
	}))

	env.DefineS("типзнч", VMFuncNParams(1, func(args VMSlice, rets *VMSlice, envout *(*Env)) error {
		*envout = env
		if args[0] == nil || args[0] == VMNil {
			rets.Append(VMString("Неопределено"))
			return nil
		}
		rets.Append(VMString(names.UniqueNames.Get(env.TypeName(reflect.TypeOf(args[0])))))
		return nil
	}))

	env.DefineS("сообщить", VMFunc(func(args VMSlice, rets *VMSlice, envout *(*Env)) error {
		*envout = env
		if len(args) == 0 {
			env.Println()
			return nil
		}
		as := args.Args()
		env.Println(as...)
		return nil
	}))

	env.DefineS("сообщитьф", VMFunc(func(args VMSlice, rets *VMSlice, envout *(*Env)) error {
		*envout = env
		if len(args) < 2 {
			return VMErrorNeedFormatAndArgs
		}
		if v, ok := args[0].(VMString); ok {
			as := VMSlice(args[1:]).Args()
			env.Printf(string(v), as...)
			return nil
		}
		return VMErrorNeedString
	}))

	env.DefineS("обработатьгорутины", VMFuncZeroParams(func(
		args VMSlice, rets *VMSlice, envout *(*Env),
	) error {
		*envout = env
		runtime.Gosched()
		return nil
	}))

	env.DefineS("переменнаяокружения", VMFuncOneParam[VMString](func(
		args VMSlice, rets *VMSlice, envout *(*Env),
	) error {
		*envout = env
		val, ok := os.LookupEnv(string(args[0].(VMString)))
		rets.Append(VMString(val))
		rets.Append(VMBool(ok))
		return nil
	}))

	// при изменении состава типов не забывать изменять их и в lexer.go
	env.DefineTypeS(ReflectVMInt)
	env.DefineTypeS(ReflectVMDecNum)
	env.DefineTypeS(ReflectVMBool)
	env.DefineTypeS(ReflectVMString)
	env.DefineTypeS(ReflectVMSlice)
	env.DefineTypeS(ReflectVMStringMap)
	env.DefineTypeS(ReflectVMTime)
	env.DefineTypeS(ReflectVMTimeDuration)
	env.DefineTypeS(ReflectVMFunc)

	env.DefineTypeS(ReflectVMWaitGroup)
	env.DefineTypeS(ReflectVMBoltDB)

	env.DefineTypeStruct(&VMServer{})
	env.DefineTypeStruct(&VMClient{})

	env.DefineTypeStruct(&RconClient{})
	env.DefineTypeStruct(&TextDocument{})
	env.DefineTypeStruct(&File{})

	env.DefineTypeStruct(&VMTable{})
	env.DefineTypeStruct(&VMTableColumn{})
	env.DefineTypeStruct(&VMTableColumns{})
	env.DefineTypeStruct(&VMTableLine{})

	//////////////////
	env.DefineTypeStruct(&TttStructTest{})

	env.DefineS("__дамп__", VMFunc(func(args VMSlice, rets *VMSlice, envout *(*Env)) error {
		*envout = env
		env.Dump()
		return nil
	}))
	/////////////////////

	return env
}

/////////////////
// TttStructTest - тестовая структура для отладки работы с системными функциональными структурами
type TttStructTest struct {
	VMMetaObj

	ПолеЦелоеЧисло VMInt
	ПолеСтрока     VMString
}

func (tst *TttStructTest) VMTypeString() string {
	return "__ФункциональнаяСтруктураТест__"
}

func (tst *TttStructTest) VMRegister() {
	tst.VMRegisterMethod("ВСтроку", tst.ВСтроку)
	tst.VMRegisterField("ПолеЦелоеЧисло", &tst.ПолеЦелоеЧисло)
	tst.VMRegisterField("ПолеСтрока", &tst.ПолеСтрока)
}

// обратите внимание - русскоязычное название метода для структуры и формат для быстрого вызова
func (tst *TttStructTest) ВСтроку(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	rets.Append(VMString(fmt.Sprintf("ПолеЦелоеЧисло=%v, ПолеСтрока=%v", tst.ПолеЦелоеЧисло, tst.ПолеСтрока)))
	return nil
}

/////////////////
