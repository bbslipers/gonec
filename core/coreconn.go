package core

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/shinanca/gonec/names"
)

func NewVMConn(data VMValue) *VMConn {
	return &VMConn{
		id:     -1,
		closed: false,
		uid:    uuid.NewV4().String(),
		data:   data,
		httpcl: nil,
		dialer: &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		},
	}
}

type VMConn struct {
	conn net.Conn

	dialer *net.Dialer
	httpcl *http.Client // клиент http
	ctx    context.Context
	cancel context.CancelFunc

	id     int
	closed bool
	uid    string
	data   VMValue
	gzip   bool
}

func (c *VMConn) VMTypeString() string { return "Соединение" }

func (c *VMConn) Interface() interface{} {
	return c.conn
}

func (c *VMConn) String() string {
	if c.closed {
		return fmt.Sprintf("Соединение (закрыто)")
	}
	if c.httpcl != nil {
		return "Соединение HTTP"
	}
	return fmt.Sprintf("Соединение TCP с %s", c.conn.RemoteAddr())
}

func urlValuesFromMap(vals VMStringMap) (url.Values, error) {
	uvs := make(url.Values)
	for k, v := range vals {
		vv, ok := v.(VMStringer)
		if !ok {
			return nil, VMErrorNeedString
		}
		uvs.Set(k, vv.String())
	}
	return uvs, nil
}

// пример работы с контекстом

// cx, cancel := context.WithCancel(context.Background())
// req, _ := http.NewRequest("GET", "http://google.com", nil)
// req = req.WithContext(cx)
// ch := make(chan error)

// go func() {
// 	_, err := http.DefaultClient.Do(req)
// 	select {
// 	case <-cx.Done():
// 		// Already timedout
// 	default:
// 		ch <- err
// 	}
// }()

// // Simulating user cancel request
// go func() {
// 	time.Sleep(100 * time.Millisecond)
// 	cancel()
// }()
// select {
// case err := <-ch:
// 	if err != nil {
// 		// HTTP error
// 		panic(err)
// 	}
// 	print("no error")
// case <-cx.Done():
// 	panic(cx.Err())
// }

// HttpReq выполняет универсальный (с любыми методами) запрос к серверу и ждет ответа
// hdrs - заголовки, которые будут помещены в запрос
// vals - если это GET, то будут помещены в URL, если POST - помещаются в FormValues тела запроса, иначе - игнорируются
func (x *VMConn) HttpReq(meth, rurl VMString, body []byte, hdrs, vals VMStringMap) (*VMHttpResponse, error) {
	var req *http.Request
	var err error

	// если указаны vals, то body игнорируется
	if meth == VMString("POST") && len(vals) > 0 {
		uvs, err := urlValuesFromMap(vals)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest(string(meth), string(rurl), strings.NewReader(uvs.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	} else if meth == VMString("GET") && len(vals) > 0 {
		uvs, err := urlValuesFromMap(vals)
		if err != nil {
			return nil, err
		}
		nurl, err := url.Parse(string(rurl))
		if err != nil {
			return nil, err
		}
		nurl.RawQuery = uvs.Encode()
		req, err = http.NewRequest(string(meth), nurl.String(), bytes.NewReader(body))

	} else {
		req, err = http.NewRequest(string(meth), string(rurl), bytes.NewReader(body))
	}

	if err != nil {
		return nil, err
	}

	// заворачиваем в контекст для возможности прерывания
	x.ctx, x.cancel = context.WithCancel(context.Background())
	req = req.WithContext(x.ctx)

	for k, v := range hdrs {
		vv, ok := v.(VMStringer)
		if !ok {
			return nil, VMErrorNeedString
		}
		req.Header.Add(k, vv.String())
	}

	var resp *http.Response

	resp, err = x.httpcl.Do(req)

	res := &VMHttpResponse{r: resp, data: x.data}
	if err != nil {
		res.Close()
		return nil, err
	}

	_, err = res.ReadBody() // читаем ответ и закрываем канал, оставив копию в слайсе, для множественного чтения

	return res, err
}

func (x *VMConn) Dial(proto, addr string, handler VMFunc, closeOnExitHandler bool) (err error) {
	x.httpcl = nil

	if proto == "tcptls" {
		// certPool := x509.NewCertPool()
		// certPool.AppendCertsFromPEM(TLSCertGonec)
		config := &tls.Config{
			// RootCAs: certPool,
			InsecureSkipVerify: true,
		}
		x.conn, err = tls.DialWithDialer(x.dialer, "tcp", addr, config)
		if err != nil {
			return err
		}
	}

	if proto == "tcp" || proto == "tcpzip" {
		x.conn, err = x.dialer.Dial("tcp", addr)
		if err != nil {
			return err
		}
		if proto == "tcpzip" {
			x.gzip = true
		}
	}

	if proto == "http" {
		tr := &http.Transport{
			Proxy:       http.ProxyFromEnvironment,
			DialContext: x.dialer.DialContext,
			// func(ctx context.Context, network, addr string) (net.Conn, error) {
			// 	c, err := x.dialer.DialContext(ctx, network, addr)
			// 	x.conn = c
			// 	return c, err
			// },
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}

		x.httpcl = &http.Client{Transport: tr}
	}

	if handler != nil {
		go x.Handle(handler, closeOnExitHandler)
	}

	return nil
}

func (x *VMConn) Handle(f VMFunc, closeOnExitHandler bool) {
	args := make(VMSlice, 1)
	rets := make(VMSlice, 0)
	args[0] = x
	err := f(args, &rets)
	// закрываем по окончании обработки
	if closeOnExitHandler {
		x.Close()
	}
	if err != nil {
		fmt.Println(err)
	}
}

func (x *VMConn) Close() (err error) {
	if x.httpcl != nil {
		x.cancel()
	}
	if x.conn != nil {
		err = x.conn.Close()
	}
	x.closed = true
	return
}

type binTCPHead struct {
	Signature [8]byte //[8]byte{'g', 'o', 'n', 'e', 'c', 't', 'c', 'p'}
	Hash      uint64  // хэш зашифрованного тела
	Len       int64   // длина тела
	Gzip      byte    //==0 - без сжатия (зашифрован), иначе сжат и зашифрован
}

func (x *VMConn) Send(val VMStringMap) error {
	b, err := val.MarshalBinary()
	if err != nil {
		return err
	}

	var be []byte
	if x.gzip {
		be, err = GZip(b)
		if err != nil {
			return err
		}
		be, err = EncryptAES128(be)
	} else {
		be, err = EncryptAES128(b)
	}

	if err != nil {
		return err
	}

	// хэш зашифрованного
	hs := HashBytes(be)

	head := binTCPHead{
		Signature: [8]byte{'g', 'o', 'n', 'e', 'c', 't', 'c', 'p'},
		Hash:      hs,
		Len:       int64(len(be)),
	}

	if x.gzip {
		head.Gzip = 1
	}

	// log.Println("out", hs, be)

	err = binary.Write(x.conn, binary.LittleEndian, head)
	if err != nil {
		if err == io.EOF {
			x.Close()
		}
		return err
	}

	_, err = io.Copy(x.conn, bytes.NewReader(be))
	if err != nil {
		if err == io.EOF {
			x.Close()
		}
		return err
	}
	return nil
}

func (x *VMConn) Receive() (VMStringMap, error) {
	rv := make(VMStringMap)
	var buf bytes.Buffer

	var head binTCPHead

	err := binary.Read(x.conn, binary.LittleEndian, &head)
	if err != nil {
		if err == io.EOF {
			x.Close()
			err = VMErrorEOF
		}
		return rv, err
	}

	// проверяем целостность полученного сообщения
	// сначала идет заголовок
	// затем тело

	if head.Signature != [8]byte{'g', 'o', 'n', 'e', 'c', 't', 'c', 'p'} {
		return rv, errors.New(VMErrorIncorrectMessage.Error() + " - неверная сигнатура")
	}

	buf.Reset()
	_, err = io.CopyN(&buf, x.conn, head.Len)
	if err != nil {
		if err == io.EOF {
			x.Close()
		}
		return rv, err
	}

	b := buf.Bytes()

	// хэш зашифрованного
	hb := HashBytes(b)
	if hb != head.Hash {
		// log.Println("in", hb, b)
		return rv, errors.New(VMErrorIncorrectMessage.Error() + " - не совпал хэш")
	}
	// проверили хэш, все ок - получаем VMStringMap

	bd, err := DecryptAES128(b)
	if err != nil {
		return rv, err
	}

	if head.Gzip != 0 {
		bd, err = UnGZip(bd)
	}
	if err != nil {
		return rv, err
	}

	if err := (&rv).UnmarshalBinary(bd); err != nil {
		return rv, err
	}
	return rv, nil
}

func (c *VMConn) MethodMember(name int) (VMFunc, bool) {
	// только эти методы будут доступны из кода на языке Гонец!

	switch names.UniqueNames.GetLowerCase(name) {
	case "получить":
		return VMFuncZeroParams(c.Получить), true
	case "отправить":
		return VMFuncOneParam(c.Отправить), true
	case "закрыто":
		return VMFuncZeroParams(c.Закрыто), true
	case "идентификатор":
		return VMFuncZeroParams(c.Идентификатор), true
	case "данные":
		return VMFuncZeroParams(c.Данные), true
	case "запрос":
		// метод, урл, тело, заголовки, параметры формы
		return VMFuncOneParam(c.Запрос), true
	case "закрыть":
		return VMFuncZeroParams(c.Закрыть), true
	}

	return nil, false
}

func (x *VMConn) Идентификатор(rets *VMSlice) error {
	rets.Append(VMString(x.uid))
	return nil
}

func (x *VMConn) Получить(rets *VMSlice) error {
	if x.httpcl != nil {
		return VMErrorWrongHTTPMethod
	}
	// TCP
	v, err := x.Receive()
	rets.Append(v)
	return err // при ошибке вызовет исключение, нужно обрабатывать в попытке
}

func (x *VMConn) Отправить(m VMStringMap, rets *VMSlice) error {
	if x.httpcl != nil {
		return VMErrorWrongHTTPMethod
	}
	// при ошибке вызовет исключение, нужно обрабатывать в попытке
	return x.Send(m)
}

func (x *VMConn) Закрыто(rets *VMSlice) error {
	rets.Append(VMBool(x.closed))
	return nil
}

func (x *VMConn) Данные(rets *VMSlice) error {
	rets.Append(x.data)
	return nil
}

func (x *VMConn) Закрыть(rets *VMSlice) error {
	x.Close()
	return nil
}

func (x *VMConn) Запрос(vsm VMStringMap, rets *VMSlice) error {
	if x.httpcl == nil {
		return VMErrorNonHTTPMethod
	}

	var m, p, b VMString
	var h, vals VMStringMap

	if v, ok := vsm["Метод"]; ok {
		if m, ok = v.(VMString); !ok {
			return VMErrorNeedString
		}
	}
	if v, ok := vsm["Путь"]; ok {
		if p, ok = v.(VMString); !ok {
			return VMErrorNeedString
		}
	}
	if v, ok := vsm["Тело"]; ok {
		if b, ok = v.(VMString); !ok {
			return VMErrorNeedString
		}
	}
	if v, ok := vsm["Заголовки"]; ok {
		if h, ok = v.(VMStringMap); !ok {
			return VMErrorNeedMap
		}
	}
	if v, ok := vsm["Параметры"]; ok {
		if vals, ok = v.(VMStringMap); !ok {
			return VMErrorNeedMap
		}
	}

	r, err := x.HttpReq(m, p, []byte(b), h, vals)
	if err != nil {
		return err
	}

	rets.Append(r)

	return nil
}
