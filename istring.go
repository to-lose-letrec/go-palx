package main

import "fmt"

type istring struct {
	index  int
	row    int
	column int
	str    string
	full   string
}

func newIstring(idx, row int, str string) *istring {
	return &istring{
		index: idx,
		row:   row,
		str:   str,
		full:  str,
	}
}

func (l *istring) String() string { return l.str }

func (l *istring) advanceColumn(n int) (col int) {
	col = l.column
	for i := 0; i < n; i++ {
		if l.str[i] == '\t' {
			col += 8 - (col % 8)
		} else {
			col++
		}
	}
	return
}

func (l *istring) consume(n int) *istring {
	col := l.advanceColumn(n)
	return &istring{l.index, l.row, col, l.str[n:], l.full}
}

func (l *istring) trunc(n int) *istring {
	return &istring{l.index, l.row, l.column, l.str[:n], l.full}
}

func (l *istring) isEmpty() bool {
	return len(l.str) == 0
}

func (l *istring) startsWith(p bytePredicate) bool {
	return len(l.str) > 0 && p(l.str[0])
}

func (l *istring) startsWithString(s string) bool {
	return len(l.str) >= len(s) && l.str[:len(s)] == s
}

func (l *istring) scanWhile(p bytePredicate) (i int) {
	for ; i < len(l.str) && p(l.str[i]); i++ {
	}
	return
}

func (l *istring) scanUntil(p bytePredicate) (i int) {
	for ; i < len(l.str) && !p(l.str[i]); i++ {
	}
	return
}

func (l *istring) consumeWhile(p bytePredicate) (consumed, remain *istring) {
	i := l.scanWhile(p)
	consumed, remain = l.trunc(i), l.consume(i)
	return
}

func (l *istring) consumeUntil(p bytePredicate) (consumed, remain *istring) {
	i := l.scanUntil(p)
	consumed, remain = l.trunc(i), l.consume(i)
	return
}

func (l *istring) consumeWhitespace() *istring {
	return l.consume(l.scanWhile(whitespace))
}

func (l *istring) stripTrailingComment() *istring {
	lastNonWS := 0
	for i := 0; i < len(l.str); i++ {
		if comment(l.str[i]) {
			break
		}
		if stringQuote(l.str[i]) {
			q := l.str[i]
			i++
			for ; i < len(l.str) && l.str[i] != q; i++ {
			}
			lastNonWS = i
			if i == len(l.str) {
				break
			}
		}
		if !whitespace(l.str[i]) {
			lastNonWS = i + 1
		}
	}
	return l.trunc(lastNonWS)
}

// Favor composition over dedicated functions
func charEqual(c byte) bytePredicate {
	return func(b byte) bool { return b == c }
}

func findUnquotedChar(c byte) bytePredicate {
	return func(b byte) bool {
		var quote byte

		if quote == 0 {
			if b == c {
				return true
			}
			if b == '\'' || b == '"' {
				quote = b
			}
		} else {
			if b == quote {
				quote = 0
			}
		}
		return false
	}
}

