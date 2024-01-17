package core

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	VMErrorNeedLess             = errors.New("Первое значение должно быть меньше второго")
	VMErrorNeedLengthOrBoundary = errors.New("Должна быть длина диапазона или начало и конец")
	VMErrorNeedFormatAndArgs    = errors.New("Должны быть форматная строка и хотя бы один параметр")
	VMErrorSmallDecodeBuffer    = errors.New("Мало данных для декодирования")

	VMErrorNeedString      = errors.New("Требуется значение типа Строка")
	VMErrorNeedBool        = errors.New("Требуется значение типа Булево")
	VMErrorNeedInt         = errors.New("Требуется значение типа ЦелоеЧисло")
	VMErrorNeedDecNum      = errors.New("Требуется значение типа Число")
	VMErrorNeedDate        = errors.New("Требуется значение типа Дата")
	VMErrorNeedMap         = errors.New("Требуется значение типа Структура")
	VMErrorNeedSlice       = errors.New("Требуется значение типа Массив")
	VMErrorNeedDuration    = errors.New("Требуется значение типа Длительность")
	VMErrorNeedHash        = errors.New("Параметр не может быть хэширован")
	VMErrorNeedBinaryTyper = errors.New("Требуется значение, которое может быть сериализовано в бинарное")

	VMErrorIndexOutOfBoundary  = errors.New("Индекс находится за пределами массива")
	VMErrorNotConverted        = errors.New("Приведение к типу невозможно")
	VMErrorUnknownType         = errors.New("Неизвестный тип данных")
	VMErrorIncorrectFieldType  = errors.New("Поле структуры имеет другой тип")
	VMErrorIncorrectStructType = errors.New("Невозможно использовать данный тип структуры")
	VMErrorNotDefined          = errors.New("Не определено")
	VMErrorNotBinaryConverted  = errors.New("Значение не может быть преобразовано в бинарный формат")

	VMErrorNoNeedArgs = errors.New("Параметры не требуются")
	VMErrorNoArgs     = errors.New("Отсутствуют аргументы")

	VMErrorIncorrectOperation = errors.New("Операция между значениями невозможна")
	VMErrorUnknownOperation   = errors.New("Неизвестная операция")

	VMErrorServerNowOnline   = errors.New("Сервер уже запущен")
	VMErrorServerOffline     = errors.New("Сервер уже остановлен")
	VMErrorIncorrectProtocol = errors.New("Неверный протокол")
	VMErrorIncorrectClientId = errors.New("Неверный идентификатор соединения")
	VMErrorIncorrectMessage  = errors.New("Неверный формат сообщения")
	VMErrorEOF               = errors.New("Недостаточно данных в источнике")

	VMErrorServiceNotReady          = errors.New("Сервис не готов") // устанавливается сервисами в случае прекращения работы
	VMErrorServiceAlreadyRegistered = errors.New("Сервис уже зарегистрирован с таким же ID")
	VMErrorServerAlreadyStarted     = errors.New("Сервер уже запущен")
	VMErrorWrongHTTPMethod          = errors.New("Метод не применим к HTTP-соединению")
	VMErrorNonHTTPMethod            = errors.New("Метод применим только к HTTP-соединению")
	VMErrorHTTPResponseMethod       = errors.New("Метод применим только к ответу HTTP сервера")
	VMErrorNilResponse              = errors.New("Отсутствует содержимое ответа")

	VMErrorTransactionIsOpened  = errors.New("Уже была открыта транзакция")
	VMErrorTransactionNotOpened = errors.New("Не открыта транзакция")
	VMErrorTableNotExists       = errors.New("Отсутствует таблица в базе данных")
	VMErrorWrongDBValue         = errors.New("Невозможно распознать значение в базе данных")

	VMErrorEan13Format = errors.New("Ean13 состоит из 12 цифр")
	VMErrorI2Of5Format = errors.New("I2Of5 состоит из чётного количества цифр")

	VMErrorSaveXML = errors.New("Ошибки при записи XML")

	VMErrorReadXlsxTemplate = errors.New("Ошибка при чтении шаблона xlsx")
	VMErrorFillXlsx         = errors.New("Ошибка при заполнении шаблона xslx данными")
	VMErrorSaveXlsx         = errors.New("Ошибка при сохранении заполненного xlsx шаблона")

	VMErrorOpenXlsxFile = errors.New("Ошибка при открытии xlsx файла")
	VMErrorSetCellValue = errors.New("Ошибка при установки значения ячейки")
	VMErrorAddSheet     = errors.New("Ошибка при добавлении новой страницы")
	VMErrorSaveXlsxFile = errors.New("Ошибка при записи xlsx файла")

	VMErrorEncoding = errors.New("Ошибка при кодировании строки")
)

func VMErrorNeedArgs(n int) error {
	return fmt.Errorf("Неверное количество параметров (требуется %d)", n)
}

func VMErrorMaxArgs(n int) error {
	return fmt.Errorf("Неверное количество параметров (максимум %d)", n)
}

var argIndexStrs = [3]string{"Первым", "Вторым", "Третьим"}

func interfaceType[V any]() reflect.Type {
	return reflect.TypeOf((*V)(nil)).Elem()
}

func typeString[V VMValue]() string {
	var v V
	if any(v) == nil {
		// an interface would initialize to nil
		vtype := reflect.TypeOf(&v).Elem()
		if vtype.Implements(interfaceType[VMStringer]()) {
			return "значение, приводимое к Строке"
		} else if vtype.Implements(interfaceType[VMNumberer]()) {
			return "значение, приводимое к Числу"
		} else if vtype.Implements(interfaceType[VMDateTimer]()) {
			return "значение, приводимое к Дате"
		} else if vtype.Implements(interfaceType[VMIndexer]()) {
			return "значение, имеющее длину"
		} else if vtype.Implements(interfaceType[VMHasher]()) {
			return "хешируемое значение"
		} else if vtype.Implements(interfaceType[VMBinaryTyper]()) {
			return "значение, сериализуемое в бинарное"
		} else if vtype.Implements(interfaceType[VMValue]()) {
			return "значение любого типа Гонца"
		}
		panic("Не известно строковое представление типа")
	}
	return "значение типа " + v.VMTypeString()
}

func VMErrorNeedArgType[V VMValue](i, n int) error {
	if i >= n {
		panic("VMErrorNeedArgType: i < n")
	}
	indStr, typeStr := argIndexStrs[i], typeString[V]()

	if n == 1 {
		return fmt.Errorf("Требуется %s", typeStr)
	} else {
		return fmt.Errorf("%s параметром требуется %s", indStr, typeStr)
	}
}
