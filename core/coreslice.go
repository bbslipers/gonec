package core

import (
	"bytes"
	cryto "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"sync"

	"github.com/shinanca/gonec/names"
)

const ChunkVMSlicePool = 64

// globalVMSlicePool используется виртуальной машиной для переиспользования в регистрах и параметрах вызова
var globalVMSlicePool = sync.Pool{
	New: func() interface{} {
		return make(VMSlice, 0, ChunkVMSlicePool)
	},
}

func GetGlobalVMSlice() VMSlice {
	sl := globalVMSlicePool.Get()
	if sl != nil {
		return sl.(VMSlice)
	}
	return make(VMSlice, 0, ChunkVMSlicePool)
}

func PutGlobalVMSlice(sl VMSlice) {
	if cap(sl) <= ChunkVMSlicePool {
		sl = sl[:0]
		globalVMSlicePool.Put(sl)
	}
}

type VMSlice []VMValue

var ReflectVMSlice = reflect.TypeOf(make(VMSlice, 0))

func (x VMSlice) Len() int { return len(x) }

func (x VMSlice) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

func (x VMSlice) VMTypeString() string { return "Массив" }

func (x VMSlice) Interface() interface{} {
	return x
}

func (x VMSlice) Slice() VMSlice {
	return x
}

func (x VMSlice) BinaryType() VMBinaryType {
	return VMSLICE
}

func (x VMSlice) Args() []interface{} {
	ai := make([]interface{}, len(x))
	for i := range x {
		ai[i] = x[i]
	}
	return ai
}

func (x *VMSlice) Append(a ...VMValue) {
	*x = append(*x, a...)
}

func (x VMSlice) Length() VMInt {
	return VMInt(len(x))
}

func (x VMSlice) IndexVal(i VMValue) VMValue {
	if ii, ok := i.(VMInt); ok {
		return x[int(ii)]
	}
	panic("Индекс должен быть целым числом")
}

func (x VMSlice) Hash() VMString {
	b, err := x.MarshalBinary()
	if err != nil {
		panic(err)
	}
	h := make([]byte, 8)
	binary.LittleEndian.PutUint64(h, HashBytes(b))
	return VMString(hex.EncodeToString(h))
}

func (x VMSlice) SortDefault() {
	sort.Sort(VMSliceUpSort{x})
}

func (x VMSlice) SortRand() {
	sort.Sort(VMSliceRandSort{x})
}

func (x VMSlice) MethodMember(name int) (VMFunc, bool) {
	// только эти методы будут доступны из кода на языке Гонец!
	switch names.UniqueNames.GetLowerCase(name) {
	case "сортировать":
		return VMFuncZeroParams(x.Сортировать), true
	case "обратныйпорядок":
		return VMFuncZeroParams(x.ОбратныйПорядок), true
	case "копировать":
		return VMFuncZeroParams(x.Копировать), true
	case "найти":
		return VMFuncNParams(1, x.Найти), true
	//case "добавить":
	//return VMFuncOneParam(x.Добавить), true
	case "случайныйпорядок":
		return VMFuncZeroParams(x.СлучайныйПорядок), true
	// case "вставить":
	// 	return VMFuncTwoParams[VMInt, VMValue](x.Вставить), true
	// case "удалить":
	// 	return VMFuncOneParam[VMInt](x.Удалить), true
	case "копироватьуникальные":
		return VMFuncZeroParams(x.КопироватьУникальные), true
	}

	return nil, false
}

func (x VMSlice) Сортировать(rets *VMSlice) error {
	x.SortDefault()
	return nil
}

// Найти (значение) (индекс, найдено) - находит индекс значения или места для его вставки (конец списка), если его еще нет
// возврат унифицирован с возвратом функции НайтиСорт
func (x VMSlice) Найти(args VMSlice, rets *VMSlice) error {
	y := args[0]
	p := 0
	for p < len(x) {
		if EqualVMValues(x[p], y) {
			rets.Append(VMInt(p))
			return nil
		}
		p++
	}
	rets.Append(VMNilType{})
	return nil
}

func (x *VMSlice) Добавить(value VMValue, rets *VMSlice) error {
	x.Append(value)
	return nil
}

func (x VMSlice) СлучайныйПорядок(rets *VMSlice) error {
	x.SortRand()
	return nil
}

// Вставить (индекс, значение) - вставляет значение по индексу.
// Индекс может быть равен длине, тогда вставка происходит в последний элемент.
// Обычно используется в связке с НайтиСорт, т.к. позволяет вставлять значения с сохранением сортировки по возрастанию
// func (x VMSlice) Вставить(args VMSlice, rets *VMSlice) error {
// 	p := args[0].(VMInt)
// 	if int(p) < 0 || int(p) > len(x) {
// 		return VMErrorIndexOutOfBoundary
// 	}
// 	y := args[1]
// 	x = append(x, VMNil)
// 	copy(x[p+1:], x[p:])
// 	x[p] = y
// 	rets.Append(x)
// 	return nil
// }

// func (x VMSlice) Удалить(args VMSlice, rets *VMSlice) error {
// 	p := args[0].(VMInt)
// 	if int(p) < 0 || int(p) >= len(x) {
// 		return VMErrorIndexOutOfBoundary
// 	}
// 	copy(x[p:], x[p+1:])
// 	x[len(x)-1] = nil
// 	x = x[:len(x)-1]
// 	rets.Append(x)
// 	return nil
// }

func (x VMSlice) ОбратныйПорядок(rets *VMSlice) error {
	for left, right := 0, len(x)-1; left < right; left, right = left+1, right-1 {
		x[left], x[right] = x[right], x[left]
	}
	return nil
}

func (x VMSlice) CopyRecursive() VMSlice {
	rv := make(VMSlice, len(x))
	for i, v := range x {
		switch vv := v.(type) {
		case VMSlice:
			rv[i] = vv.CopyRecursive()
		case VMStringMap:
			rv[i] = vv.CopyRecursive()
		default:
			rv[i] = v
		}
	}
	return rv
}

// Копировать - помимо обычного копирования еще и рекурсивно копирует и слайсы/структуры, находящиеся в элементах
func (x VMSlice) Копировать(rets *VMSlice) error { // VMSlice {
	rv := make(VMSlice, len(x))
	copy(rv, x)
	for i, v := range rv {
		switch vv := v.(type) {
		case VMSlice:
			rv[i] = vv.CopyRecursive()
		case VMStringMap:
			rv[i] = vv.CopyRecursive()
		}
	}
	rets.Append(rv)
	return nil
}

func (x VMSlice) КопироватьУникальные(rets *VMSlice) error { // VMSlice {
	rv := make(VMSlice, 0)
	seen := make(map[VMValue]bool)
	for i, v := range x {
		if _, ok := seen[v]; ok {
			continue
		}
		switch vv := v.(type) {
		case VMSlice:
			rv = append(rv, vv.CopyRecursive())
		case VMStringMap:
			rv = append(rv, vv.CopyRecursive())
		default:
			rv = append(rv, x[i])
		}
		seen[v] = true
	}
	rets.Append(rv)
	return nil
}

func (x VMSlice) EvalBinOp(op VMOperation, y VMOperationer) (VMValue, error) {
	switch op {
	case ADD:
		switch yy := y.(type) {
		case VMSlice:
			// добавляем второй слайс в конец первого
			return append(x, yy...), nil
		case VMValue:
			return append(x, yy), nil
		}
		return append(x, y), nil
		// return VMNil, VMErrorIncorrectOperation
	case SUB:
		// удаляем из первого слайса любые элементы второго слайса, встречающиеся в первом
		switch yy := y.(type) {
		case VMSlice:
			// проходим слайс и переставляем ненайденные в вычитаемом слайсе элементы
			rv := make(VMSlice, len(x))
			il := 0
			for i := range x {
				fnd := false
				for j := range yy {
					if EqualVMValues(x[i], yy[j]) {
						fnd = true
						break
					}
				}
				if !fnd {
					rv[il] = x[i]
					il++
				}
			}
			return rv[:il], nil
		}
		return VMNil, VMErrorIncorrectOperation
	case MUL:
		return VMNil, VMErrorIncorrectOperation
	case QUO:
		return VMNil, VMErrorIncorrectOperation
	case REM:
		// оставляем только элементы, которые есть в первом и нет во втором и есть во втором но нет в первом
		// эквивалентно (С1 | С2) - (С1 & С2), или (С1-С2)|(С2-С1), или С2-(С1-С2), внешнее соединение

		switch yy := y.(type) {
		case VMSlice:
			rvx := make(VMSlice, len(x))
			rvy := make(VMSlice, len(yy))
			// С1-С2
			il := 0
			for i := range x {
				fnd := false
				for j := range yy {
					if EqualVMValues(x[i], yy[j]) {
						fnd = true
						break
					}
				}
				if !fnd {
					// оставляем
					rvx[il] = x[i]
					il++
				}
			}

			rvx = rvx[:il]

			// С2-(С1-C2)
			il = 0
			for j := range yy {
				fnd := false
				for i := range x {
					if EqualVMValues(x[i], yy[j]) {
						fnd = true
						break
					}
				}
				if !fnd {
					// оставляем
					rvy[il] = yy[j]
					il++
				}
			}

			rvy = rvy[:il]

			return append(rvx, rvy...), nil
		}

		return VMNil, VMErrorIncorrectOperation
	case EQL:
		// равенство по глубокому равенству элементов
		switch yy := y.(type) {
		case VMSlice:
			if len(x) != len(yy) {
				return VMBool(false), nil
			}
			for i := range x {
				if !EqualVMValues(x[i], yy[i]) {
					return VMBool(false), nil
				}
			}
			return VMBool(true), nil
		}
		return VMNil, VMErrorIncorrectOperation
	case NEQ:
		switch yy := y.(type) {
		case VMSlice:
			if len(x) != len(yy) {
				return VMBool(true), nil
			}
			for i := range x {
				if !EqualVMValues(x[i], yy[i]) {
					return VMBool(true), nil
				}
			}
			return VMBool(false), nil
		}
		return VMNil, VMErrorIncorrectOperation
	case GTR:
		return VMNil, VMErrorIncorrectOperation
	case GEQ:
		return VMNil, VMErrorIncorrectOperation
	case LSS:
		return VMNil, VMErrorIncorrectOperation
	case LEQ:
		return VMNil, VMErrorIncorrectOperation
	case OR:
		// добавляем в конец первого слайса только те элементы второго слайса, которые не встречаются в первом
		switch yy := y.(type) {
		case VMSlice:
			rv := x[:]
			for j := range yy {
				fnd := false
				for i := range x {
					if EqualVMValues(x[i], yy[j]) {
						fnd = true
						break
					}
				}
				if !fnd {
					rv = append(rv, yy[j])
				}
			}
			return rv, nil
		}
		return VMNil, VMErrorIncorrectOperation
	case LOR:
		return VMNil, VMErrorIncorrectOperation
	case AND:
		// оставляем только те элементы, которые есть в обоих слайсах
		switch yy := y.(type) {
		case VMSlice:
			rv := make(VMSlice, 0, len(x))
			for i := range x {
				fnd := false
				for j := range yy {
					if EqualVMValues(x[i], yy[j]) {
						fnd = true
						break
					}
				}
				if fnd {
					rv = append(rv, x[i])
				}
			}
			return rv, nil
		}
		return VMNil, VMErrorIncorrectOperation
	case LAND:
		return VMNil, VMErrorIncorrectOperation
	case POW:
		return VMNil, VMErrorIncorrectOperation
	case SHR:
		return VMNil, VMErrorIncorrectOperation
	case SHL:
		return VMNil, VMErrorIncorrectOperation
	}
	return VMNil, VMErrorUnknownOperation
}

func (x VMSlice) ConvertToType(nt reflect.Type) (VMValue, error) {
	switch nt {
	case ReflectVMString:
		// сериализуем в json
		b, err := json.Marshal(x)
		if err != nil {
			return VMNil, err
		}
		return VMString(string(b)), nil
		// case ReflectVMInt:
		// case ReflectVMTime:
		// case ReflectVMBool:
		// case ReflectVMDecNum:
	case ReflectVMSlice:
		return x, nil
		// case ReflectVMStringMap:
	}

	return VMNil, VMErrorNotConverted
}

func (x VMSlice) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	// количество элементов
	binary.Write(&buf, binary.LittleEndian, uint64(len(x)))
	for i := range x {
		if v, ok := x[i].(VMBinaryTyper); ok {
			bb, err := v.MarshalBinary()
			if err != nil {
				return nil, err
			}
			// тип
			buf.WriteByte(byte(v.BinaryType()))
			// длина
			binary.Write(&buf, binary.LittleEndian, uint64(len(bb)))
			// байты
			buf.Write(bb)
		} else {
			return nil, VMErrorNotBinaryConverted
		}
	}
	return buf.Bytes(), nil
}

func (x *VMSlice) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	var l, lv uint64
	// количество элементов
	if err := binary.Read(buf, binary.LittleEndian, &l); err != nil {
		return err
	}
	var rv VMSlice
	if x == nil || len(*x) < int(l) {
		rv = make(VMSlice, int(l))
	} else {
		rv = (*x)[:int(l)]
	}

	for i := 0; i < int(l); i++ {
		// тип
		if tt, err := buf.ReadByte(); err != nil {
			return err
		} else {
			// длина
			if err := binary.Read(buf, binary.LittleEndian, &lv); err != nil {
				return err
			}
			// байты
			bb := buf.Next(int(lv))
			vv, err := VMBinaryType(tt).ParseBinary(bb)
			if err != nil {
				return err
			}
			rv[i] = vv
		}
	}
	*x = rv
	return nil
}

func (x VMSlice) GobEncode() ([]byte, error) {
	return x.MarshalBinary()
}

func (x *VMSlice) GobDecode(data []byte) error {
	return x.UnmarshalBinary(data)
}

func (x VMSlice) String() string {
	b, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (x VMSlice) MarshalText() ([]byte, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (x *VMSlice) UnmarshalText(data []byte) error {
	sl, err := VMSliceFromJson(string(data))
	if err != nil {
		return err
	}
	*x = sl
	return nil
}

func (x VMSlice) MarshalJSON() ([]byte, error) {
	var err error
	rm := make([]json.RawMessage, len(x))
	for i, v := range x {
		rm[i], err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal(rm)
}

func (x *VMSlice) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	sl, err := VMSliceFromJson(string(data))
	if err != nil {
		return err
	}
	*x = sl
	return nil
}

// VMSliceUpSort - обертка для сортировки слайса по возрастанию
type VMSliceUpSort struct {
	VMSlice
}

func (x VMSliceUpSort) Less(i, j int) bool { return SortLessVMValues(x.VMSlice[i], x.VMSlice[j]) }

// VMSliceRandSort - обертка для сортировки слайса рандомом
type VMSliceRandSort struct {
	VMSlice
}

func (x VMSliceRandSort) Less(i, j int) bool {
	n1, err := cryto.Int(cryto.Reader, big.NewInt(int64(len(x.VMSlice))))
	if err != nil {
		fmt.Println("Ошибка генерации случайного числа:", err)
		return false
	}
	n2, err := cryto.Int(cryto.Reader, big.NewInt(int64(len(x.VMSlice))))
	if err != nil {
		fmt.Println("Ошибка генерации случайного числа:", err)
		return false
	}

	return n1.Int64() < n2.Int64()
}

// NewVMSliceFromStrings создает слайс вирт. машины []VMString из слайса строк []string на языке Го
func NewVMSliceFromStrings(ss []string) (rv VMSlice) {
	for i := range ss {
		rv = append(rv, VMString(ss[i]))
	}
	return
}
