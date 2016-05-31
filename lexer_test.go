package golexer

import "testing"
import "strings"

func TestLexer(t *testing.T) {
	text := `
	package golexer
	import "testing"
	import "strings"
	
	var NullArgumentError = errors.New("Null argument.")

	type TokenType int

	const (
		TOKEN_NULL TokenType = itoa
		TOKEN_EOF
	)
	`
	reader := strings.NewReader(text)
	lexer, err := NewLexer(reader, nil)
	if err != nil {
		t.Error("Cannot create a lexer.")
		t.Fail()
		return
	}

	for !lexer.IsEnd() {
		token, err := lexer.GetToken()
		if err != nil {
			t.Logf("Error: %s", err.Error())
		} else {
			t.Logf("%d(%d): %s (%v)", token.LineNumber, token.LinePos, token.Literal, token.Type)
		}
	}
}
