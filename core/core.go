// Package core implements core interface for gonec script.
package core

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/covrom/decnum"
	uuid "github.com/satori/go.uuid"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/shinanca/gonec/externallibs/go-xlsx-templater"
	"github.com/shinanca/gonec/names"
	"moul.io/number-to-words"
)

func fractionWordMap(decimalPlaces int) string {
	fractionWords := map[int]string{
		1: "дес.",
		2: "сот.",
		3: "тыс.",
		4: "десятитыс.",
		5: "стотыс.",
		6: "миллионных",
		7: "десятимиллионных",
		8: "стомиллионных",
	}

	word, found := fractionWords[decimalPlaces]
	if !found {
		return "Неизвестное количество знаков после запятой"
	}

	return word
}

func intToWords(num int) string {
	numStr := ntw.IntegerToRuRu(num)
	words := strings.Fields(numStr)

	// Проверяем, что есть хотя бы одно слово в строке
	if len(words) > 0 {
		// Получаем последнее слово
		lastWord := words[len(words)-1]

		switch lastWord {
		case "один":
			words[len(words)-1] = "одна"
		case "два":
			words[len(words)-1] = "две"
		}
	}

	// Собираем строку обратно
	resultString := strings.Join(words, " ")
	return resultString
}

func XmlToJson(xmlString string) (string, error) {
	node := &XMLNode{}

	if err := xml.Unmarshal([]byte(xmlString), node); err != nil {
		return "", VMErrorIncorrectFieldType
	}

	json, err := node.MarshalJSON()
	if err != nil {
		return "", VMErrorIncorrectFieldType
	}

	return string(json), nil
}

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

	env.DefineS("округлить", VMFuncTwoParams(func(f VMValue, n VMInt, rets *VMSlice) error {
		number, ok := f.(VMInt)
		if ok {
			rets.Append(number)
			return nil
		}
		decNum, ok := f.(VMDecNum)
		if !ok {
			return VMErrorNeedDecNum
		}
		if n == 0 {
			rets.Append(VMDecNum{num: decNum.num.RoundWithMode(int32(n), decnum.RoundDown)})
		} else if n < 0 {
			rets.Append(VMDecNum{num: decNum.num.RoundWithMode(int32(n)*(-1), decnum.RoundHalfDown)})
		} else {
			rets.Append(VMDecNum{num: decNum.num.RoundWithMode(int32(n), decnum.RoundHalfUp)})
		}
		return nil
	}))

	env.DefineS("длиначисла", VMFuncOneParam(func(f VMDecNum, rets *VMSlice) error {
		rets.Append(VMInt(len(f.String())))
		return nil
	}))

	env.DefineS("точностьчисла", VMFuncOneParam(func(f VMDecNum, rets *VMSlice) error {
		numStr := f.String()
		dotIndex := strings.Index(numStr, ".")

		if dotIndex == -1 {
			rets.Append(VMInt(0))
		} else {
			rets.Append(VMInt(len(numStr) - dotIndex - 1))
		}
		return nil
	}))

	env.DefineS("числовстроку", VMFuncNParams(5, func(args VMSlice, rets *VMSlice) error {
		if len(args) != 5 {
			return VMErrorNeedArgs(5)
		}
		number, ok := args[0].(VMDecNum)
		if !ok {
			intnumber, ok := args[0].(VMInt)
			if !ok {
				return VMErrorIncorrectFieldType
			}
			number.ParseGoType(intnumber)
		}
		length, ok := args[1].(VMInt)
		if !ok {
			return VMErrorNeedInt
		}
		accuracy, ok := args[2].(VMInt)
		if !ok {
			return VMErrorNeedInt
		}
		needLeadingNuls, ok := args[3].(VMBool)
		if !ok {
			return VMErrorNeedInt
		}
		NulValue, ok := args[4].(VMString)
		if !ok {
			return VMErrorNeedInt
		}

		if number.Int() == 0 {
			rets.Append(NulValue)
			return nil
		}

		numberString := number.String()
		dotIndex := strings.Index(numberString, ".")
		if dotIndex == -1 && accuracy > 0 {
			numberString = numberString + "."
			dotIndex = len(numberString) - 1
		}

		decimalPlaces := len(numberString) - dotIndex - 1
		if decimalPlaces > int(accuracy) {
			numberString = numberString[:dotIndex+int(accuracy)+1]
		} else {
			numberString = numberString + strings.Repeat("0", int(accuracy)-decimalPlaces)
		}

		if len(numberString) > int(length) {
			return fmt.Errorf("Oшибка: Длина числа больше чем %d", length)
		}

		if needLeadingNuls {
			numberString = strings.Repeat("0", int(length)-len(numberString)) + numberString
		}

		rets.Append(VMString(numberString))

		return nil
	}))

	env.DefineS("числопрописью", VMFuncOneParam(func(f VMDecNum, rets *VMSlice) error {
		stringNum := f.String()
		dotIndex := strings.Index(stringNum, ".")
		floatPart := ""
		if dotIndex != -1 {
			decimalPartString := stringNum[dotIndex+1:]
			decimalPart, _ := strconv.Atoi(decimalPartString)
			floatPart = intToWords(decimalPart) + " " + fractionWordMap(len(decimalPartString))
		}

		res := intToWords(int(f.Int())) + " " + "цел." + " " + floatPart

		r, i := utf8.DecodeRuneInString(res)
		res = string(unicode.ToTitle(r)) + res[i:]

		rets.Append(VMString(res))
		return nil
	}))

	env.DefineS("суммапрописью", VMFuncOneParam(func(f VMDecNum, rets *VMSlice) error {
		stringNum := f.String()
		dotIndex := strings.Index(stringNum, ".")
		floatPart := ""
		if dotIndex != -1 {
			decimalPartString := stringNum[dotIndex+1:]
			decimalPart, _ := strconv.Atoi(decimalPartString)
			decimalPartStr := intToWords(decimalPart)
			if decimalPart > 99 {
				return fmt.Errorf("Oшибка: Неверное представление количества копеек")
			}
			floatPart = decimalPartStr + " " + "коп."
		}

		res := ntw.IntegerToRuRu(int(f.Int())) + " " + "руб." + " " + floatPart
		r, i := utf8.DecodeRuneInString(res)
		res = string(unicode.ToTitle(r)) + res[i:]

		rets.Append(VMString(res))
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
		rets.Append(VMString(names.UniqueNames.Get(env.TypeName(reflect.TypeOf(args[0])))))
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

	env.DefineS("ЗаполнитьМакетXLSX", VMFunc(func(args VMSlice, rets *VMSlice) error {
		if len(args) != 3 {
			return VMErrorNeedArgs(3)
		}
		doc := xlst.New()
		err := doc.ReadTemplate(string(args[0].(VMString)))
		if err != nil {
			return VMErrorReadXlsxTemplateError
		}

		data := args[1].(VMStringMap)
		jsonStr, err := data.MarshalJSON()
		if err != nil {
			return VMErrorFillXlsxError
		}
		var newData map[string]interface{}
		if err := json.Unmarshal(jsonStr, &newData); err != nil {
			return VMErrorFillXlsxError
		}
		err = doc.Render(newData)
		if err != nil {
			return VMErrorFillXlsxError
		}

		err = doc.Save(string(args[2].(VMString)))
		if err != nil {
			return VMErrorSaveXlsxError
		}
		return nil
	}))

	env.DefineS("чтениеизстрокиxml", VMFunc(func(args VMSlice, rets *VMSlice) error {
		if len(args) != 1 {
			env.Println()
			return nil
		}

		json, err := XmlToJson(string(args[0].(VMString)))
		if err != nil {
			return VMErrorIncorrectFieldType
		}

		rets.Append(VMString(json))

		return nil
	}))

	env.DefineS("чтениеизфайлаxml", VMFunc(func(args VMSlice, rets *VMSlice) error {
		if len(args) != 1 {
			env.Println()
			return nil
		}

		xmlString, err := os.ReadFile(string(args[0].(VMString)))
		if err != nil {
			return VMErrorIncorrectFieldType
		}

		json, err := XmlToJson(string(xmlString))
		if err != nil {
			return VMErrorIncorrectFieldType
		}

		rets.Append(VMString(json))

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

	env.DefineS("текущаядата", VMFuncZeroParams(func(rets *VMSlice) error {
		rets.Append(VMTime(time.Now()))
		return nil
	}))

	// при изменении состава типов не забывать изменять их и в lexer.go
	env.DefineTypeS(ReflectVMNul)
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

	env.DefineTypeStruct(&QrCode{})
	env.DefineTypeStruct(&DataMatrix{})
	env.DefineTypeStruct(&I2of5{})
	env.DefineTypeStruct(&Ean13{})

	env.DefineTypeStruct(&VMTable{})
	env.DefineTypeStruct(&VMTableColumn{})
	env.DefineTypeStruct(&VMTableColumns{})
	env.DefineTypeStruct(&VMTableLine{})

	env.DefineTypeStruct(&EmailProfile{})
	env.DefineTypeStruct(&EmailData{})

	env.DefineTypeStruct(&XMLDoc{})
	env.DefineTypeStruct(&XMLElem{})

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
