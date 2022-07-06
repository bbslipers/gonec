package core

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

const (
	CRLF = "\r\n"
	LF   = "\n"
)

type TextDocument struct {
	VMMetaObj

	lines     []string
	separator string
}

func (t *TextDocument) VMTypeString() string {
	return "ТекстовыйДокумент"
}

func (t *TextDocument) String() string {
	return strings.Join(t.lines, t.separator)
}

func (t *TextDocument) fromText(text string) error {
	if !utf8.ValidString(text) {
		return errors.New("Текст не является корректным UTF-8 документом")
	}

	// Пытаемся сохранить вместимость
	if t.lines == nil {
		t.lines = make([]string, 0)
	} else {
		t.lines = t.lines[:0]
	}

	s := bufio.NewScanner(strings.NewReader(text))
	for s.Scan() {
		t.lines = append(t.lines, s.Text())
	}

	if err := s.Err(); err != nil {
		t.lines = t.lines[:0]
		return errors.New("Не удалось считать текст: " + err.Error())
	}

	// Угадываем окончания строк
	if strings.Contains(text, CRLF) {
		t.separator = CRLF
	} else {
		t.separator = LF
	}
	return nil
}

func (t *TextDocument) Read(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Файла '%s' не существует", filename)
		} else if os.IsPermission(err) {
			return fmt.Errorf("Доступ на чтение файла '%s' запрещён", filename)
		}
		return fmt.Errorf("Не удалось прочитать файл '%s': %s", filename, err.Error())
	}
	return t.fromText(string(content))
}

func (t *TextDocument) Write(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("Нет прав на запись файла '%s'", filename)
		}
		return fmt.Errorf("Не удалось открыть файл '%s' на запись: %s", filename, err.Error())
	}

	_, err = file.WriteString(t.String())
	if err != nil {
		return fmt.Errorf("Не удалось записать файл '%s': %s", filename, err.Error())
	}
	return nil
}

func (t *TextDocument) VMRegister() {
	t.VMRegisterMethod("Прочитать", VMFuncOneParam[VMString](t.Прочитать))
	t.VMRegisterMethod("Записать", VMFuncOneParam[VMString](t.Записать))
	t.VMRegisterMethod("ПолучитьТекст", VMFuncZeroParams(t.ПолучитьТекст))
	t.VMRegisterMethod("УстановитьТекст", VMFuncOneParam[VMString](t.УстановитьТекст))
	t.VMRegisterMethod("КоличествоСтрок", VMFuncZeroParams(t.КоличествоСтрок))
	t.VMRegisterMethod("ДобавитьСтроку", VMFuncOneParam[VMString](t.ДобавитьСтроку))
	t.VMRegisterMethod("УдалитьСтроку", VMFuncOneParam[VMInt](t.УдалитьСтроку))
	t.VMRegisterMethod("ЗаменитьСтроку", VMFuncTwoParams[VMInt, VMString](t.ЗаменитьСтроку))
}

func (t *TextDocument) Прочитать(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	return t.Read(string(args[0].(VMString)))
}

func (t *TextDocument) Записать(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	return t.Write(string(args[0].(VMString)))
}

func (t *TextDocument) ПолучитьТекст(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	rets.Append(VMString(t.String()))
	return nil
}

func (t *TextDocument) УстановитьТекст(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	return t.fromText(string(args[0].(VMString)))
}

func (t *TextDocument) КоличествоСтрок(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	rets.Append(VMInt(len(t.lines)))
	return nil
}

func (t *TextDocument) isSingleLine(line string) bool {
	lineEndings := strings.Count(line, t.separator)
	if lineEndings > 1 || (lineEndings == 1 && !strings.HasSuffix(line, t.separator)) {
		return false
	}
	return true
}

// O(1)
func (t *TextDocument) ДобавитьСтроку(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	line := string(args[0].(VMString))

	// Добавляем строку только если она действительно является ОДНОЙ строкой
	if !t.isSingleLine(line) {
		return errors.New("Добавляемый текст должен состоять ровно из одной строки")
	}

	t.lines = append(t.lines, line)
	return nil
}

func (t *TextDocument) parseLineNumberArg(args VMSlice) (int, error) {
	n := int(args[0].(VMInt))
	if n < 1 || n > len(t.lines) {
		return 0, fmt.Errorf("Строки с номером %d не существует", n)
	}
	return n - 1, nil
}

// O(n)
func (t *TextDocument) УдалитьСтроку(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	n, err := t.parseLineNumberArg(args)
	if err != nil {
		return err
	}

	t.lines = append(t.lines[:n], t.lines[n+1:]...)
	return nil
}

func (t *TextDocument) ЗаменитьСтроку(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	n, err := t.parseLineNumberArg(args)
	if err != nil {
		return err
	}

	line := string(args[1].(VMString))
	// Изменяем только если новая строка действительно является ОДНОЙ строкой
	if !t.isSingleLine(line) {
		return errors.New("Новая строка должна состоять ровно из одной строки")
	}

	t.lines[n] = string(line)
	return nil
}
