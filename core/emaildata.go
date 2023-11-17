package core

import (
	"encoding/json"
	"fmt"
)

type EmailInfo struct {
	RecipientAdr string   `json:"recipient_adr"`
	SenderAdr    string   `json:"sender_adr"`
	SenderName   string   `json:"sender_name"`
	Subject      string   `json:"subject"`
	Text         string   `json:"text"`
	Attachments  []string `json:"attachments"`
}

type EmailData struct {
	VMMetaObj
	EmailInfo
}

func (f *EmailData) VMTypeString() string {
	return "ПочтовоеОтправление"
}

func (f *EmailData) String() string {
	jsonData, err := json.Marshal(f.EmailInfo)
	if err != nil {
		fmt.Println("Error marshalling:", err)
		return ""
	}
	return string(jsonData)
}

func (f *EmailData) VMRegister() {
	f.VMRegisterConstructor(func(args VMSlice) error {
		if len(args) != 5 {
			return VMErrorNeedArgs(4)
		}

		RecipientAdr, ok := args[0].(VMString)
		if !ok {
			return VMErrorNeedString
		}

		SenderAdr, ok := args[1].(VMString)
		if !ok {
			return VMErrorNeedString
		}

		SenderName, ok := args[2].(VMString)
		if !ok {
			return VMErrorNeedString
		}

		Subject, ok := args[3].(VMString)
		if !ok {
			return VMErrorNeedString
		}

		Text, ok := args[4].(VMString)
		if !ok {
			return VMErrorNeedString
		}

		f.EmailInfo = EmailInfo{RecipientAdr: string(RecipientAdr), SenderAdr: string(SenderAdr), SenderName: string(SenderName), Subject: string(Subject), Text: string(Text)}

		return nil
	})

	f.VMRegisterMethod("ПрикрепитьВложение", VMFuncOneParam(f.ПрикрепитьВложение))
}

func (f *EmailData) ПрикрепитьВложение(file VMString, rets *VMSlice) error {
	f.Attachments = append(f.Attachments, string(file))
	return nil
}
