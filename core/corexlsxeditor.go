package core

import (
	"errors"
	"fmt"
	"github.com/xuri/excelize/v2"
	"log"
	"os"
)

type XLSXReader struct {
	VMMetaObj
	f *excelize.File
}

func (reader *XLSXReader) VMTypeString() string {
	return "ЧтениеXLSX"
}

func (reader *XLSXReader) VMRegister() {
	reader.VMRegisterConstructor(func(args VMSlice) error {
		if len(args) != 1 {
			return VMErrorNeedArgs(1)
		}

		name, ok := args[0].(VMString)
		if !ok {
			return VMErrorNeedString
		}

		var err error
		reader.f, err = excelize.OpenFile(string(name))
		if err != nil {
			fmt.Println(err)
			return VMErrorOpenXlsxFile
		}

		return nil
	})

	reader.VMRegisterMethod("ПолучитьЗначениеЯчейки", VMFuncTwoParams(reader.ПолучитьДанныеФайла))
}

func (reader *XLSXReader) ПолучитьДанныеФайла(sheet VMString, cell VMString, rets *VMSlice) error {
	value, err := reader.f.GetCellValue(string(sheet), string(cell))
	if err == nil {
		rets.Append(VMString(value))
	}
	return err
}

type XLSXWriter struct {
	VMMetaObj
	f *excelize.File
}

func (writer *XLSXWriter) VMTypeString() string {
	return "ЗаписьXLSX"
}

func (writer *XLSXWriter) VMRegister() {
	writer.VMRegisterConstructor(func(args VMSlice) error {
		if len(args) != 1 {
			return VMErrorNeedArgs(1)
		}

		name, ok := args[0].(VMString)
		if !ok {
			return VMErrorNeedString
		}

		if _, err := os.Stat(string(name)); errors.Is(err, os.ErrNotExist) {
			f, err := os.Create(string(name))
			if err != nil {
				log.Fatal(err)
			}
			f.Close()
		}

		var err error
		writer.f, err = excelize.OpenFile(string(name))
		if err != nil {
			return VMErrorOpenXlsxFile
		}

		return nil
	})

	writer.VMRegisterMethod("УстановитьЗначениеЯчейки", VMFuncThreeParams(writer.УстановитьЗначениеЯчейки))
	writer.VMRegisterMethod("ДобавитьЛист", VMFuncOneParam(writer.ДобавитьЛист))
	writer.VMRegisterMethod("ЗаписатьКнигу", VMFuncZeroParams(writer.ЗаписатьКнигу))
	writer.VMRegisterMethod("ЗаписатьКнигуКак", VMFuncOneParam(writer.ЗаписатьКнигуКак))
}

func (writer *XLSXWriter) УстановитьЗначениеЯчейки(sheet VMString, cell VMString, value VMValue, rets *VMSlice) error {
	err := writer.f.SetCellValue(string(sheet), string(cell), value)
	if err != nil {
		return VMErrorSetCellValue
	}
	return nil
}

func (writer *XLSXWriter) ДобавитьЛист(sheet VMString, rets *VMSlice) error {
	_, err := writer.f.NewSheet(string(sheet))
	if err != nil {
		return VMErrorAddSheet
	}
	return err
}

func (writer *XLSXWriter) ЗаписатьКнигу(rets *VMSlice) error {
	if err := writer.f.Save(); err != nil {
		return VMErrorSaveXlsxFile
	}
	return nil
}

func (writer *XLSXWriter) ЗаписатьКнигуКак(fileName VMString, rets *VMSlice) error {
	if err := writer.f.SaveAs(string(fileName)); err != nil {
		return VMErrorSaveXlsxFile
	}
	return nil
}
