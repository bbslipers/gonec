package core

import (
	"errors"
	"math/rand"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

func init() {
	// Для рандома используем не криптографически-безопасный генератор
	rand.Seed(time.Now().UnixNano())
}

// ImportStrings импортирует стандартную библиотеку для работы со строками
func ImportStrings(env *Env) {
	env.DefineS("случайнаястрока", VMFuncTwoParams(func(vma VMString, vmn VMInt, rets *VMSlice) error {
		a, n := []rune(vma.String()), vmn.Int()
		var b strings.Builder

		for i := int64(0); i < n; i++ {
			b.WriteRune(a[rand.Int63n(int64(len(a)))])
		}

		rets.Append(VMString(b.String()))
		return nil
	}))

	env.DefineS("нрег", VMFuncOneParam(func(s VMStringer, rets *VMSlice) error {
		rets.Append(VMString(strings.ToLower(s.String())))
		return nil
	}))

	env.DefineS("врег", VMFuncOneParam(func(s VMStringer, rets *VMSlice) error {
		rets.Append(VMString(strings.ToUpper(s.String())))
		return nil
	}))

	env.DefineS("лев", VMFuncTwoParams(func(s VMStringer, n VMInt, rets *VMSlice) error {
		str := []rune(s.String())
		rets.Append(VMString(str[:n]))
		return nil
	}))

	env.DefineS("прав", VMFuncTwoParams(func(s VMStringer, n VMInt, rets *VMSlice) error {
		str := []rune(s.String())
		rets.Append(VMString(str[len(str)-int(n.Int()):]))
		return nil
	}))

	env.DefineS("сред", VMFuncTwoParamsOptionals(1, func(s VMStringer, l VMInt, rest VMSlice,
		rets *VMSlice,
	) error {
		str := []rune(s.String())
		if len(rest) == 0 {
			rets.Append(VMString(str[l.Int()-1:]))
		} else {
			length, ok := rest[0].(VMInt)
			if !ok {
				return VMErrorNeedInt
			}
			rets.Append(VMString(str[l.Int()-1 : l.Int()-1+length.Int()]))
		}
		return nil
	}))

	env.DefineS("сокрл", VMFuncOneParam(func(s VMStringer, rets *VMSlice) error {
		rets.Append(VMString(strings.TrimLeftFunc(s.String(), unicode.IsSpace)))
		return nil
	}))

	env.DefineS("сокрп", VMFuncOneParam(func(s VMStringer, rets *VMSlice) error {
		rets.Append(VMString(strings.TrimRightFunc(s.String(), unicode.IsSpace)))
		return nil
	}))

	env.DefineS("сокрлп", VMFuncOneParam(func(s VMStringer, rets *VMSlice) error {
		rets.Append(VMString(strings.TrimSpace(s.String())))
		return nil
	}))

	env.DefineS("стрчислострок", VMFuncOneParam(func(s VMStringer, rets *VMSlice) error {
		var doc TextDocument
		doc.fromText(s.String())
		return doc.КоличествоСтрок(rets)
	}))

	env.DefineS("стрполучитьстроку", VMFuncTwoParams(func(s VMStringer, n VMInt, rets *VMSlice) error {
		var doc TextDocument
		doc.fromText(s.String())

		line, err := doc.validateLineNumber(int(n.Int()))
		if err != nil {
			return err
		}
		rets.Append(VMString(doc.lines[line]))
		return nil
	}))

	env.DefineS("стрдлина", VMFuncOneParam(func(s VMStringer, rets *VMSlice) error {
		rets.Append(VMInt(utf8.RuneCountInString(s.String())))
		return nil
	}))

	env.DefineS("стрпустая", VMFuncOneParam(func(s VMStringer, rets *VMSlice) error {
		rets.Append(VMBool(strings.TrimSpace(s.String()) == ""))
		return nil
	}))

	env.DefineS("стрначинаетсяс", VMFuncTwoParams(func(s1, s2 VMStringer, rets *VMSlice) error {
		if s2.String() == "" {
			return errors.New("Префикс не может быть пустым")
		}
		rets.Append(VMBool(strings.HasPrefix(s1.String(), s2.String())))
		return nil
	}))

	env.DefineS("стрзаканчиваетсяна", VMFuncTwoParams(func(s1, s2 VMStringer, rets *VMSlice) error {
		if s2.String() == "" {
			return errors.New("Суффикс не может быть пустым")
		}
		rets.Append(VMBool(strings.HasSuffix(s1.String(), s2.String())))
		return nil
	}))

	env.DefineS("стрсодержит", VMFuncTwoParams(func(s1, s2 VMStringer, rets *VMSlice) error {
		rets.Append(VMBool(strings.Contains(s1.String(), s2.String())))
		return nil
	}))

	env.DefineS("стрсодержитлюбой", VMFuncTwoParams(func(s1, s2 VMStringer, rets *VMSlice) error {
		rets.Append(VMBool(strings.ContainsAny(s1.String(), s2.String())))
		return nil
	}))

	env.DefineS("стрколичество", VMFuncTwoParams(func(s1, s2 VMStringer, rets *VMSlice) error {
		rets.Append(VMInt(strings.Count(s1.String(), s2.String())))
		return nil
	}))

	env.DefineS("стрнайти", VMFuncTwoParams(func(s1, s2 VMStringer, rets *VMSlice) error {
		rets.Append(VMInt(strings.Index(s1.String(), s2.String())))
		return nil
	}))

	env.DefineS("стрнайтилюбой", VMFuncTwoParams(func(s1, s2 VMStringer, rets *VMSlice) error {
		rets.Append(VMInt(strings.IndexAny(s1.String(), s2.String())))
		return nil
	}))

	env.DefineS("стрнайтипоследний", VMFuncTwoParams(func(s1, s2 VMStringer, rets *VMSlice) error {
		rets.Append(VMInt(strings.LastIndex(s1.String(), s2.String())))
		return nil
	}))

	env.DefineS("стрзаменить", VMFuncThreeParams(func(s1, s2, s3 VMStringer, rets *VMSlice) error {
		rets.Append(VMString(strings.Replace(s1.String(), s2.String(), s3.String(), -1)))
		return nil
	}))
}
