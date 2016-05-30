package golexer

import "testing"
import "strings"

func TestLexer(t *testing.T) {
	text := "func TestLexer(t *testing.T) = \"Hello, World\""
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
			t.Logf("Token: %s", token.Literal)
		}
	}
}
