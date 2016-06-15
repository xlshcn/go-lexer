package golexer

import "strconv"
import "errors"

type NumberType int

const (
	NUMBER_TYPE_INTEGER NumberType = iota
	NUMBER_TYPE_FLOAT
)

type NumberValue struct {
	Type    NumberType
	Integer int64
	Float   float64
}

func isOctDigit(r rune) bool {
	return r >= '0' && r <= '7'
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isHexDigit(r rune) bool {
	return r >= '0' && r <= '9' || r >= 'a' && r <= 'f' || r >= 'A' && r <= 'F'
}

var InvalidNumberFormatError = errors.New("Invalid number format.")
var MaximumNumberRangeError = errors.New("The maximum number detected.")

func cNumberParser(scanner IRuneScanner, builder ITokenBuilder) (TokenType, error) {
	r := scanner.Rune()
	if !isDigit(r) {
		return TOKEN_NULL, nil
	}
	base := 10
	if r == '0' {
		r = scanner.NextRune()
		if r == 'x' || r == 'X' {
			// this number is in hexadecimal format
			r = scanner.NextRune()
			if isHexDigit(r) {
				for isHexDigit(r) {
					builder.AppendRune()
					r = scanner.NextRune()
				}
				base = 16
			} else {
				return TOKEN_NULL, InvalidNumberFormatError
			}
		} else {
			// this number is in octal format
			for isOctDigit(r) {
				builder.AppendRune()
				r = scanner.NextRune()
			}
			base = 8
		}
	} else {
		// this number is in decimal format
		for isDigit(r) {
			builder.AppendRune()
			r = scanner.NextRune()
		}
		base = 10
	}

	if r != '.' {
		// this number is integer type
		value, err := strconv.ParseInt(builder.TokenLiteral(), base, 64)
		if err != nil {
			numErr, _ := err.(*strconv.NumError)
			if numErr.Err == strconv.ErrSyntax {
				return TOKEN_NULL, InvalidNumberFormatError
			}
			if numErr.Err == strconv.ErrRange {
				return TOKEN_NULL, MaximumNumberRangeError
			}
		}
		builder.SetValue(NumberValue{
			Type:    NUMBER_TYPE_INTEGER,
			Integer: value,
			Float:   0.0,
		})
		return TOKEN_NUMBER, nil
	} else {
		// this number is float type
		builder.AppendRune()
		r = scanner.NextRune()
		if isDigit(r) {
			for isDigit(r) {
				builder.AppendRune()
				r = scanner.NextRune()
			}
		}
		if r == 'e' || r == 'E' {
			builder.AppendRune()
			r = scanner.NextRune()
			if r == '+' || r == '-' {
				builder.AppendRune()
				r = scanner.NextRune()
			}
			if isDigit(r) {
				for isDigit(r) {
					builder.AppendRune()
					r = scanner.NextRune()
				}
			} else {
				return TOKEN_NULL, InvalidNumberFormatError
			}
		}
		value, err := strconv.ParseFloat(builder.TokenLiteral(), 64)
		if err != nil {
			numErr, _ := err.(*strconv.NumError)
			if numErr.Err == strconv.ErrSyntax {
				return TOKEN_NULL, InvalidNumberFormatError
			}
			if numErr.Err == strconv.ErrRange {
				return TOKEN_NULL, MaximumNumberRangeError
			}
		}
		builder.SetValue(NumberValue{
			Type:    NUMBER_TYPE_FLOAT,
			Integer: 0,
			Float:   value,
		})
		return TOKEN_NUMBER, nil
	}
}

func (self *Token) NumberValue() *NumberValue {
	value, ok := self.Value.(*NumberValue)
	if ok {
		return value
	} else {
		return nil
	}
}
