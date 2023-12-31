package core

import (
	"errors"
	"fmt"
)

type VMClient struct {
	VMMetaObj // должен передаваться по ссылке, поэтому это будет объект метаданных

	addr     string  // [addr]:port
	protocol string  // tcp, json, http
	conn     *VMConn // клиент tcp, каждому соединению присваивается GUID
}

func (x *VMClient) VMTypeString() string {
	return "Клиент"
}

func (x *VMClient) String() string {
	return fmt.Sprintf("Клиент %s %s", x.protocol, x.addr)
}

func (x *VMClient) IsOnline() bool {
	return x.conn != nil && !x.conn.closed
}

func (x *VMClient) Open(proto, addr string, handler VMFunc, data VMValue, closeOnExitHandler bool) error {
	switch proto {
	case "tcp", "tcpzip", "tcptls", "http", "https":

		x.conn = NewVMConn(data)
		err := x.conn.Dial(proto, addr, handler, closeOnExitHandler)
		if err != nil {
			return err
		}

	default:
		return VMErrorIncorrectProtocol
	}
	return nil
}

func (x *VMClient) Close() {
	x.conn.Close()
}

func (x *VMClient) VMRegister() {
	x.VMRegisterMethod("Закрыть", x.Закрыть)
	x.VMRegisterMethod("Работает", x.Работает)
	x.VMRegisterMethod("Открыть", VMFuncNParams(4, x.Открыть))     // асинхронно
	x.VMRegisterMethod("Соединить", VMFuncNParams(2, x.Соединить)) // синхронно

	// tst.VMRegisterField("ПолеСтрока", &tst.ПолеСтрока)
}

func (x *VMClient) Открыть(args VMSlice, rets *VMSlice) error {
	p, ok := args[0].(VMString)
	if !ok {
		return errors.New("Первый аргумент должен быть строкой с типом канала")
	}
	adr, ok := args[1].(VMString)
	if !ok {
		return errors.New("Второй аргумент должен быть строкой с адресом")
	}
	f, ok := args[2].(VMFunc)
	if !ok {
		return errors.New("Третий аргумент должен быть функцией с одним аргументом-соединением")
	}

	return x.Open(string(p), string(adr), f, args[3], true)
}

func (x *VMClient) Соединить(args VMSlice, rets *VMSlice) error {
	p, ok := args[0].(VMString)
	if !ok {
		return errors.New("Первый аргумент должен быть строкой с типом канала")
	}
	adr, ok := args[1].(VMString)
	if !ok {
		return errors.New("Второй аргумент должен быть строкой с адресом")
	}

	err := x.Open(string(p), string(adr), nil, VMNil, false) // не запускает handler
	if err != nil {
		return err
	}

	if x.conn == nil {
		return errors.New("Соединение не было установлено")
	}

	rets.Append(x.conn)

	return nil
}

func (x *VMClient) Закрыть(args VMSlice, rets *VMSlice) error {
	x.Close()
	return nil
}

func (x *VMClient) Работает(args VMSlice, rets *VMSlice) error {
	rets.Append(VMBool(x.IsOnline()))
	return nil
}
