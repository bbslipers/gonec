package core

import (
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/ean"
	"image/png"
	"os"
	"path/filepath"
	"unicode"
)

type Ean13 struct {
	VMMetaObj
	text   string
	width  int
	height int
	path   string
}

func (f *Ean13) VMTypeString() string {
	return "EAN13Код"
}

func isDigitsOnly(s string) bool {
	for _, char := range s {
		if !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}

func (f *Ean13) VMRegister() {
	f.VMRegisterConstructor(func(args VMSlice) error {
		if len(args) != 4 {
			return VMErrorNeedArgs(4)
		}

		text, ok := args[0].(VMString)
		if !ok {
			return VMErrorNeedString
		}

		if len(text) != 12 || !isDigitsOnly(string(text)) {
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

		eanCode, _ := ean.Encode(f.text)
		barcode, err := barcode.Scale(eanCode, f.width, f.height)
		file, _ := os.Create(f.path)
		defer file.Close()
		png.Encode(file, barcode)

		return nil
	})

	f.VMRegisterMethod("ПолучитьКод", VMFuncZeroParams(f.ПолучитьКод))
	f.VMRegisterMethod("ПолучитьПуть", VMFuncZeroParams(f.ПолучитьПуть))
	f.VMRegisterMethod("ПолучитьВысоту", VMFuncZeroParams(f.ПолучитьВысоту))
	f.VMRegisterMethod("ПолучитьШирину", VMFuncZeroParams(f.ПолучитьШирину))
}

func (f *Ean13) ПолучитьКод(rets *VMSlice) error {
	rets.Append(VMString(f.text))
	return nil
}

func (f *Ean13) ПолучитьПуть(rets *VMSlice) error {
	rets.Append(VMString(f.path))
	return nil
}

func (f *Ean13) ПолучитьВысоту(rets *VMSlice) error {
	rets.Append(VMInt(f.height))
	return nil
}

func (f *Ean13) ПолучитьШирину(rets *VMSlice) error {
	rets.Append(VMInt(f.width))
	return nil
}
