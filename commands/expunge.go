package commands

import (
	imap "github.com/emersion/imap/common"
)

// An EXPUNGE command.
// See https://tools.ietf.org/html/rfc3501#section-6.4.3
type Expunge struct {}

func (cmd *Expunge) Command() *imap.Command {
	return &imap.Command{Name: imap.Expunge}
}

func (cmd *Expunge) Parse(fields []interface{}) error {
	return nil
}
