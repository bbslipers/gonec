package core

import (
	"errors"
	"fmt"

	mcrcon "github.com/Kelwing/mc-rcon"
)

type RconClient struct {
	VMMetaObj

	addr     string
	password string
	conn     *mcrcon.MCConn
}

func (x *RconClient) VMTypeString() string {
	return "СерверМайнкрафт"
}

func (x *RconClient) String() string {
	return fmt.Sprintf("СерверМайнкрафт %s %s", x.password, x.addr)
}

func (x *RconClient) Open(addr, passwd string, closeOnExitHandler bool) error {
	x.addr = addr
	x.password = passwd

	x.conn = new(mcrcon.MCConn)
	err := x.conn.Open(x.addr, x.password)
	if err != nil {
		return err
	}

	err = x.conn.Authenticate()
	if err != nil {
		return err
	}

	return nil
}

func (x *RconClient) SendCommand(command string) (string, error) {
	resp, err := x.conn.SendCommand(command)
	if err != nil {
		return "", err
	}

	return resp, nil
}

func (x *RconClient) IsOnline() bool {
	return x.conn != nil
}

func (x *RconClient) Close() {
	x.conn.Close()
}

func (x *RconClient) VMRegister() {
	x.VMRegisterMethod("Закрыть", x.Закрыть)
	x.VMRegisterMethod("Работает", x.Работает)
	x.VMRegisterMethod("Открыть", x.Открыть)
	x.VMRegisterMethod("Выполнить", x.Выполнить)
}

func (x *RconClient) Открыть(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	if len(args) != 2 {
		return VMErrorNeedArgs(2)
	}
	adr, ok := args[0].(VMString)
	if !ok {
		return errors.New("Первый аргумент должен быть строкой с адресом")
	}
	pass, ok := args[1].(VMString)
	if !ok {
		return errors.New("Второй аргумент должен быть паролем")
	}

	return x.Open(string(adr), string(pass), true)
}

func (x *RconClient) Закрыть(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	x.Close()
	return nil
}

func (x *RconClient) Работает(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	rets.Append(VMBool(x.IsOnline()))
	return nil
}

func (x *RconClient) Выполнить(args VMSlice, rets *VMSlice, envout *(*Env)) error {
	if len(args) != 1 {
		return VMErrorNeedArgs(1)
	}
	command, ok := args[0].(VMString)
	if !ok {
		return errors.New("Первый аргумент должен быть командой")
	}
	res, err := x.SendCommand(string(command))
	if err != nil {
		return err
	}
	rets.Append(VMString(res))
	return nil
}
