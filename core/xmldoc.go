package core

import (
	"github.com/beevik/etree"
)

type XMLDoc struct {
	VMMetaObj
	doc *etree.Document
}

func (f *XMLDoc) VMTypeString() string {
	return "ЗаписьXML"
}

func (f *XMLDoc) VMRegister() {
	f.VMRegisterConstructor(func(args VMSlice) error {
		return nil
	})

	f.doc = etree.NewDocument()

	f.VMRegisterMethod("ЗаписатьОбъявленияXML", VMFuncOneParam(f.ЗаписатьОбъявленияXML))
	f.VMRegisterMethod("ЗаписатьКомментарийXML", VMFuncOneParam(f.ЗаписатьКомментарийXML))
	f.VMRegisterMethod("ФорматироватьXML", VMFuncOneParam(f.ФорматироватьXML))
	f.VMRegisterMethod("ЗаписатьСтрокуXML", VMFuncZeroParams(f.ЗаписатьСтрокуXML))
	f.VMRegisterMethod("ЗаписатьФайлXML", VMFuncOneParam(f.ЗаписатьФайлXML))
}

func (f *XMLDoc) ЗаписатьОбъявленияXML(params VMStringMap, rets *VMSlice) error {
	for key, value := range params {
		f.doc.CreateProcInst(key, string(value.(VMString)))
	}
	return nil
}

func (f *XMLDoc) ЗаписатьКомментарийXML(params VMString, rets *VMSlice) error {
	f.doc.CreateComment(string(params))
	return nil
}

func (f *XMLDoc) ФорматироватьXML(params VMInt, rets *VMSlice) error {
	f.doc.Indent(int(params))
	return nil
}

func (f *XMLDoc) ЗаписатьСтрокуXML(rets *VMSlice) error {
	str, err := f.doc.WriteToString()
	if err != nil {
		return VMErrorSaveXML
	}
	rets.Append(VMString(str))
	return nil
}

func (f *XMLDoc) ЗаписатьФайлXML(filePath VMString, rets *VMSlice) error {
	err := f.doc.WriteToFile(string(filePath))
	if err != nil {
		return VMErrorSaveXML
	}
	return nil
}
