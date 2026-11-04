package parser

import (
	"errors"
	"go/scanner"
	"go/token"
)

// An errorList collects the syntax errors of one .golpe file: the markup
// parser's and go/parser's. The zero value is empty and ready to use.
type errorList struct {
	list scanner.ErrorList
}

// add records the error msg at pos.
func (l *errorList) add(pos token.Position, msg string) {
	l.list.Add(pos, msg)
}

// addGo records the error that go/parser.ParseFile returned: every entry of
// a scanner.ErrorList, or any other error as one entry without a position.
// A nil error adds nothing.
func (l *errorList) addGo(err error) {
	var list scanner.ErrorList
	switch {
	case err == nil:
	case errors.As(err, &list):
		l.list = append(l.list, list...)
	default:
		l.list.Add(token.Position{}, err.Error())
	}
}

// err returns the errors as a scanner.ErrorList sorted by position, keeping
// only the first error of each line, as go/parser does with its own (later
// errors on a line are most likely caused by the first). Errors without a
// position are all kept, first. It returns nil if there are no errors.
//
// It does not use ErrorList.RemoveMultiples, which also drops the errors
// without a position.
func (l *errorList) err() error {
	if len(l.list) == 0 {
		return nil
	}
	sorted := append(scanner.ErrorList(nil), l.list...)
	sorted.Sort()
	var list scanner.ErrorList
	for _, e := range sorted {
		if n := len(list); n > 0 && e.Pos.Line > 0 &&
			list[n-1].Pos.Filename == e.Pos.Filename && list[n-1].Pos.Line == e.Pos.Line {
			continue
		}
		list = append(list, e)
	}
	return list
}
