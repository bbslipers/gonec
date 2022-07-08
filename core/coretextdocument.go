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
		return errors.New("Текст не является корректной UTF-8 строкой")
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

	// Убираем последнюю пустую строку если документ заканчивается на окончание строки
	if strings.HasSuffix(text, t.separator) && len(t.lines) > 0 && t.lines[len(t.lines)-1] == "" {
		t.lines = t.lines[:len(t.lines)-1]
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
	t.VMRegisterMethod("Прочитать", VMFuncOneParam(t.Прочитать))
	t.VMRegisterMethod("Записать", VMFuncOneParam(t.Записать))
	t.VMRegisterMethod("ПолучитьТекст", VMFuncZeroParams(t.ПолучитьТекст))
	t.VMRegisterMethod("УстановитьТекст", VMFuncOneParam(t.УстановитьТекст))
	t.VMRegisterMethod("КоличествоСтрок", VMFuncZeroParams(t.КоличествоСтрок))
	t.VMRegisterMethod("ДобавитьСтроку", VMFuncOneParam(t.ДобавитьСтроку))
	t.VMRegisterMethod("УдалитьСтроку", VMFuncOneParam(t.УдалитьСтроку))
	t.VMRegisterMethod("ЗаменитьСтроку", VMFuncTwoParams(t.ЗаменитьСтроку))
}

func (t *TextDocument) Прочитать(name VMString, rets *VMSlice) error {
	return t.Read(string(name))
}

func (t *TextDocument) Записать(name VMString, rets *VMSlice) error {
	return t.Write(string(name))
}

func (t *TextDocument) ПолучитьТекст(rets *VMSlice) error {
	rets.Append(VMString(t.String()))
	return nil
}

func (t *TextDocument) УстановитьТекст(text VMString, rets *VMSlice) error {
	return t.fromText(string(text))
}

func (t *TextDocument) КоличествоСтрок(rets *VMSlice) error {
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
func (t *TextDocument) ДобавитьСтроку(vmline VMString, rets *VMSlice) error {
	line := string(vmline)

	// Добавляем строку только если она действительно является ОДНОЙ строкой
	if !t.isSingleLine(line) {
		return errors.New("Добавляемый текст должен состоять ровно из одной строки")
	}

	t.lines = append(t.lines, line)
	return nil
}

func (t *TextDocument) validateLineNumber(n int) (int, error) {
	if n < 1 || n > len(t.lines) {
		return 0, fmt.Errorf("Строки с номером %d не существует", n)
	}
	return n - 1, nil
}

// O(n)
func (t *TextDocument) УдалитьСтроку(vmn VMInt, rets *VMSlice) error {
	n, err := t.validateLineNumber(int(vmn))
	if err != nil {
		return err
	}

	t.lines = append(t.lines[:n], t.lines[n+1:]...)
	return nil
}

func (t *TextDocument) ЗаменитьСтроку(vmn VMInt, vmline VMString, rets *VMSlice) error {
	n, err := t.validateLineNumber(int(vmn))
	if err != nil {
		return err
	}

	line := string(vmline)
	// Изменяем только если новая строка действительно является ОДНОЙ строкой
	if !t.isSingleLine(line) {
		return errors.New("Новая строка должна состоять ровно из одной строки")
	}

	t.lines[n] = string(line)
	return nil
}
