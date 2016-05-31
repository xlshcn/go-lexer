package golexer

import "io"
import "errors"
import "unicode"
import "bytes"
import "container/list"

var NullArgumentError = errors.New("Null argument.")
var EofError = errors.New("End of file.")

type TokenType int

const (
	TOKEN_NULL TokenType = iota
	TOKEN_EOF
	TOKEN_IDENTIFIER
	TOKEN_NUMBER
	TOKEN_STRING
	TOKEN_UNKNOWN
)

type Token struct {
	Type       TokenType
	Literal    string
	LineNumber int
	LinePos    int
}

type IRuneScanner interface {
	NextRune() rune
	Rune() rune
}

type ITokenBuilder interface {
	AppendRune() bool
}

// The TokenParser reads the runes from the scanner, recognize the rune and build the token via
// ITokenBuilder.
// The function returns TOKEN_NULL to indicate that no match found. Otherwise, returns the recognized
// token type. The value of the token should be built and stored in ITokenBuilder.
type TokenParser func(scanner IRuneScanner, builder ITokenBuilder) TokenType

type TokenParsers struct {
	SkipWhitespaces TokenParser
	Parsers         []TokenParser
}

func NewTokenParsers(skipWhitespaces TokenParser, parsers ...TokenParser) *TokenParsers {
	tp := TokenParsers{
		skipWhitespaces,
		make([]TokenParser, len(parsers)),
	}
	for index, parser := range parsers {
		tp.Parsers[index] = parser
	}
	return &tp
}

func NewDefaultTokenParsers() *TokenParsers {
	return NewTokenParsers(
		DefaultSkipWritespaces,
		DefaultIdentifierParser,
		DefaultNumberParser,
		DefaultQuotedStringParser,
		DefaultCommentParser)
}

func DefaultSkipWritespaces(scanner IRuneScanner, builder ITokenBuilder) TokenType {
	if unicode.IsSpace(scanner.Rune()) {
		for unicode.IsSpace(scanner.Rune()) {
			if scanner.NextRune() == 0 {
				break
			}
		}
	}
	return TOKEN_NULL
}

func DefaultIdentifierParser(scanner IRuneScanner, builder ITokenBuilder) TokenType {
	r := scanner.Rune()
	if unicode.IsLetter(r) || r == '_' {
		for unicode.IsLetter(r) || r == '_' || unicode.IsDigit(r) {
			builder.AppendRune()
			r = scanner.NextRune()
		}
		return TOKEN_IDENTIFIER
	}
	return TOKEN_NULL
}

func DefaultNumberParser(scanner IRuneScanner, builder ITokenBuilder) TokenType {
	r := scanner.Rune()
	if unicode.IsDigit(r) {
		for unicode.IsDigit(r) {
			builder.AppendRune()
			r = scanner.NextRune()
		}
		return TOKEN_NUMBER
	}
	return TOKEN_NULL
}

func DefaultQuotedStringParser(scanner IRuneScanner, builder ITokenBuilder) TokenType {
	quotemark := scanner.Rune()
	if quotemark == '"' || quotemark == '\'' {
		for scanner.NextRune() != 0 {
			if scanner.Rune() == quotemark {
				scanner.NextRune()
				break
			}
			builder.AppendRune()
		}
		return TOKEN_STRING
	}
	return TOKEN_NULL
}

func DefaultCommentParser(scanner IRuneScanner, builder ITokenBuilder) TokenType {
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
	lastPos       int
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
		tokenParsers = NewDefaultTokenParsers()
	}

	lexer := new(Lexer)
	lexer.tokenParsers = tokenParsers
	lexer.scanner = scanner
	lexer.putbackTokens = list.New()

	// read the first rune to kick off the lexer scan process.
	lexer.NextRune()

	return lexer, nil
}

func (self *Lexer) NextRune() rune {
	if !self.eof {
		r, size, err := self.scanner.ReadRune()
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
		LinePos:    self.lastPos,
	}
	return self.lastToken
}

func (self *Lexer) IsEnd() bool {
	return self.eof
}

func (self *Lexer) GetToken() (Token, error) {
	// Handles the putback tokens first.
	if self.putbackTokens.Len() > 0 {
		e := self.putbackTokens.Back()
		ptoken, _ := e.Value.(*Token)
		self.lastToken = *ptoken
		self.putbackTokens.Remove(e)
		return self.lastToken, nil
	}

	// Skip all whitespaces
	if self.tokenParsers.SkipWhitespaces(self, self) == TOKEN_EOF || self.IsEnd() {
		return self.token(TOKEN_EOF), EofError
	}

	self.lastPos = self.pos - 1

	// Clear the token buffer.
	self.buf.Reset()

	for _, parser := range self.tokenParsers.Parsers {
		if parser != nil {
			tokenType := parser(self, self)
			if tokenType != TOKEN_NULL {
				return self.token(tokenType), nil
			}
		}
	}

	// Any unrecognized runes is treated as an unknown token.
	self.AppendRune()
	self.NextRune()
	return self.token(TOKEN_UNKNOWN), nil
}

func (self *Lexer) PutBack(token Token) {
	self.putbackTokens.PushBack(&token)
}
