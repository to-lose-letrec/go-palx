package main

import "fmt"

type bytePredicate func(c byte) bool

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

// Byte Predicates
func whitespace(c byte) bool { return c == ' ' || c == '\t' }

func wordChar(c byte) bool { return !whitespace(c) }

func comment(c byte) bool { return c == ';' }

// No need for hexadecimal here - the PDP-8 never used it.
func binary(c byte) bool { return c == '0' || c == '1' }

func octal(c byte) bool { return decimal(c) || (c >= '0' && c <= '7') }

func decimal(c byte) bool { return (c >= '0' && c <= '9') }

func alpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func indirect(c byte) bool {
	return c == '@'
}

func immediate(c byte) bool {
	return c == '['
}

func labelStartChar(c byte) bool {
	return alpha(c) || c == '_' || c == '.'
}

func labelChar(c byte) bool { return labelStartChar(c) || decimal(c) }

func identifierStartChar(c byte) bool { return labelStartChar(c) }

func identifierChar(c byte) bool { return labelChar(c) || c == ':' }

func stringQuote(c byte) bool { return c == '"' || c == '\'' }


func Foo() {
	istr := newIstring(0, 1, "LD X,#0 '; Quoted comment' ; Trailing comment")

	o := istr.stripTrailingComment()
	fmt.Printf("stripTrailingComment() returned: %s\n",
		o)
}
