package core

import (
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/twooffive"
	"image/png"
	"os"
	"path/filepath"
)

type I2of5 struct {
	VMMetaObj
	text   string
	width  int
	height int
	path   string
}

func (f *I2of5) VMTypeString() string {
	return "I2OF5Код"
}

func (f *I2of5) VMRegister() {
	f.VMRegisterConstructor(func(args VMSlice) error {
		if len(args) != 4 {
			return VMErrorNeedArgs(4)
		}

		text, ok := args[0].(VMString)
		if !ok {
			return VMErrorNeedString
		}

		if len(text)%2 != 0 || !isDigitsOnly(string(text)) {
			return VMErrorEan13Format
		}

		width, ok := args[1].(VMInt)
		if !ok {
			return VMErrorNeedInt
		}

		height, ok := args[2].(VMInt)
		if !ok {
			return VMErrorNeedInt
		}

		path, ok := args[3].(VMString)
		if !ok {
			return VMErrorNeedString
		}

		f.text = string(text)
		f.width = int(width)
		f.height = int(height)
		abspath, err := filepath.Abs(string(path))
		if err != nil {
			return VMErrorIncorrectFieldType
		}

		f.path = abspath

		i2of5Code, err := twooffive.Encode(f.text, true)
		if err != nil {
			return VMErrorIncorrectOperation
		}
		i2of5Code, err = barcode.Scale(i2of5Code, f.width, f.height)
		if err != nil {
			return VMErrorIncorrectOperation
		}
		file, err := os.Create(f.path)
		if err != nil {
			return VMErrorIncorrectOperation
		}
		defer file.Close()
		png.Encode(file, i2of5Code)

		return nil
	})

	f.VMRegisterMethod("ПолучитьКод", VMFuncZeroParams(f.ПолучитьКод))
	f.VMRegisterMethod("ПолучитьПуть", VMFuncZeroParams(f.ПолучитьПуть))
	f.VMRegisterMethod("ПолучитьВысоту", VMFuncZeroParams(f.ПолучитьВысоту))
	f.VMRegisterMethod("ПолучитьШирину", VMFuncZeroParams(f.ПолучитьШирину))
}

func (f *I2of5) ПолучитьКод(rets *VMSlice) error {
	rets.Append(VMString(f.text))
	return nil
}

func (f *I2of5) ПолучитьПуть(rets *VMSlice) error {
	rets.Append(VMString(f.path))
	return nil
}

func (f *I2of5) ПолучитьВысоту(rets *VMSlice) error {
	rets.Append(VMInt(f.height))
	return nil
}

func (f *I2of5) ПолучитьШирину(rets *VMSlice) error {
	rets.Append(VMInt(f.width))
	return nil
}
