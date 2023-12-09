package core

import (
	"github.com/beevik/etree"
)

type XMLElem struct {
	VMMetaObj
	elem *etree.Element
}

func (f *XMLElem) VMTypeString() string {
	return "ЭлементXML"
}

func (f *XMLElem) VMRegister() {
	f.VMRegisterConstructor(func(args VMSlice) error {
		if len(args) != 2 {
			return VMErrorNeedArgs(2)
		}

		name, ok := args[0].(VMString)
		if !ok {
			return VMErrorNeedString
		}

		//TODO: inheritance?
		var parent *etree.Element
		xmlElem, ok := args[1].(*XMLElem)
		if ok {
			parent = xmlElem.elem
		} else {
			xmlDoc, ok := args[1].(*XMLDoc)
			parent = &xmlDoc.doc.Element
			if !ok {
				return VMErrorIncorrectFieldType
			}
		}

		f.elem = parent.CreateElement(string(name))
		return nil
	})

	f.VMRegisterMethod("ЗаписатьТекстXML", VMFuncOneParam(f.ЗаписатьТекстXML))
	f.VMRegisterMethod("ЗаписатьАтрибутыXML", VMFuncOneParam(f.ЗаписатьАтрибутыXML))
}

func (f *XMLElem) ЗаписатьТекстXML(params VMString, rets *VMSlice) error {
	f.elem.SetText(string(params))
	return nil
}

func (f *XMLElem) ЗаписатьАтрибутыXML(params VMStringMap, rets *VMSlice) error {
	for key, value := range params {
		f.elem.CreateAttr(key, string(value.(VMString)))
	}
	return nil
}
