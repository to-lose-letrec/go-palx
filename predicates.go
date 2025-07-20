package main

// Parameter count predicates
type countPredicate func(int) bool

func any(a int) bool {
	return true
}

func atLeast(a int) countPredicate {
	return func(b int) bool {
		return b >= a
	}
}

func atMost(a int) countPredicate {
	return func(b int) bool {
		return b <= a
	}
}

func exactly(a int) countPredicate {
	return func(b int) bool {
		return a == b
	}
}

func none(a int) countPredicate {
	return exactly(0)
}

// Byte Predicates
type bytePredicate func(c byte) bool

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
