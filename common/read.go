package common

import (
	"errors"
	"io"
	"strconv"
	"strings"
)

const (
	sp = ' '
	cr = '\r'
	lf = '\n'
	dquote = '"'
	literalStart = '{'
	literalEnd = '}'
	listStart = '('
	listEnd = ')'
	respCodeStart = '['
	respCodeEnd = ']'
)

// A string reader.
type StringReader interface {
	// ReadString reads until the first occurrence of delim in the input,
	// returning a string containing the data up to and including the delimiter.
	// See https://golang.org/pkg/bufio/#Reader.ReadString
	ReadString(delim byte) (line string, err error)
}

type reader interface {
	io.Reader
	io.RuneScanner
	StringReader
}

// Convert a field to a number.
func ParseNumber(input interface{}) uint32 {
	str, ok := input.(string)
	if !ok {
		return 0
	}

	nbr, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		panic(err)
	}

	return uint32(nbr)
}

func trimSuffix(str string, suffix rune) string {
	return str[:len(str)-1]
}

// An IMAP reader.
type Reader struct {
	reader

	inRespCode bool
}

func (r *Reader) ReadSp() error {
	char, _, err := r.ReadRune()
	if err != nil {
		return err
	}
	if char != sp {
		return errors.New("Not a space")
	}
	return nil
}

func (r *Reader) ReadCrlf() (err error) {
	var char rune

	if char, _, err = r.ReadRune(); err != nil {
		return
	}
	if char != cr {
		err = errors.New("Line doesn't end with a CR")
	}

	if char, _, err = r.ReadRune(); err != nil {
		return
	}
	if char != lf {
		err = errors.New("Line doesn't end with a LF")
	}

	return
}

func (r *Reader) ReadAtom() (interface{}, error) {
	var atom string
	for {
		char, _, err := r.ReadRune()
		if err != nil {
			return nil, err
		}

		// TODO: list-wildcards and \
		if char == listStart || char == literalStart || char == dquote {
			return nil, errors.New("Atom contains forbidden char: " + string(char))
		}
		if char == sp || char == listEnd || char == cr {
			break
		}
		if char == respCodeEnd && r.inRespCode {
			break
		}

		atom += string(char)
	}

	if atom == "NIL" {
		return nil, nil
	}
	return atom, nil
}

func (r *Reader) ReadLiteral() (literal *Literal, err error) {
	char, _, err := r.ReadRune()
	if err != nil {
		return
	}
	if char != literalStart {
		err = errors.New("Literal string doesn't start with an open brace")
		return
	}

	lstr, err := r.ReadString(byte(literalEnd))
	if err != nil {
		return
	}
	lstr = trimSuffix(lstr, literalEnd)

	l, err := strconv.Atoi(lstr)
	if err != nil {
		return
	}

	if err = r.ReadCrlf(); err != nil {
		return
	}

	b := make([]byte, l)
	if _, err = io.ReadFull(r, b); err != nil {
		return
	}

	literal = NewLiteral(b)
	return
}

func (r *Reader) ReadQuotedString() (str string, err error) {
	char, _, err := r.ReadRune()
	if err != nil {
		return
	}
	if char != dquote {
		err = errors.New("Quoted string doesn't start with a double quote")
		return
	}

	str, err = r.ReadString(byte(dquote))
	if err != nil {
		return
	}
	str = trimSuffix(str, dquote)
	return
}

func (r *Reader) ReadFields() (fields []interface{}, err error) {
	var char rune
	for {
		if char, _, err = r.ReadRune(); err != nil {
			return
		}
		if err = r.UnreadRune(); err != nil {
			return
		}

		var field interface{}
		ok := true
		switch char {
		case literalStart:
			field, err = r.ReadLiteral()
		case dquote:
			field, err = r.ReadQuotedString()
		case listStart:
			field, err = r.ReadList()
		case listEnd:
			ok = false
		default:
			field, err = r.ReadAtom()
			r.UnreadRune()
		}

		if err != nil {
			return
		}
		if ok {
			fields = append(fields, field)
		}

		if char, _, err = r.ReadRune(); err != nil {
			return
		}
		if char == cr || char == listEnd || char == respCodeEnd {
			return
		}
		if char == listStart {
			r.UnreadRune()
			continue
		}
		if char != sp {
			err = errors.New("Fields are not separated by a space")
			return
		}
	}
}

func (r *Reader) ReadList() (fields []interface{}, err error) {
	char, _, err := r.ReadRune()
	if err != nil {
		return
	}
	if char != listStart {
		err = errors.New("List doesn't start with an open parenthesis")
		return
	}

	fields, err = r.ReadFields()
	if err != nil {
		return
	}

	r.UnreadRune()
	if char, _, err = r.ReadRune(); err != nil {
		return
	}
	if char != listEnd {
		err = errors.New("List doesn't end with a close parenthesis")
	}
	return
}

func (r *Reader) ReadLine() (fields []interface{}, err error) {
	fields, err = r.ReadFields()
	if err != nil {
		return
	}

	r.UnreadRune()
	err = r.ReadCrlf()
	return
}

func (r *Reader) ReadRespCode() (code string, fields []interface{}, err error) {
	char, _, err := r.ReadRune()
	if err != nil {
		return
	}
	if char != respCodeStart {
		err = errors.New("Response code doesn't start with an open bracket")
		return
	}

	r.inRespCode = true

	fields, err = r.ReadFields()
	r.inRespCode = false
	if err != nil {
		return
	}

	if len(fields) == 0 {
		err = errors.New("Response code doesn't contain any field")
		return
	}

	code, ok := fields[0].(string)
	if !ok {
		err = errors.New("Response code doesn't start with a string atom")
		return
	}

	fields = fields[1:]

	r.UnreadRune()
	char, _, err = r.ReadRune()
	if err != nil {
		return
	}
	if char != respCodeEnd {
		err = errors.New("Response code doesn't end with a close bracket")
	}
	return
}

func (r *Reader) ReadInfo() (info string, err error) {
	info, err = r.ReadString(byte(cr))
	if err != nil {
		return
	}
	info = strings.TrimSuffix(info, string(cr))
	info = strings.TrimLeft(info, " ")

	var char rune
	if char, _, err = r.ReadRune(); err != nil {
		return
	}
	if char != lf {
		err = errors.New("Line doesn't end with a LF")
	}
	return
}

func NewReader(r reader) *Reader {
	return &Reader{reader: r}
}

type Parser interface {
	Parse(fields []interface{}) error
}
