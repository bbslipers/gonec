package gonec

import (
	"io"
	"log"
	"net/http"
	"sync"
	"time"
	"fmt"

	"github.com/covrom/gonec/gonecparser/ast"
	"github.com/covrom/gonec/gonecparser/parser"
	"github.com/covrom/gonec/gonecparser/token"
)

// APIPath содержит путь к api интерпретатора
const APIPath = "/gonec"

type clientConnection struct {
	IP string
}

type interpreter struct {
	sync.RWMutex
	clientConnections []clientConnection
}

// Interpreter возвращает новый интерпретатор
func Interpreter() *interpreter {
	return &interpreter{}
}

// handlerMain обрабатывает входящие запросы к интерпретатору через POST-запросы
// тело запроса - это код для интерпретации
func (i *interpreter) handlerMain(w http.ResponseWriter, r *http.Request) {

	i.RLock()
	//лимит количества одновременных подключений к одному интерпретатору
	overconn := len(i.clientConnections) >= 2
	i.RUnlock()

	if overconn {
		time.Sleep(300 * time.Millisecond) //анти-ddos
		http.Error(w, "Слишком много запросов обрабатывается в данный момент", http.StatusForbidden)
		return
	}

	clconn := clientConnection{
		IP: r.RemoteAddr,
	}

	i.Lock()
	i.clientConnections = append(i.clientConnections, clconn)
	i.Unlock()

	defer func(cc clientConnection) {
		i.Lock()
		for n := range i.clientConnections {
			if i.clientConnections[n] == cc {
				i.clientConnections = append(i.clientConnections[:n], i.clientConnections[n+1:]...)
				break
			}
		}
		i.Unlock()
	}(clconn)

	if r.ContentLength > 1<<26 {
		time.Sleep(time.Second) //анти-ddos
		http.Error(w, "Слишком большой запрос", http.StatusForbidden)
		return
	}

	switch r.Method {

	case http.MethodPost:

		defer r.Body.Close()
		//интерпретируется код и возвращается вывод как простой текст
		w.Header().Set("Content-Type", "text/plain")
		err := i.ParseAndRun(r.Body, w)
		if err != nil {
			time.Sleep(time.Second) //анти-ddos
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	default:
		time.Sleep(time.Second) //анти-ddos
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
}

// Run запускает микросервис интерпретатора по адресу и порту
func (i *interpreter) Run(srv string) {
	http.HandleFunc(APIPath, i.handlerMain)
	log.Fatal(http.ListenAndServe(srv, nil))
}

// ParseAndRun разбирает запрос и запускает интерпретацию
func (i *interpreter) ParseAndRun(r io.Reader, w io.Writer) (err error) {

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", r, parser.Trace)

	if err != nil {
		return
	}

	for _, u := range f.Unresolved {
		fmt.Println("Неразыменовано: " + u.Name)
	}

	ast.Inspect(f, func(n ast.Node) bool {
		var s string
		switch x := n.(type) {
		case *ast.BasicLit:
			s = x.Value
		case *ast.Ident:
			s = x.Name
		}
		if s != "" {
			fmt.Printf("%s:\t%s\n", fset.Position(n.Pos()), s)
		}
		return true
	})

	// TODO: синхронно запускается код модуля, но он может создавать вэб-сервера и горутины, которые будут работать и после возврата

	return nil
}
