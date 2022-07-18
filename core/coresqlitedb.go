package core

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/shinanca/gonec/names"
	_ "modernc.org/sqlite"
)

// VMSqliteDB - база данных Sqlite
type VMSqliteDB struct {
	VMMetaObj
	*sql.DB
}

func (x *VMSqliteDB) VMTypeString() string { return "SqliteБазаДанных" }

func (d *VMSqliteDB) VMRegister() {
	d.VMRegisterConstructor(func(args VMSlice) error {
		if len(args) != 1 {
			return VMErrorNeedArgs(1)
		}

		file, ok := args[0].(VMString)
		if !ok {
			return errors.New("Первым аргументом требуется путь/имя базы sqlite3")
		}

		db, err := sql.Open("sqlite", string(file))
		if err != nil {
			return err
		}

		d.DB = db
		return nil
	})

	d.VMRegisterMethod("Выполнить", VMFuncOneParamOptionals(-1, d.Выполнить))
	d.VMRegisterMethod("Выбрать", VMFuncOneParamOptionals(-1, d.Выбрать))
	d.VMRegisterMethod("Закрыть", VMFuncZeroParams(d.Закрыть))
}

func (d *VMSqliteDB) convertParam(v VMValue) (any, error) {
	e, ok := v.(VMInterfacer)
	if !ok {
		return nil, errors.New("Параметрами запроса должны быть элементы базовых типов")
	}

	return e.Interface(), nil
}

func (d *VMSqliteDB) convertParams(params VMSlice) ([]any, error) {
	var args []any
	for _, arg := range params {
		if m, ok := arg.(VMStringMap); ok {
			for k, v := range m {
				e, err := d.convertParam(v)
				if err != nil {
					return nil, err
				}

				args = append(args, sql.Named(k, e))
			}
		} else {
			e, err := d.convertParam(arg)
			if err != nil {
				return nil, err
			}
			args = append(args, e)
		}
	}
	return args, nil
}

func (d *VMSqliteDB) Выполнить(query VMString, rest VMSlice, rets *VMSlice) error {
	args, err := d.convertParams(rest)
	if err != nil {
		return err
	}

	_, err = d.Exec(string(query), args...)
	return err
}

type VMSqliteQueryResult struct {
	*sql.Rows
	colTypes []reflect.Type
	colNames []string
}

func (r *VMSqliteQueryResult) VMTypeString() string {
	return "РезультатыSqliteЗапроса"
}

func (r *VMSqliteQueryResult) MethodMember(name int) (VMFunc, bool) {
	switch names.UniqueNames.GetLowerCase(name) {
	case "следующий":
		return VMFuncZeroParams(r.Следующий), true
	case "получить":
		return VMFuncZeroParams(r.Получить), true
	case "выгрузить":
		return VMFuncZeroParams(r.Выгрузить), true
	case "закрыть":
		return VMFuncZeroParams(r.Закрыть), true
	}
	return nil, false
}

// prepareColumnInfo парсит типы и названия столбцов и кеширует их
func (r *VMSqliteQueryResult) prepareColumnInfo() error {
	if r.colTypes == nil {
		ct, err := r.ColumnTypes()
		if err != nil {
			return err
		}

		var types []reflect.Type
		for _, t := range ct {
			dbtype := t.DatabaseTypeName()
			switch dbtype {
			case "DATE":
				fallthrough
			case "DATETIME":
				fallthrough
			case "TIMESTAMP":
				types = append(types, reflect.TypeOf(time.Time{}))
			default:
				types = append(types, t.ScanType())
			}
		}
		r.colTypes = types
	}

	if r.colNames == nil {
		names, err := r.Columns()
		if err != nil {
			return err
		}

		r.colNames = names
	}
	return nil
}

// prepareScanRow возвращает строку с правильными типами для сканирования
func (r *VMSqliteQueryResult) prepareScanRow() ([]any, error) {
	if err := r.prepareColumnInfo(); err != nil {
		return nil, err
	}

	// Типы те же, что и в modernc.org/sqlite/sqlite.go:ColumnTypeScanType
	var scanRow []any
	for _, rt := range r.colTypes {
		switch rt.Kind() {
		case reflect.Bool:
			fallthrough
		case reflect.Int64:
			fallthrough
		case reflect.Float64:
			fallthrough
		case reflect.String:
			scanRow = append(scanRow, reflect.New(rt).Interface())
		default:
			if rt == reflect.TypeOf(time.Time{}) {
				scanRow = append(scanRow, reflect.New(rt).Interface())
			} else if rt == reflect.TypeOf(nil) {
				var x any
				scanRow = append(scanRow, &x)
			} else {
				return nil, fmt.Errorf("Неизвестный тип столбца результатов: %s", rt)
			}
		}
	}
	return scanRow, nil
}

func (r *VMSqliteQueryResult) scanRow() (VMStringMap, error) {
	row, err := r.prepareScanRow()
	if err != nil {
		return nil, err
	}

	if err = r.Scan(row...); err != nil {
		return nil, err
	}

	results := VMStringMap{}
	for i, v := range row {
		name := r.colNames[i]
		switch val := v.(type) {
		case *bool:
			results[name] = VMBool(*val)
		case *int64:
			results[name] = VMInt(*val)
		case *float64:
			var d VMDecNum
			d.ParseGoType(v)
			results[name] = d
		case *string:
			results[name] = VMString(*val)
		case *time.Time:
			results[name] = VMTime(*val)
		case *interface{}:
			if *val == nil {
				results[name] = VMNil
			} else {
				return nil, fmt.Errorf("Значение неизвестного типа считанного столбца результатов: %+v", v)
			}
		}
	}
	return results, nil
}

func (r *VMSqliteQueryResult) Следующий(rets *VMSlice) error {
	rets.Append(VMBool(r.Next()))
	return nil
}

func (r *VMSqliteQueryResult) Получить(rets *VMSlice) error {
	vals, err := r.scanRow()
	if err != nil {
		return err
	}

	rets.Append(vals)
	return nil
}

func (r *VMSqliteQueryResult) Выгрузить(rets *VMSlice) error {
	var results VMSlice
	for r.Next() {
		vals, err := r.scanRow()
		if err != nil {
			return err
		}
		results.Append(vals)
	}

	rets.Append(results)
	// возвращаем Err после неявного закрытия
	return r.Err()
}

func (r *VMSqliteQueryResult) Закрыть(rets *VMSlice) error {
	return r.Close()
}

func (d *VMSqliteDB) Выбрать(query VMString, rest VMSlice, rets *VMSlice) error {
	args, err := d.convertParams(rest)
	if err != nil {
		return err
	}

	rows, err := d.Query(string(query), args...)
	if err != nil {
		return err
	}

	rets.Append(&VMSqliteQueryResult{Rows: rows})
	return nil
}

func (d *VMSqliteDB) Закрыть(rets *VMSlice) error {
	return d.Close()
}
