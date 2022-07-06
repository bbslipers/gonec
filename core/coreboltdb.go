package core

import (
	"bytes"
	"reflect"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/shinanca/gonec/names"
)

// VMBoltDB - группа ожидания исполнения горутин
type VMBoltDB struct {
	sync.Mutex
	name string
	db   *bolt.DB
}

var ReflectVMBoltDB = reflect.TypeOf(VMBoltDB{})

func (x *VMBoltDB) VMTypeString() string { return "ФайловаяБазаДанных" }

func (x *VMBoltDB) Interface() interface{} {
	return x
}

func (x *VMBoltDB) String() string {
	return "Файловая база данных BoltDB " + x.name
}

func (x *VMBoltDB) Open(filename string) (err error) {
	x.Lock()
	defer x.Unlock()
	x.db, err = bolt.Open(filename, 0o600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	x.name = filename
	return nil
}

func (x *VMBoltDB) Close() {
	x.Lock()
	defer x.Unlock()
	if x.db != nil {
		x.db.Close()
		x.db = nil
	}
}

func (x *VMBoltDB) Begin(writable bool) (tr *VMBoltTransaction, err error) {
	x.Lock()
	defer x.Unlock()
	var tx *bolt.Tx
	tx, err = x.db.Begin(writable)
	if err != nil {
		return tr, err
	}
	tr = &VMBoltTransaction{tx: tx, writable: writable}
	return
}

func (x *VMBoltDB) MethodMember(name int) (VMFunc, bool) {
	// только эти методы будут доступны из кода на языке Гонец!
	switch names.UniqueNames.GetLowerCase(name) {
	case "открыть":
		return VMFuncOneParam(x.Открыть), true
	case "закрыть":
		return VMFuncZeroParams(x.Закрыть), true
	case "начатьтранзакцию":
		return VMFuncOneParam(x.НачатьТранзакцию), true
	}
	return nil, false
}

func (x *VMBoltDB) Открыть(s VMString, rets *VMSlice) error {
	return x.Open(string(s))
}

func (x *VMBoltDB) НачатьТранзакцию(writable VMBool, rets *VMSlice) error {
	tr, err := x.Begin(writable.Bool())
	if err != nil {
		return err
	}
	rets.Append(tr)
	return nil
}

func (x *VMBoltDB) Закрыть(rets *VMSlice) error {
	x.Close()
	return nil
}

// VMBoltTransaction реализует функционал Transaction для BoltDB
type VMBoltTransaction struct {
	tx       *bolt.Tx
	writable bool
}

func (x *VMBoltTransaction) VMTypeString() string {
	return "ТранзакцияФайловойБазыДанных"
}

func (x *VMBoltTransaction) Interface() interface{} {
	return x
}

func (x *VMBoltTransaction) String() string {
	return "Транзакция файловой базы данных BoltDB"
}

func (x *VMBoltTransaction) Commit() error {
	if x.tx == nil {
		return VMErrorTransactionNotOpened
	}
	err := x.tx.Commit()
	x.tx = nil
	return err
}

func (x *VMBoltTransaction) Rollback() error {
	if x.tx == nil {
		return VMErrorTransactionNotOpened
	}
	x.tx.Rollback()
	x.tx = nil
	return nil
}

func (x *VMBoltTransaction) CreateTableIfNotExists(name string) (*VMBoltTable, error) {
	if x.tx == nil {
		return nil, VMErrorTransactionNotOpened
	}
	if x.writable {
		b, err := x.tx.CreateBucketIfNotExists([]byte(name))
		t := &VMBoltTable{name: name, b: b}
		return t, err
	} else {
		return x.OpenTable(name)
	}
}

func (x *VMBoltTransaction) OpenTable(name string) (*VMBoltTable, error) {
	if x.tx == nil {
		return nil, VMErrorTransactionNotOpened
	}
	b := x.tx.Bucket([]byte(name))
	if b == nil {
		return nil, VMErrorTableNotExists
	}
	t := &VMBoltTable{name: name, b: b}
	return t, nil
}

func (x *VMBoltTransaction) DeleteTable(name string) error {
	if x.tx == nil {
		return VMErrorTransactionNotOpened
	}
	return x.tx.DeleteBucket([]byte(name))
}

func (x *VMBoltTransaction) BackupDBToFile(name string) error {
	if x.tx == nil {
		return VMErrorTransactionNotOpened
	}
	return x.tx.CopyFile(name, 0o644)
}

func (x *VMBoltTransaction) MethodMember(name int) (VMFunc, bool) {
	// только эти методы будут доступны из кода на языке Гонец!
	switch names.UniqueNames.GetLowerCase(name) {
	case "зафиксироватьтранзакцию":
		return VMFuncZeroParams(x.ЗафиксироватьТранзакцию), true
	case "отменитьтранзакцию":
		return VMFuncZeroParams(x.ОтменитьТранзакцию), true
	case "таблица":
		return VMFuncOneParam(x.Таблица), true
	case "удалитьтаблицу":
		return VMFuncOneParam(x.УдалитьТаблицу), true
	case "полныйбэкап":
		return VMFuncOneParam(x.ПолныйБэкап), true
	}
	return nil, false
}

func (x *VMBoltTransaction) ЗафиксироватьТранзакцию(rets *VMSlice) error {
	return x.Commit()
}

func (x *VMBoltTransaction) ОтменитьТранзакцию(rets *VMSlice) error {
	return x.Rollback()
}

func (x *VMBoltTransaction) Таблица(table VMString, rets *VMSlice) error {
	t, err := x.CreateTableIfNotExists(string(table))
	if err != nil {
		return err
	}
	rets.Append(t)
	return nil
}

func (x *VMBoltTransaction) УдалитьТаблицу(table VMString, rets *VMSlice) error {
	return x.DeleteTable(string(table))
}

func (x *VMBoltTransaction) ПолныйБэкап(table VMString, rets *VMSlice) error {
	return x.BackupDBToFile(string(table))
}

// VMBoltTable реализует функционал Bucket для BoltDB
type VMBoltTable struct {
	name string
	b    *bolt.Bucket
}

func (x *VMBoltTable) VMTypeString() string {
	return "ТаблицаФайловойБазыДанных"
}

func (x *VMBoltTable) Interface() interface{} {
	return x
}

func (x *VMBoltTable) String() string {
	return "Таблица '" + x.name + "' файловой базы данных BoltDB"
}

func (x *VMBoltTable) Set(k string, v VMBinaryTyper) error {
	i := []byte{byte(v.BinaryType())}
	ii, err := v.MarshalBinary()
	if err != nil {
		return err
	}
	return x.b.Put([]byte(k), append(i, ii...))
}

func parseBoltValue(sl []byte) (VMValue, error) {
	if len(sl) < 1 {
		return nil, VMErrorWrongDBValue
	}
	tt := sl[0]
	var bb []byte
	if len(sl) > 1 {
		bb = sl[1:]
	}
	vv, err := VMBinaryType(tt).ParseBinary(bb)
	if err != nil {
		return VMNil, err
	}
	return vv, nil
}

func (x *VMBoltTable) Get(k string) (VMValue, bool, error) {
	sl := x.b.Get([]byte(k))
	if sl == nil {
		return VMNil, false, nil
	}
	vv, err := parseBoltValue(sl)
	return vv, true, err
}

func (x *VMBoltTable) Delete(k string) error {
	return x.b.Delete([]byte(k))
}

func (x *VMBoltTable) NextId() (VMInt, error) {
	id, err := x.b.NextSequence()
	return VMInt(id), err
}

func (x *VMBoltTable) GetPrefix(pref string) (VMStringMap, error) {
	c := x.b.Cursor()
	vsm := make(VMStringMap)
	for k, v := c.Seek([]byte(pref)); k != nil && bytes.HasPrefix(k, []byte(pref)); k, v = c.Next() {
		vx, err := parseBoltValue(v)
		if err != nil {
			return vsm, err
		}
		vsm[string(k)] = vx
	}
	return vsm, nil
}

func (x *VMBoltTable) GetRange(kmin, kmax string) (VMStringMap, error) {
	c := x.b.Cursor()
	vsm := make(VMStringMap)
	for k, v := c.Seek([]byte(kmin)); k != nil && bytes.Compare(k, []byte(kmax)) <= 0; k, v = c.Next() {
		vx, err := parseBoltValue(v)
		if err != nil {
			return vsm, err
		}
		vsm[string(k)] = vx
	}
	return vsm, nil
}

func (x *VMBoltTable) GetAll() (VMStringMap, error) {
	c := x.b.Cursor()
	vsm := make(VMStringMap)
	for k, v := c.First(); k != nil; k, v = c.Next() {
		vx, err := parseBoltValue(v)
		if err != nil {
			return vsm, err
		}
		vsm[string(k)] = vx
	}
	return vsm, nil
}

func (x *VMBoltTable) SetByMap(m VMStringMap) error {
	mm := make(map[string]VMBinaryTyper)
	for ks, vs := range m {
		v, ok := vs.(VMBinaryTyper)
		if !ok {
			return VMErrorNeedBinaryTyper
		}
		mm[ks] = v
	}

	for ks, vs := range mm {

		i := []byte{byte(vs.BinaryType())}
		ii, err := vs.MarshalBinary()
		if err != nil {
			return err
		}
		err = x.b.Put([]byte(ks), append(i, ii...))
		if err != nil {
			return err
		}
	}

	return nil
}

func (x *VMBoltTable) MethodMember(name int) (VMFunc, bool) {
	// только эти методы будут доступны из кода на языке Гонец!
	switch names.UniqueNames.GetLowerCase(name) {
	case "получить":
		return VMFuncOneParam(x.Получить), true
	case "установить":
		return VMFuncTwoParams(x.Установить), true
	case "удалить":
		return VMFuncOneParam(x.Удалить), true
	case "следующийидентификатор":
		return VMFuncZeroParams(x.СледующийИдентификатор), true
	case "получитьдиапазон":
		return VMFuncTwoParams(x.ПолучитьДиапазон), true
	case "получитьпрефикс":
		return VMFuncOneParam(x.ПолучитьПрефикс), true
	case "получитьвсе":
		return VMFuncZeroParams(x.ПолучитьВсе), true
	case "установитьструктуру":
		return VMFuncOneParam(x.УстановитьСтруктуру), true
	}
	return nil, false
}

func (x *VMBoltTable) Получить(key VMString, rets *VMSlice) error {
	rv, ok, err := x.Get(string(key))
	if err != nil {
		return err
	}
	rets.Append(rv)
	rets.Append(VMBool(ok))
	return nil
}

func (x *VMBoltTable) Установить(key VMString, value VMBinaryTyper, rets *VMSlice) error {
	return x.Set(string(key), value)
}

func (x *VMBoltTable) Удалить(key VMString, rets *VMSlice) error {
	return x.Delete(string(key))
}

func (x *VMBoltTable) СледующийИдентификатор(rets *VMSlice) error {
	v, err := x.NextId()
	rets.Append(v)
	return err
}

func (x *VMBoltTable) ПолучитьДиапазон(kmin, kmax VMString, rets *VMSlice) error {
	vsm, err := x.GetRange(string(kmin), string(kmax))
	if err != nil {
		return err
	}
	rets.Append(vsm)
	return nil
}

func (x *VMBoltTable) ПолучитьПрефикс(pref VMString, rets *VMSlice) error {
	vsm, err := x.GetPrefix(string(pref))
	if err != nil {
		return err
	}
	rets.Append(vsm)
	return nil
}

func (x *VMBoltTable) ПолучитьВсе(rets *VMSlice) error {
	vsm, err := x.GetAll()
	if err != nil {
		return err
	}
	rets.Append(vsm)
	return nil
}

func (x *VMBoltTable) УстановитьСтруктуру(m VMStringMap, rets *VMSlice) error {
	return x.SetByMap(m)
}
