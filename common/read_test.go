package common_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/emersion/imap/common"
)

func newReader(s string) (b *bytes.Buffer, r *common.Reader) {
	b = bytes.NewBuffer([]byte(s))
	r = common.NewReader(b)
	return
}

func TestReader_ReadSp(t *testing.T) {
	b, r := newReader(" ")

	if err := r.ReadSp(); err != nil {
		t.Error(err)
	}
	if len(b.Bytes()) > 0 {
		t.Error("Buffer is not empty after read")
	}
}

func TestReader_ReadCrlf(t *testing.T) {
	b, r := newReader("\r\n")

	if err := r.ReadCrlf(); err != nil {
		t.Error(err)
	}
	if len(b.Bytes()) > 0 {
		t.Error("Buffer is not empty after read")
	}
}

func TestReader_ReadAtom_Nil(t *testing.T) {
	b, r := newReader("NIL\r\n")

	atom, err := r.ReadAtom()
	if err != nil {
		t.Error(err)
	}
	if atom != nil {
		t.Error("NIL atom is not nil:", atom)
	}
	if err := r.ReadCrlf(); err != nil && err != io.EOF {
		t.Error("Cannot read CRLF after atom:", err)
	}
	if len(b.Bytes()) > 0 {
		t.Error("Buffer is not empty after read")
	}
}

func TestReader_ReadAtom_String(t *testing.T) {
	b, r := newReader("atom\r\n")

	atom, err := r.ReadAtom()
	if err != nil {
		t.Error(err)
	}
	if s, ok := atom.(string); !ok || s != "atom" {
		t.Error("String atom has not the expected value:", atom)
	}
	if err := r.ReadCrlf(); err != nil && err != io.EOF {
		t.Error("Cannot read CRLF after atom:", err)
	}
	if len(b.Bytes()) > 0 {
		t.Error("Buffer is not empty after read")
	}
}

func TestReader_ReadLiteral(t *testing.T) {
	b, r := newReader("{7}\r\nabcdefg")

	literal, err := r.ReadLiteral()
	if err != nil {
		t.Error(err)
	}
	if literal.String() != "abcdefg" {
		t.Error("Literal has not the expected value:", literal.String())
	}
	if len(b.Bytes()) > 0 {
		t.Error("Buffer is not empty after read")
	}
}

func TestReader_ReadQuotedString(t *testing.T) {
	b, r := newReader("\"hello gopher\"\r\n")

	s, err := r.ReadQuotedString()
	if err != nil {
		t.Error(err)
	}
	if s != "hello gopher" {
		t.Error("Quoted string has not the expected value:", s)
	}
	if err := r.ReadCrlf(); err != nil && err != io.EOF {
		t.Error("Cannot read CRLF after quoted string:", err)
	}
	if len(b.Bytes()) > 0 {
		t.Error("Buffer is not empty after read")
	}
}

func TestReader_ReadFields(t *testing.T) {
	b, r := newReader("field1 \"field2\"\r\n")

	fields, err := r.ReadFields()
	if err != nil {
		t.Error(err)
	}
	if len(fields) != 2 {
		t.Error("Excepted 2 fields, but got", len(fields))
	}
	if s, ok := fields[0].(string); !ok || s != "field1" {
		t.Error("Field 1 has not the expected value:", fields[0])
	}
	if s, ok := fields[1].(string); !ok || s != "field2" {
		t.Error("Field 2 has not the expected value:", fields[1])
	}
	if err := r.ReadCrlf(); err != nil && err != io.EOF {
		t.Error("Cannot read CRLF after fields:", err)
	}
	if len(b.Bytes()) > 0 {
		t.Error("Buffer is not empty after read")
	}
}

func TestReader_ReadList(t *testing.T) {
	b, r := newReader("(field1 \"field2\" {6}\r\nfield3 field4)")

	fields, err := r.ReadList()
	if err != nil {
		t.Error(err)
	}
	if len(fields) != 4 {
		t.Error("Excepted 2 fields, but got", len(fields))
	}
	if s, ok := fields[0].(string); !ok || s != "field1" {
		t.Error("Field 1 has not the expected value:", fields[0])
	}
	if s, ok := fields[1].(string); !ok || s != "field2" {
		t.Error("Field 2 has not the expected value:", fields[1])
	}
	if literal, ok := fields[2].(*common.Literal); !ok || literal.String() != "field3" {
		t.Error("Field 3 has not the expected value:", fields[2])
	}
	if s, ok := fields[3].(string); !ok || s != "field4" {
		t.Error("Field 4 has not the expected value:", fields[3])
	}
	if len(b.Bytes()) > 0 {
		t.Error("Buffer is not empty after read")
	}
}

func TestReader_ReadLine(t *testing.T) {
	b, r := newReader("field1 field2\r\n")

	fields, err := r.ReadLine()
	if err != nil {
		t.Error(err)
	}
	if len(fields) != 2 {
		t.Error("Excepted 2 fields, but got", len(fields))
	}
	if s, ok := fields[0].(string); !ok || s != "field1" {
		t.Error("Field 1 has not the expected value:", fields[0])
	}
	if s, ok := fields[1].(string); !ok || s != "field2" {
		t.Error("Field 2 has not the expected value:", fields[1])
	}
	if len(b.Bytes()) > 0 {
		t.Error("Buffer is not empty after read")
	}
}

func TestReader_ReadRespCode(t *testing.T) {
	b, r := newReader("[CAPABILITY NOOP STARTTLS]")

	code, fields, err := r.ReadRespCode()
	if err != nil {
		t.Error(err)
	}
	if code != "CAPABILITY" {
		t.Error("Response code has not the expected value:", code)
	}
	if len(fields) != 2 {
		t.Error("Excepted 2 fields, but got", len(fields))
	}
	if s, ok := fields[0].(string); !ok || s != "NOOP" {
		t.Error("Field 1 has not the expected value:", fields[0])
	}
	if s, ok := fields[1].(string); !ok || s != "STARTTLS" {
		t.Error("Field 2 has not the expected value:", fields[1])
	}
	if len(b.Bytes()) > 0 {
		t.Error("Buffer is not empty after read")
	}
}

func TestReader_ReadInfo(t *testing.T) {
	b, r := newReader("I love potatoes.\r\n")

	info, err := r.ReadInfo()
	if err != nil {
		t.Error(err)
	}
	if info != "I love potatoes." {
		t.Error("Info has not the expected value:", info)
	}
	if len(b.Bytes()) > 0 {
		t.Error("Buffer is not empty after read")
	}
}
