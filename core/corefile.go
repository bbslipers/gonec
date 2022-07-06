package core

import (
	"fmt"
	"io/fs"
	"os"
	"time"
)

type File struct {
	VMMetaObj
	name string
}

func (f *File) VMTypeString() string {
	return "Файл"
}

func (f *File) wrapError(err error) error {
	if os.IsNotExist(err) {
		return fmt.Errorf("Файла '%s' не существует", f.name)
	} else if os.IsPermission(err) {
		return fmt.Errorf("Доступ к файлу '%s' запрещён", f.name)
	} else {
		return fmt.Errorf("Возникла неизвестная ошибка: %s", err.Error())
	}
}

func (f *File) stat(shouldExist bool) (fs.FileInfo, error) {
	if s, err := os.Stat(f.name); err == nil {
		return s, nil
	} else if !shouldExist && os.IsNotExist(err) {
		return nil, nil
	} else {
		return nil, f.wrapError(err)
	}
}

func (f *File) Exists() (bool, error) {
	if s, err := f.stat(false); err == nil {
		return s != nil, nil
	} else {
		return false, err
	}
}

func (f *File) Size() (int64, error) {
	if s, err := f.stat(true); err == nil {
		return s.Size(), nil
	} else {
		return 0, err
	}
}

func (f *File) IsReadOnly() (bool, error) {
	// Кроссплатформенный способ проверить можно ли писать в файл - открыть его
	file, err := os.OpenFile(f.name, os.O_WRONLY, 0o666)
	if err != nil {
		if os.IsPermission(err) {
			return true, nil
		}
		return false, f.wrapError(err)
	}
	file.Close()
	return false, nil
}

func (f *File) SetReadOnly(readonly bool) error {
	s, err := f.stat(true)
	if err == nil {
		// Очищаем биты записи
		if err = os.Chmod(f.name, s.Mode()&^0o222); err == nil {
			return nil
		}
	}
	return f.wrapError(err)
}

func (f *File) ModificationTime() (time.Time, error) {
	if s, err := f.stat(true); err == nil {
		return s.ModTime(), nil
	} else {
		return time.Time{}, err
	}
}

func (f *File) SetModificationTime(t time.Time) error {
	if err := os.Chtimes(f.name, t, t); err != nil {
		return f.wrapError(err)
	}
	return nil
}

func (f *File) VMRegister() {
	f.VMRegisterConstructor(func(args VMSlice) error {
		if len(args) != 1 {
			return VMErrorNeedArgs(1)
		}

		name, ok := args[0].(VMString)
		if !ok {
			return VMErrorNeedString
		}

		f.name = string(name)
		return nil
	})

	f.VMRegisterMethod("Существует", VMFuncZeroParams(f.Существует))
	f.VMRegisterMethod("Размер", VMFuncZeroParams(f.Размер))
	f.VMRegisterMethod("ПолучитьТолькоЧтение", VMFuncZeroParams(f.ПолучитьТолькоЧтение))
	f.VMRegisterMethod("ПолучитьВремяИзменения", VMFuncOneParam[VMBool](f.ПолучитьВремяИзменения))
	f.VMRegisterMethod("УстановитьТолькоЧтение", VMFuncOneParam[VMBool](f.УстановитьТолькоЧтение))
	f.VMRegisterMethod("УстановитьВремяИзменения", VMFuncOneParam[VMDateTimer](
		f.УстановитьВремяИзменения))
}

func (f *File) Существует(args VMSlice, rets *VMSlice) error {
	exists, err := f.Exists()
	if err == nil {
		rets.Append(VMBool(exists))
	}
	return err
}

func (f *File) Размер(args VMSlice, rets *VMSlice) error {
	size, err := f.Size()
	if err == nil {
		rets.Append(VMInt(size))
	}
	return err
}

func (f *File) ПолучитьТолькоЧтение(args VMSlice, rets *VMSlice) error {
	readonly, err := f.IsReadOnly()
	if err == nil {
		rets.Append(VMBool(readonly))
	}
	return err
}

func (f *File) ПолучитьВремяИзменения(args VMSlice, rets *VMSlice) error {
	modtime, err := f.ModificationTime()
	if err == nil {
		if args[0].(VMBool) {
			rets.Append(VMString(VMTime(modtime).String()))
		} else {
			rets.Append(VMTime(modtime))
		}
	}
	return err
}

func (f *File) УстановитьТолькоЧтение(args VMSlice, rets *VMSlice) error {
	return f.SetReadOnly(args[0].(VMBool).Bool())
}

func (f *File) УстановитьВремяИзменения(args VMSlice, rets *VMSlice) error {
	return f.SetModificationTime(args[0].(VMDateTimer).Time().GolangTime())
}
