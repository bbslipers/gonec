package core

import (
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/datamatrix"
	"image/png"
	"os"
	"path/filepath"
)

type DataMatrix struct {
	VMMetaObj
	text string
	size int
	path string
}

func (f *DataMatrix) VMTypeString() string {
	return "DMКод"
}

func (f *DataMatrix) VMRegister() {
	f.VMRegisterConstructor(func(args VMSlice) error {
		if len(args) != 3 {
			return VMErrorNeedArgs(3)
		}

		text, ok := args[0].(VMString)
		if !ok {
			return VMErrorNeedString
		}

		size, ok := args[1].(VMInt)
		if !ok {
			return VMErrorNeedInt
		}

		path, ok := args[2].(VMString)
		if !ok {
			return VMErrorNeedString
		}

		f.text = string(text)
		f.size = int(size)
		abspath, err := filepath.Abs(string(path))
		if err != nil {
			return VMErrorIncorrectFieldType
		}

		f.path = abspath

		dmCode, _ := datamatrix.Encode(f.text)
		dmCode, _ = barcode.Scale(dmCode, f.size, f.size)
		file, _ := os.Create(f.path)
		defer file.Close()
		png.Encode(file, dmCode)

		return nil
	})

	f.VMRegisterMethod("ПолучитьКод", VMFuncZeroParams(f.ПолучитьКод))
	f.VMRegisterMethod("ПолучитьПуть", VMFuncZeroParams(f.ПолучитьПуть))
	f.VMRegisterMethod("ПолучитьВысоту", VMFuncZeroParams(f.ПолучитьВысоту))
	f.VMRegisterMethod("ПолучитьШирину", VMFuncZeroParams(f.ПолучитьШирину))
}

func (f *DataMatrix) ПолучитьКод(rets *VMSlice) error {
	rets.Append(VMString(f.text))
	return nil
}

func (f *DataMatrix) ПолучитьПуть(rets *VMSlice) error {
	rets.Append(VMString(f.path))
	return nil
}

func (f *DataMatrix) ПолучитьВысоту(rets *VMSlice) error {
	rets.Append(VMInt(f.size))
	return nil
}

func (f *DataMatrix) ПолучитьШирину(rets *VMSlice) error {
	rets.Append(VMInt(f.size))
	return nil
}
