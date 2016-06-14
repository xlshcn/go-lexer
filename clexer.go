package golexer

import "unicode"

func cSkipWhitespaces(scanner IRuneScanner, builder ITokenBuilder) TokenType {
	if unicode.IsSpace(scanner.Rune()) {
		for unicode.IsSpace(scanner.Rune()) {
			if scanner.NextRune() == 0 {
				break
			}
		}
	}
	return TOKEN_NULL
}
