package common

import (
	"errors"
	"time"
)

// Layouts suitable for passing to time.Parse.
var dateLayouts []string

func init() {
	// Generate layouts based on RFC 5322, section 3.3.
	dows := []string{"", "Mon, "}   // day-of-week
	days := []string{"2", "02"}     // day = 1*2DIGIT
	years := []string{"2006", "06"} // year = 4*DIGIT / 2*DIGIT
	seconds := []string{":05", ""}  // second
	// "-0700 (MST)" is not in RFC 5322, but is common.
	zones := []string{"-0700", "MST", "-0700 (MST)"} // zone = (("+" / "-") 4DIGIT) / "GMT" / ...

	for _, dow := range dows {
		for _, day := range days {
			for _, year := range years {
				for _, second := range seconds {
					for _, zone := range zones {
						s := dow + day + " Jan " + year + " 15:04" + second + " " + zone
						dateLayouts = append(dateLayouts, s)
					}
				}
			}
		}
	}
}

// Parse an IMAP date.
// Borrowed from https://golang.org/pkg/net/mail/#Header.Date
func ParseDate(date string) (*time.Time, error) {
	for _, layout := range dateLayouts {
		t, err := time.Parse(layout, date)
		if err == nil {
			return &t, nil
		}
	}
	return nil, errors.New("Cannot parse date")
}

// A message.
type Message struct {
	// The message identifier. Can be either a sequence number or a UID, depending
	// of how this message has been retrieved.
	Id uint32
	// The message envelope.
	Envelope *Envelope
	// The message body parts.
	Body map[string]*Literal
	// The message body structure.
	BodyStructure *BodyStructure
	// The message flags.
	Flags []string
	// The date the message was received by th server.
	InternalDate *time.Time
	// The message size.
	Size uint32
	// The message UID.
	Uid uint32
}

// Parse a message from fields.
func (m *Message) Parse(fields []interface{}) error {
	var key string
	for i, f := range fields {
		if i % 2 == 0 { // It's a key
			var ok bool
			if key, ok = f.(string); !ok {
				return errors.New("Key is not a string")
			}
		} else { // It's a value
			switch key {
			case "ENVELOPE":
				env, ok := f.([]interface{})
				if !ok {
					return errors.New("ENVELOPE is not a list")
				}

				m.Envelope = &Envelope{}
				if err := m.Envelope.Parse(env); err != nil {
					return err
				}
			case "BODYSTRUCTURE", "BODY":
				bs, ok := f.([]interface{})
				if !ok {
					return errors.New("BODYSTRUCTURE is not a list")
				}

				m.BodyStructure = &BodyStructure{}
				if err := m.BodyStructure.Parse(bs); err != nil {
					return err
				}
			case "FLAGS":
				flags, ok := f.([]interface{})
				if !ok {
					return errors.New("FLAGS is not a list")
				}

				m.Flags = make([]string, len(flags))
				for i, flag := range flags {
					m.Flags[i], _ = flag.(string)
				}
			case "INTERNALDATE":
				date, _ := f.(string)
				m.InternalDate, _ = ParseDate(date)
			case "SIZE":
				m.Size = ParseNumber(f)
			case "UID":
				m.Uid = ParseNumber(f)
			}
		}
	}
	return nil
}

// A message envelope, ie. message metadata from its headers.
type Envelope struct {
	// The message date.
	Date *time.Time
	// The message subject.
	Subject string
	// The From header addresses.
	From []*Address
	// The message senders.
	Sender []*Address
	// The Reply-To header addresses.
	ReplyTo []*Address
	// The To header addresses.
	To []*Address
	// The Cc header addresses.
	Cc []*Address
	// The Bcc header addresses.
	Bcc []*Address
	// The In-Reply-To header. Contains the parent Message-Id.
	InReplyTo string
	// The Message-Id header.
	MessageId string
}

// Parse an envelope from fields.
func (e *Envelope) Parse(fields []interface{}) error {
	if len(fields) < 10 {
		return errors.New("ENVELOPE doesn't contain 10 fields")
	}

	if date, ok := fields[0].(string); ok {
		e.Date, _ = ParseDate(date)
	}
	if subject, ok := fields[1].(string); ok {
		e.Subject = subject
	}
	if from, ok := fields[2].([]interface{}); ok {
		e.From = ParseAddressList(from)
	}
	if sender, ok := fields[3].([]interface{}); ok {
		e.Sender = ParseAddressList(sender)
	}
	if replyTo, ok := fields[4].([]interface{}); ok {
		e.ReplyTo = ParseAddressList(replyTo)
	}
	if to, ok := fields[5].([]interface{}); ok {
		e.To = ParseAddressList(to)
	}
	if cc, ok := fields[6].([]interface{}); ok {
		e.Cc = ParseAddressList(cc)
	}
	if bcc, ok := fields[7].([]interface{}); ok {
		e.Bcc = ParseAddressList(bcc)
	}
	if inReplyTo, ok := fields[8].(string); ok {
		e.InReplyTo = inReplyTo
	}
	if msgId, ok := fields[9].(string); ok {
		e.MessageId = msgId
	}

	return nil
}

// An address.
type Address struct {
	// The personal name.
	PersonalName string
	// The SMTP at-domain-list (source route).
	AtDomainList string
	// The mailbox name.
	MailboxName string
	// The host name.
	HostName string
}

func (addr *Address) String() string {
	return addr.MailboxName + "@" + addr.HostName
}

// Parse an address from fields.
func (addr *Address) Parse(fields []interface{}) error {
	if len(fields) < 4 {
		return errors.New("Address doesn't contain 4 fields")
	}

	if f, ok := fields[0].(string); ok {
		addr.PersonalName = f
	}
	if f, ok := fields[1].(string); ok {
		addr.AtDomainList = f
	}
	if f, ok := fields[2].(string); ok {
		addr.MailboxName = f
	}
	if f, ok := fields[3].(string); ok {
		addr.HostName = f
	}

	return nil
}

// Parse an address list from fields.
func ParseAddressList(fields []interface{}) (addrs []*Address) {
	for _, f := range fields {
		if addrFields, ok := f.([]interface{}); ok {
			addr := &Address{}
			if err := addr.Parse(addrFields); err == nil {
				addrs = append(addrs, addr)
			}
		}
	}
	return
}

// A body structure.
type BodyStructure struct {
	// The MIME type.
	MimeType string
	// The MIME subtype.
	MimeSubType string
	// The MIME parameters.
	Params map[string]string

	// The Content-Id header.
	Id string
	// The Content-Description header.
	Description string
	// The Content-Encoding header.
	Encoding string
	// The Content-Length header.
	Size uint32

	// The children parts, if multipart.
	Parts []*BodyStructure
}

// (("text" "plain" ("charset" "UTF-8") NIL NIL "quoted-printable" 2215 74)("text" "html" ("charset" "UTF-8") NIL NIL "quoted-printable" 4062 53) "alternative")

func (bs *BodyStructure) Parse(fields []interface{}) error {
	if len(fields) == 0 {
		return nil
	}

	switch fields[0].(type) {
	case []interface{}: // A multipart body part
		bs.MimeType = "multipart"

		end := 0
		for i, fi := range fields {
			switch f := fi.(type) {
			case []interface{}: // A part
				part := &BodyStructure{}
				if err := part.Parse(f); err != nil {
					return err
				}
				bs.Parts = append(bs.Parts, part)
			case string:
				end = i
			}

			if end > 0 {
				break
			}
		}

		bs.MimeSubType, _ = fields[end].(string)

		// TODO: extension data
	case string: // A non-multipart body part
		if len(fields) < 7 {
			return errors.New("Non-multipart body part doesn't have 7 fields")
		}

		bs.MimeType, _ = fields[0].(string)
		bs.MimeSubType, _ = fields[1].(string)

		params, _ := fields[2].([]interface{})
		bs.Params = map[string]string{}

		var key string
		for i, f := range params {
			p, ok := f.(string)
			if !ok {
				return errors.New("Parameter list contains a non-string")
			}

			if i % 2 == 0 {
				key = p
			} else {
				bs.Params[key] = p
			}
		}

		bs.Id, _ = fields[3].(string)
		bs.Description, _ = fields[4].(string)
		bs.Encoding, _ = fields[5].(string)
		bs.Size = ParseNumber(fields[6])

		// TODO: extension data
	}

	return nil
}
