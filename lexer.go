package golexer

import "io"
import "errors"
import "unicode"
import "bytes"
import "container/list"
import "fmt"

var NullArgumentError = errors.New("Null argument.")

type TokenType int

const (
	TOKEN_NULL TokenType = iota
	TOKEN_EOF
	TOKEN_IDENTIFIER
	TOKEN_NUMBER
	TOKEN_STRING
	TOKEN_UNKNOWN
)

const (
	EMPTY_RUNE            rune = 0
	MAX_TOKEN_BUFFER_SIZE      = 1024
)

type Token struct {
	Type       TokenType
	Literal    string
	LineNumber int
	LinePos    int
}

type ILexerScanner interface {
	NextRune() rune
	Rune() rune
	AppendRune() bool
}

type TokenParser func(scanner ILexerScanner) TokenType

type TokenParsers struct {
	WhitespaceParser TokenParser
	Parsers          []TokenParser
}

func NewTokenParsers() *TokenParsers {
	return &TokenParsers{
		WhitespaceParser: DefaultWhitespaceParser,
		Parsers: []TokenParser{
			DefaultIdentifierParser,
			DefaultNumberParser,
			DefaultQuotedStringParser,
			DefaultCommentParser,
		},
	}
}

func DefaultWhitespaceParser(scanner ILexerScanner) TokenType {
	if unicode.IsSpace(scanner.Rune()) {
		for unicode.IsSpace(scanner.Rune()) {
			if scanner.NextRune() == 0 {
				break
			}
		}
	}
	return TOKEN_NULL
}

func DefaultIdentifierParser(scanner ILexerScanner) TokenType {
	r := scanner.Rune()
	if unicode.IsLetter(r) || r == '_' {
		for unicode.IsLetter(r) || r == '_' || unicode.IsDigit(r) {
			scanner.AppendRune()
			r = scanner.NextRune()
		}
		return TOKEN_IDENTIFIER
	}
	return TOKEN_NULL
}

func DefaultNumberParser(scanner ILexerScanner) TokenType {
	r := scanner.Rune()
	if unicode.IsDigit(r) {
		for unicode.IsDigit(r) {
			scanner.AppendRune()
			r = scanner.NextRune()
		}
		return TOKEN_NUMBER
	}
	return TOKEN_NULL
}

func DefaultQuotedStringParser(scanner ILexerScanner) TokenType {
	quotemark := scanner.Rune()
	if quotemark == '"' || quotemark == '\'' {
		for scanner.NextRune() != 0 {
			if scanner.Rune() == quotemark {
				break
			}
			scanner.AppendRune()
		}
		return TOKEN_STRING
	}
	return TOKEN_NULL
}

func DefaultCommentParser(scanner ILexerScanner) TokenType {
	r := scanner.Rune()
	if r == '#' {
		for scanner.NextRune() != 0 {
			if scanner.Rune() == '\n' {
				break
			}
		}
	}
	return TOKEN_NULL
}

type Lexer struct {
	tokenParsers  *TokenParsers
	scanner       io.RuneScanner
	lineno        int
	pos           int
	r             rune
	eof           bool
	buf           bytes.Buffer
	lastToken     Token
	putbackTokens *list.List
}

func NewLexer(scanner io.RuneScanner, tokenParsers *TokenParsers) (*Lexer, error) {
	if scanner == nil {
		return nil, NullArgumentError
	}
	if tokenParsers == nil {
		tokenParsers = NewTokenParsers()
	}

	lexer := new(Lexer)
	lexer.tokenParsers = tokenParsers
	lexer.scanner = scanner
	lexer.putbackTokens = list.New()

	lexer.NextRune()

	return lexer, nil
}

func (self *Lexer) NextRune() rune {
	if !self.eof {
		r, size, err := self.scanner.ReadRune()
		fmt.Printf("Next Rune: %q (%d)\n", r, size)
		if err != nil {
			r = 0
			self.eof = true
		} else if r == '\n' {
			self.lineno++
			self.pos = 0
		} else {
			self.pos += size
		}
		self.r = r
	}
	return self.r
}

func (self *Lexer) Rune() rune {
	return self.r
}

func (self *Lexer) AppendRune() bool {
	self.buf.WriteRune(self.r)
	return true
}

func (self *Lexer) token(tokenType TokenType) Token {
	self.lastToken = Token{
		Type:       tokenType,
		Literal:    self.buf.String(),
		LineNumber: self.lineno,
		LinePos:    self.pos,
	}
	return self.lastToken
}

func (self *Lexer) IsEnd() bool {
	return self.eof
}

func (self *Lexer) GetToken() (Token, error) {
	if self.putbackTokens.Len() > 0 {
		e := self.putbackTokens.Back()
		ptoken, _ := e.Value.(*Token)
		self.lastToken = *ptoken
		self.putbackTokens.Remove(e)
		return self.lastToken, nil
	}

	self.tokenParsers.WhitespaceParser(self)
	if self.eof {
		return self.token(TOKEN_EOF), nil
	}

	self.buf.Reset()
	for _, parser := range self.tokenParsers.Parsers {
		if parser != nil {
			tokenType := parser(self)
			if tokenType != TOKEN_NULL {
				return self.token(tokenType), nil
			}
		}
	}

	self.AppendRune()
	self.NextRune()
	return self.token(TOKEN_UNKNOWN), nil
}

func (self *Lexer) PutBack(token Token) {
	self.putbackTokens.PushBack(&token)
}
