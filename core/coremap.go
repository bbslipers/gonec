package core

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"reflect"

	"github.com/shinanca/gonec/names"
)

type VMStringMap map[string]VMValue

var ReflectVMStringMap = reflect.TypeOf(make(VMStringMap, 0))

func (x VMStringMap) VMTypeString() string { return "Структура" }

func (x VMStringMap) Interface() interface{} {
	return x
}

func (x VMStringMap) StringMap() VMStringMap {
	return x
}

func (x VMStringMap) Length() VMInt {
	return VMInt(len(x))
}

func (x VMStringMap) IndexVal(i VMValue) VMValue {
	if ii, ok := i.(VMStringer); ok {
		return x[ii.String()]
	}
	panic("Индекс должен быть строкой")
}

func (x VMStringMap) Index(i VMValue) VMValue {
	if s, ok := i.(VMString); ok {
		return x[string(s)]
	}
	panic("Индекс должен быть строкой")
}

func (x VMStringMap) BinaryType() VMBinaryType {
	return VMSTRINGMAP
}

func (x VMStringMap) Hash() VMString {
	b, err := x.MarshalBinary()
	if err != nil {
		panic(err)
	}
	h := make([]byte, 8)
	binary.LittleEndian.PutUint64(h, HashBytes(b))
	return VMString(hex.EncodeToString(h))
}

func (x VMStringMap) MethodMember(name int) (VMFunc, bool) {
	// только эти методы будут доступны из кода на языке Гонец!

	switch names.UniqueNames.GetLowerCase(name) {
	case "скопировать":
		return VMFuncZeroParams(x.Скопировать), true
	case "ключи":
		return VMFuncZeroParams(x.Ключи), true
	case "значения":
		return VMFuncZeroParams(x.Значения), true
	case "удалить":
		return VMFuncOneParam(x.Удалить), true
	}

	return nil, false
}

// Ключи возвращаются отсортированными по возрастанию
func (x VMStringMap) Ключи(rets *VMSlice) error { // VMSlice {
	rv := make(VMSlice, len(x))
	i := 0
	for k := range x {
		rv[i] = VMString(k)
		i++
	}
	rv.SortDefault()
	rets.Append(rv)
	return nil
}

// Значения возвращаются в случайном порядке
func (x VMStringMap) Значения(rets *VMSlice) error { // VMSlice {
	rv := make(VMSlice, len(x))
	i := 0
	for _, v := range x {
		rv[i] = v
		i++
	}
	rets.Append(rv)
	return nil
}

func (x VMStringMap) Удалить(key VMString, rets *VMSlice) error { // VMSlice {
	delete(x, string(key))
	return nil
}

func (x VMStringMap) CopyRecursive() VMStringMap {
	rv := make(VMStringMap, len(x))
	for k, v := range x {
		switch vv := v.(type) {
		case VMSlice:
			rv[k] = vv.CopyRecursive()
		case VMStringMap:
			rv[k] = vv.CopyRecursive()
		default:
			rv[k] = v
		}
	}
	return rv
}

func (x VMStringMap) Скопировать(rets *VMSlice) error {
	rv := x.CopyRecursive()
	rets.Append(rv)
	return nil
}

func (x VMStringMap) EvalBinOp(op VMOperation, y VMOperationer) (VMValue, error) {
	switch op {
	case ADD:
		// новые добавляются, а существующие обновляются
		switch yy := y.(type) {
		case VMStringMap:
			rv := make(VMStringMap)
			for k, v := range x {
				if _, ok := yy[k]; !ok {
					rv[k] = v
				}
			}
			for k, v := range yy {
				rv[k] = v
			}
			return rv, nil
		}
		return VMNil, VMErrorIncorrectOperation
	case SUB:
		switch yy := y.(type) {
		case VMStringMap:
			rv := make(VMStringMap)
			for k, v := range x {
				if _, ok := yy[k]; !ok {
					rv[k] = v
				}
			}
			return rv, nil
		}
		return VMNil, VMErrorIncorrectOperation
	case MUL:
		return VMNil, VMErrorIncorrectOperation
	case QUO:
		return VMNil, VMErrorIncorrectOperation
	case REM:
		switch yy := y.(type) {
		case VMStringMap:
			rv := make(VMStringMap)
			for k, v := range x {
				if _, ok := yy[k]; !ok {
					rv[k] = v
				}
			}
			for k := range yy {
				// if _, ok := rv[k]; ok {
				delete(rv, k)
				// }
			}
			return rv, nil
		}
		return VMNil, VMErrorIncorrectOperation
	case EQL:
		switch yy := y.(type) {
		case VMStringMap:
			if len(x) != len(yy) {
				return VMBool(false), nil
			}
			for k, v := range x {
				if yv, ok := yy[k]; ok {
					if !EqualVMValues(v, yv) {
						return VMBool(false), nil
					}
				} else {
					return VMBool(false), nil
				}
			}
			return VMBool(true), nil
		}
		return VMNil, VMErrorIncorrectOperation
	case NEQ:
		switch yy := y.(type) {
		case VMStringMap:
			if len(x) != len(yy) {
				return VMBool(true), nil
			}
			for k, v := range x {
				if yv, ok := yy[k]; ok {
					if !EqualVMValues(v, yv) {
						return VMBool(true), nil
					}
				} else {
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
		// только добавляем только те, которых еще нет, в отличие от ADD, где существующие обновляются
		switch yy := y.(type) {
		case VMStringMap:
			rv := make(VMStringMap)
			for k, v := range x {
				rv[k] = v
			}
			for k, v := range yy {
				if _, ok := rv[k]; !ok {
					rv[k] = v
				}
			}
			return rv, nil
		}
		return VMNil, VMErrorIncorrectOperation
	case LOR:
		return VMNil, VMErrorIncorrectOperation
	case AND:
		// оставляем только те элементы, которые есть в обоих структурах
		switch yy := y.(type) {
		case VMStringMap:
			rv := make(VMStringMap)
			for k, v := range x {
				if _, ok := yy[k]; ok {
					rv[k] = v
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

func (x VMStringMap) ConvertToType(nt reflect.Type) (VMValue, error) {
	// fmt.Println(nt)

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
	// case ReflectVMSlice:
	case ReflectVMStringMap:
		return x, nil
	}

	if nt.Kind() == reflect.Struct {
		rv := reflect.ValueOf(x)
		// для приведения в структурные типы - можно использовать мапу для заполнения полей
		rs := reflect.New(nt) // указатель на новую структуру
		// заполняем экспортируемые неанонимные поля, если их находим в мапе
		for i := 0; i < nt.NumField(); i++ {
			f := nt.Field(i)
			if f.PkgPath == "" && !f.Anonymous {
				setv := reflect.Indirect(rv.MapIndex(reflect.ValueOf(f.Name)))
				if setv.Kind() == reflect.Interface {
					setv = setv.Elem()
				}
				fv := rs.Elem().FieldByName(f.Name)
				if setv.IsValid() && fv.IsValid() && fv.CanSet() {
					if fv.Kind() != setv.Kind() {
						if setv.Type().ConvertibleTo(fv.Type()) {
							setv = setv.Convert(fv.Type())
						} else {
							return nil, VMErrorIncorrectFieldType
						}
					}
					fv.Set(setv)
				}
			}
		}
		if vv, ok := rs.Interface().(VMValue); ok {
			if vobj, ok := vv.(VMMetaObject); ok {
				vobj.VMInit(vobj)
				vobj.VMRegister()
				return vobj, nil
			} else {
				return nil, VMErrorIncorrectStructType
				// return vv, nil
			}
		} else {
			return nil, VMErrorUnknownType
		}

	}

	return VMNil, VMErrorNotConverted
}

func (x VMStringMap) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint64(len(x))) // количество пар ключ-значение
	for i := range x {
		if v, ok := x[i].(VMBinaryTyper); ok {
			bb, err := v.MarshalBinary()
			if err != nil {
				return nil, err
			}
			// ключ
			bws := []byte(i)
			binary.Write(&buf, binary.LittleEndian, uint64(len(bws))) // длина
			buf.Write(bws)                                            // строка ключа
			// значение
			buf.WriteByte(byte(v.BinaryType()))                      // тип
			binary.Write(&buf, binary.LittleEndian, uint64(len(bb))) // длина в байтах
			buf.Write(bb)                                            // байты
		} else {
			return nil, VMErrorNotBinaryConverted
		}
	}
	return buf.Bytes(), nil
}

func (x *VMStringMap) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	var l, li, lv uint64
	// количество пар
	if err := binary.Read(buf, binary.LittleEndian, &l); err != nil {
		return err
	}
	rv := make(VMStringMap, int(l))

	for i := 0; i < int(l); i++ {
		// длина ключа
		if err := binary.Read(buf, binary.LittleEndian, &li); err != nil {
			return err
		}
		// строка ключа
		bi := buf.Next(int(li))

		// тип
		if tt, err := buf.ReadByte(); err != nil {
			return err
		} else {
			// длина значения
			if err := binary.Read(buf, binary.LittleEndian, &lv); err != nil {
				return err
			}
			// байты значения
			bb := buf.Next(int(lv))

			vv, err := VMBinaryType(tt).ParseBinary(bb)
			if err != nil {
				return err
			}
			rv[string(bi)] = vv
		}
	}
	*x = rv
	return nil
}

func (x VMStringMap) GobEncode() ([]byte, error) {
	return x.MarshalBinary()
}

func (x *VMStringMap) GobDecode(data []byte) error {
	return x.UnmarshalBinary(data)
}

func (x VMStringMap) String() string {
	b, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (x VMStringMap) MarshalText() ([]byte, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (x *VMStringMap) UnmarshalText(data []byte) error {
	sm, err := VMStringMapFromJson(string(data))
	if err != nil {
		return err
	}
	*x = sm
	return nil
}

func (x VMStringMap) MarshalJSON() ([]byte, error) {
	var err error
	rm := make(map[string]json.RawMessage, len(x))
	for k, v := range x {
		rm[k], err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal(rm)
}

func (x *VMStringMap) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	sm, err := VMStringMapFromJson(string(data))
	if err != nil {
		return err
	}
	*x = sm
	return nil
}
