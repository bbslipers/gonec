package main

import (
	"fmt"
	"log"
	"testing"

	"github.com/covrom/gonec/bincode"
	"github.com/covrom/gonec/core"
	"github.com/covrom/gonec/parser"
)

func TestRun(t *testing.T) {
	env := core.NewEnv()

	script := `
	дтнач = ТекущаяДата()
	а = [](0,1000000)
	для н=1 по 1000000 цикл
	  а=а+[н]+[н*10]
	конеццикла
	к=0
	для каждого н из а цикл
	  к=к+н
	конеццикла
	сообщить(к, ПрошлоВремениС(дтнач))
	
	#gonec.exe -web -t
	#go tool pprof -svg ./gonec.exe http://localhost:5000/debug/pprof/profile?seconds=10 > cpu.svg
	
	а=Новый("__функциональнаяструктуратест__",{"ПолеСтрока": "цщушаццке", "ПолеЦелоеЧисло": 3456})
  сообщить(а)
	сообщить("Хэш", Хэш(а))
	а=Новый("__функциональнаяструктуратест__")
	а.Полецелоечисло = 2243
	а.ПолеСтрока = "авузлхвсзщл"
	сообщить(а)
	сообщить("Хэш", Хэш(а))
	сообщить(а.ВСтроку(), а.ПолеСтрока)
	Сообщить(Строка(а))
	Сообщить(Структура(Строка(а)))
	б=Новый("__ФункциональнаяСтруктураТест__", Строка(а)) //получаем объект из строки json
	сообщить("Из json:",б.ВСтроку(), б.ПолеСтрока)
		
	# массив с вложенным массивом со структурой и датой
	а=[2, 1, [3, {"привет":2, "а":1}, Дата("2017-08-17T09:23:00+03:00")], 4]
	а[2][1].а=Дата("2017-08-17T09:23:00+03:00")
	# приведение массива или структуры к типу "строка" означает сериализацию в json, со всеми вложенными массивами и структурами
	Сообщить(а, а[2][1].а)
	Сообщить(Строка(а))
	Сообщить("Ключи в порядке сортировки:", Ключи(а[2][1]))
	# приведение строки к массиву или структуре означает десериализацию из json
	Сообщить(Массив("[1,2,3.5,4]"))
	Сообщить(Массив(Строка(а)))
	а=[2,1,4.5,3]
	Сообщить(а, а[2]*2)
	Сообщить(Строка(а))
	а.Сортировать()
	а.Обратить()
	Сообщить("После сортировки и обращения:",а)
	
	функция фиб(н)
	  если н = 0 тогда
		возврат 0
	  иначеесли н = 1 тогда
		возврат 1
	  конецесли
	  возврат фиб(н-1) + фиб(н-2)
	конецфункции
	
	сообщить(фиб(10))
	
	функция фибт(н0, н1, к)
		если к = 0 тогда
		  возврат н0
		иначеесли к = 1 тогда
		  возврат н1
		конецесли
		возврат фибт(н1, н0+н1, к-1)
	конецфункции
	  
	функция фиб2(н)
		возврат фибт(0, 1, н)
	конецфункции
	
	сообщить(фиб2(10))
	
	функция фиб3(н)
	  если н = 0 тогда
		возврат 0
	  иначеесли н = 1 тогда
		возврат 1
	  конецесли
	  н0, н1 = 0, 1
	  для к = н по 2 цикл
		тмп = н0 + н1
		н0 = н1
		н1 = тмп
	  конеццикла
	  возврат н1
	конецфункции
	
	сообщить(фиб3(10))
	`
	
	parser.EnableErrorVerbose()
	_, stmts, err := bincode.ParseSrc(script)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stmts)
	_, err = bincode.Run(stmts, env)
	if err != nil {
		log.Fatal(err)
	}
}
