package golexer

import "io"
import "errors"
import "unicode"
import "bytes"

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
	LinPos     int
}

type Categorizer func(r rune) bool

type Categorizers struct {
	IsWhitespace    func(r rune) bool
	IsIdentifier    func(r rune, pos int) bool
	IsNumber        func(r rune, pos int) bool
	IsQuotationMark func(lastMark rune, r rune) (mark rune, result bool)
}

func NewCategorizers() *Categorizers {
	return &Categorizers{
		IsWhitespace:     unicode.IsSpace,
		isIdentifierLead: isIdentifierLead,
		IsIdentifier:     isIdentifier,
		IsNumber:         unicode.IsDigit,
		IsQuotationMark:  isQuotationMark,
		IsComment:        isComment,
	}
}

func isIdentifier(r rune, pos int) bool {
	if pos == 0 {
		return unicode.IsLetter(r) || r == '_'
	} else {
		return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
	}
}

func isNumber(r rune, pos int) bool {
	return unicode.IsDigit(r)
}

func isQuotationMark(lastMark rune, r rune) (mark rune, result bool) {
	if r == '"' || r == '\'' {
		if lastMark == 0 || lastMark == r {
			return r, true
		}
	}
	return 0, false
}

func isComment(r rune) bool {
	return r == '#'
}

type Lexer struct {
	categorizers Categorizers
	scanner      io.RuneScanner
	lineno       int
	pos          int
	r            rune
	eof          bool
	buf          bytes.Buffer
	lastToken    Token
}

func NewLexer(scanner io.RuneScanner, categorizers *Categorizers) (*Lexer, error) {
	if scanner == nil {
		return nil, NullArgumentError
	}
	if categorizers == nil {
		categorizers = NewCategorizers()
	}

	lexer := new(Lexer)
	lexer.categorizers = categorizers
	lexer.scanner = scanner
	lexer.buf = new(bytes.Buffer)

	return lexer
}

func (self *Lexer) nextRune() (validRune bool) {
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
	return !self.eof
}

func (self *Lexer) appendRune() {
	self.buf.WriteRune(self.r)
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

func (self *Lexer) isWhitespace() bool {
	if self.eof {
		return false
	}
	if self.r == 0 {
		return true
	}
	if self.categorizers.IsWhitespace != nil {
		return self.categorizers.IsWhitespace(self.r)
	}
	return false
}

func (self *Lexer) isIdentifierLead() bool {
	if self.eof {
		return false
	}
	if self.categorizers.IsIdentifierLead != nil {
		return self.categorizers.IsIdentifierLead(self.r)
	}
	return false
}

func (self *Lexer) isIdentifier() bool {
	if self.eof {
		return false
	}
	if self.categorizers.IsIdentifier != nil {
		return self.categorizers.IsIdentifier(self.r)
	}
	return false
}

func (self *Lexer) isNumber() bool {
	if self.eof {
		return false
	}
	if self.categorizers.IsNumber != nil {
		return self.categorizers.IsNumber(self.r)
	}
	return false
}

func (self *Lexer) isQuotationMark(lastMark rune) (mark rune, result bool) {
	if self.eof {
		return 0, false
	}
	if self.categorizers.IsQuotationMark != nil {
		return self.categorizers.IsQuotationMark(lastMark, self.r)
	}
	return 0, false
}

func (self *Lexer) IsEnd() bool {
	return self.eof
}

func (self *Lexer) GetToken() (Token, error) {
	for self.isWhitespace() && self.nextRune() {
	}

	self.buf.Reset()

	if self.isIdentifierLead() {
		for self.isIdentifier() {
			self.appendRune()
			self.nextRune()
		}
		return self.token(TOKEN_IDENTIFIER)
	}

	if self.isNumber() {
		for self.isNumber() {
			self.appendRune()
			self.nextRune()
		}
		return self.token(TOKEN_NUMBER)
	}

	lastMark, result := self.isQuotationMark(0)
	if result {
		for self.nextRune() {
			r, result := self.isQuotationMark(lastMark)
			if result {
				break
			}
			self.appendRune()
		}
		return self.token(TOKEN_STRING)
	}

	self.appendRune()
	return self.token(TOKEN_UNKNOWN)
}
