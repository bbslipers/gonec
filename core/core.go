// Package core implements core interface for gonec script.
package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

	env.DefineS("импорт", VMFuncOneParam(func(s VMString, rets *VMSlice) error {
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
	env.DefineS("длина", VMFuncOneParam(func(v VMIndexer, rets *VMSlice) error {
		rets.Append(v.Length())
		return nil
	}))

	env.DefineS("диапазон", VMFuncOneParamOptionals(1, func(arg1 VMInt, rest VMSlice,
		rets *VMSlice,
	) error {
		var min, max int64
		var arr VMSlice
		if len(rest) == 0 {
			min, max = 0, arg1.Int()
			if max == 0 {
				rets.Append(make(VMSlice, 0))
				return nil
			} else if max < 0 {
				return errors.New("Диапазон не может быть до отрицательного числа")
			}
			max--
		} else {
			min = arg1.Int()
			if maxvm, ok := rest[0].(VMInt); ok {
				max = maxvm.Int()
			} else {
				return VMErrorNeedInt
			}
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

	env.DefineS("текущаядата", VMFuncZeroParams(func(rets *VMSlice) error {
		rets.Append(Now())
		return nil
	}))

	env.DefineS("прошловременис", VMFuncOneParam(func(date VMDateTimer, rets *VMSlice) error {
		rets.Append(Now().Sub(date.Time()))
		return nil
	}))

	env.DefineS("пауза", VMFuncOneParam(func(n VMNumberer, rets *VMSlice) error {
		sec1 := NewVMDecNumFromInt64(int64(VMSecond))
		time.Sleep(time.Duration(n.DecNum().Mul(sec1).Int()))
		return nil
	}))

	env.DefineS("длительностьнаносекунды", VMNanosecond)
	env.DefineS("длительностьмикросекунды", VMMicrosecond)
	env.DefineS("длительностьмиллисекунды", VMMillisecond)
	env.DefineS("длительностьсекунды", VMSecond)
	env.DefineS("длительностьминуты", VMMinute)
	env.DefineS("длительностьчаса", VMHour)
	env.DefineS("длительностьдня", VMDay)

	env.DefineS("хэш", VMFuncOneParam(func(h VMHasher, rets *VMSlice) error {
		rets.Append(h.Hash())
		return nil
	}))

	env.DefineS("уникальныйидентификатор", VMFuncZeroParams(func(rets *VMSlice) error {
		rets.Append(VMString(uuid.NewV1().String()))
		return nil
	}))

	env.DefineS("получитьмассивизпула", VMFuncZeroParams(func(rets *VMSlice) error {
		rets.Append(GetGlobalVMSlice())
		return nil
	}))

	env.DefineS("вернутьмассиввпул", VMFuncOneParam(func(args VMSlice, rets *VMSlice) error {
		PutGlobalVMSlice(args[0].(VMSlice))
		return nil
	}))

	env.DefineS("окр", VMFuncTwoParams(func(f VMDecNum, n VMInt, rets *VMSlice) error {
		rets.Append(VMDecNum{num: f.num.RoundWithMode(int32(n), decnum.RoundHalfUp)})
		return nil
	}))

	env.DefineS("формат", VMFunc(func(args VMSlice, rets *VMSlice) error {
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

	env.DefineS("кодсимвола", VMFuncOneParam(func(vms VMStringer, rets *VMSlice) error {
		s := vms.String()
		if len(s) == 0 {
			rets.Append(VMInt(0))
		} else {
			rets.Append(VMInt([]rune(s)[0]))
		}
		return nil
	}))

	env.DefineS("типзнч", VMFuncNParams(1, func(args VMSlice, rets *VMSlice) error {
		if args[0] == nil || args[0] == VMNil {
			rets.Append(VMString("Неопределено"))
			return nil
		}
		rets.Append(VMString(names.UniqueNames.GetLowerCase(env.TypeName(reflect.TypeOf(args[0])))))
		return nil
	}))

	env.DefineS("сообщить", VMFunc(func(args VMSlice, rets *VMSlice) error {
		if len(args) == 0 {
			env.Println()
			return nil
		}
		as := args.Args()
		env.Println(as...)
		return nil
	}))

	env.DefineS("сообщитьф", VMFunc(func(args VMSlice, rets *VMSlice) error {
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

	env.DefineS("обработатьгорутины", VMFuncZeroParams(func(rets *VMSlice) error {
		runtime.Gosched()
		return nil
	}))

	env.DefineS("переменнаяокружения", VMFuncOneParam(func(s VMString, rets *VMSlice) error {
		val, ok := os.LookupEnv(string(s))
		rets.Append(VMString(val))
		rets.Append(VMBool(ok))
		return nil
	}))

	env.DefineS("создатьвременныйфайл", VMFuncZeroParams(func(rets *VMSlice) error {
		file, err := os.CreateTemp("", "gonectmp")
		defer file.Close()
		if err != nil {
			return fmt.Errorf("Не удалось создать временный файл: %s", err.Error())
		}

		path, err := filepath.Abs(file.Name())
		if err != nil {
			return fmt.Errorf("Не удалось получить путь до временного файла: %s", err.Error())
		}

		rets.Append(VMString(path))
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
	env.DefineTypeStruct(&VMSqliteDB{})

	env.DefineTypeStruct(&VMServer{})
	env.DefineTypeStruct(&VMClient{})

	env.DefineTypeStruct(&RconClient{})
	env.DefineTypeStruct(&TextDocument{})
	env.DefineTypeStruct(&File{})

	env.DefineTypeStruct(&VMTable{})
	env.DefineTypeStruct(&VMTableColumn{})
	env.DefineTypeStruct(&VMTableColumns{})
	env.DefineTypeStruct(&VMTableLine{})

	env.DefineTypeStruct(&EmailProfile{})
	env.DefineTypeStruct(&EmailData{})

	ImportStrings(env)

	//////////////////
	env.DefineTypeStruct(&TttStructTest{})

	env.DefineS("__дамп__", VMFunc(func(args VMSlice, rets *VMSlice) error {
		env.Dump()
		return nil
	}))
	/////////////////////

	return env
}

// ///////////////
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
func (tst *TttStructTest) ВСтроку(args VMSlice, rets *VMSlice) error {
	rets.Append(VMString(fmt.Sprintf("ПолеЦелоеЧисло=%v, ПолеСтрока=%v", tst.ПолеЦелоеЧисло, tst.ПолеСтрока)))
	return nil
}

/////////////////
