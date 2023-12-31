package core

import (
	"fmt"
	"github.com/jordan-wright/email"
	"net/smtp"
	"os"
	"path/filepath"
	"time"
)

type EmailProfile struct {
	VMMetaObj
	smtp.Auth
	SmtpServer string
}

func (f *EmailProfile) VMTypeString() string {
	return "ПочтовыйПрофиль"
}

func (f *EmailProfile) VMRegister() {
	f.VMRegisterConstructor(func(args VMSlice) error {
		if len(args) != 4 {
			return VMErrorNeedArgs(4)
		}

		smtpHost, ok := args[0].(VMString)
		if !ok {
			return VMErrorNeedString
		}

		smtpPort, ok := args[1].(VMInt)
		if !ok {
			return VMErrorNeedInt
		}

		UserName, ok := args[2].(VMString)
		if !ok {
			return VMErrorNeedString
		}

		Password, ok := args[3].(VMString)
		if !ok {
			return VMErrorNeedString
		}

		f.SmtpServer = fmt.Sprintf("%s:%d", smtpHost, smtpPort)
		f.Auth = smtp.PlainAuth("", string(UserName), string(Password), string(smtpHost))

		return nil
	})

	f.VMRegisterMethod("Отправить", VMFuncOneParam(f.Отправить))
}

// Определение Content-Type для вложения
func getContentType(fileName string) string {
	ext := filepath.Ext(fileName)
	switch ext {
	case ".txt":
		return "text/plain"
	case ".png":
		return "image/png"
	default:
		return "application/octet-stream"
	}
}

func (f *EmailProfile) Отправить(emailInfo *EmailData, rets *VMSlice) error {
	// Создание нового объекта email
	e := email.NewEmail()

	// Установка параметров письма
	e.To = []string{emailInfo.RecipientAdr}
	e.From = emailInfo.SenderAdr
	e.Subject = emailInfo.Subject

	// Добавление текстовой части
	e.Text = []byte(emailInfo.Text)

	// Добавление вложений
	for _, filePath := range emailInfo.Attachments {
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}

		// Определение Content-Type для вложения
		contentType := getContentType(filePath)

		// Добавление вложения в email
		_, err = e.Attach(file, filepath.Base(filePath), contentType)
		if err != nil {
			return err
		}

		file.Close()
	}

	// Настройка таймаута
	timeout := 10 * time.Second // установите таймаут в секундах

	// Создание канала для управления таймаутом
	done := make(chan error)

	// Горутина для выполнения отправки почты
	go func() {
		// Отправка сообщения
		err := e.Send(f.SmtpServer, f.Auth)
		if err != nil {
			done <- err
		} else {
			done <- nil
		}
	}()

	// Ожидание либо завершения горутины, либо таймаута
	select {
	case err := <-done:
		if err != nil {
			fmt.Println("Ошибка при отправке почты")
		} else {
			fmt.Println("Почта отправлена успешно!")
		}
	case <-time.After(timeout):
		fmt.Println("Превышено время ожидания ответа: проверьте антивирус, возможно он блокирует исходящую почту")
	}

	return nil
}
