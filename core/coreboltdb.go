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
		return VMFuncOneParam[VMString](x.Открыть), true
	case "закрыть":
		return VMFuncZeroParams(x.Закрыть), true
	case "начатьтранзакцию":
		return VMFuncOneParam[VMBool](x.НачатьТранзакцию), true
	}
	return nil, false
}

func (x *VMBoltDB) Открыть(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	return x.Open(string(args[0].(VMString)))
}

func (x *VMBoltDB) НачатьТранзакцию(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	tr, err := x.Begin(args[0].(VMBool).Bool())
	if err != nil {
		return err
	}
	rets.Append(tr)
	return nil
}

func (x *VMBoltDB) Закрыть(args VMSlice, rets *VMSlice, envout *(*Env)) error {
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
		return VMFuncOneParam[VMString](x.Таблица), true
	case "удалитьтаблицу":
		return VMFuncOneParam[VMString](x.УдалитьТаблицу), true
	case "полныйбэкап":
		return VMFuncOneParam[VMString](x.ПолныйБэкап), true
	}
	return nil, false
}

func (x *VMBoltTransaction) ЗафиксироватьТранзакцию(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	return x.Commit()
}

func (x *VMBoltTransaction) ОтменитьТранзакцию(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	return x.Rollback()
}

func (x *VMBoltTransaction) Таблица(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	t, err := x.CreateTableIfNotExists(string(args[0].(VMString)))
	if err != nil {
		return err
	}
	rets.Append(t)
	return nil
}

func (x *VMBoltTransaction) УдалитьТаблицу(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	return x.DeleteTable(string(args[0].(VMString)))
}

func (x *VMBoltTransaction) ПолныйБэкап(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	return x.BackupDBToFile(string(args[0].(VMString)))
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
		return VMFuncOneParam[VMString](x.Получить), true
	case "установить":
		return VMFuncTwoParams[VMString, VMBinaryTyper](x.Установить), true
	case "удалить":
		return VMFuncOneParam[VMString](x.Удалить), true
	case "следующийидентификатор":
		return VMFuncZeroParams(x.СледующийИдентификатор), true
	case "получитьдиапазон":
		return VMFuncTwoParams[VMString, VMString](x.ПолучитьДиапазон), true
	case "получитьпрефикс":
		return VMFuncOneParam[VMString](x.ПолучитьПрефикс), true
	case "получитьвсе":
		return VMFuncZeroParams(x.ПолучитьВсе), true
	case "установитьструктуру":
		return VMFuncOneParam[VMStringMap](x.УстановитьСтруктуру), true
	}
	return nil, false
}

func (x *VMBoltTable) Получить(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	rv, ok, err := x.Get(string(args[0].(VMString)))
	if err != nil {
		return err
	}
	rets.Append(rv)
	rets.Append(VMBool(ok))
	return nil
}

func (x *VMBoltTable) Установить(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	return x.Set(string(args[0].(VMString)), args[1].(VMBinaryTyper))
}

func (x *VMBoltTable) Удалить(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	return x.Delete(string(args[0].(VMString)))
}

func (x *VMBoltTable) СледующийИдентификатор(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	v, err := x.NextId()
	rets.Append(v)
	return err
}

func (x *VMBoltTable) ПолучитьДиапазон(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	vsm, err := x.GetRange(string(args[0].(VMString)), string(args[1].(VMString)))
	if err != nil {
		return err
	}
	rets.Append(vsm)
	return nil
}

func (x *VMBoltTable) ПолучитьПрефикс(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	vsm, err := x.GetPrefix(string(args[0].(VMString)))
	if err != nil {
		return err
	}
	rets.Append(vsm)
	return nil
}

func (x *VMBoltTable) ПолучитьВсе(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	vsm, err := x.GetAll()
	if err != nil {
		return err
	}
	rets.Append(vsm)
	return nil
}

func (x *VMBoltTable) УстановитьСтруктуру(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	return x.SetByMap(args[0].(VMStringMap))
}
