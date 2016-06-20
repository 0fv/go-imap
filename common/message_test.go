package common_test

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/emersion/go-imap/common"
)

func TestParseDate(t *testing.T) {
	tests := []struct{
		dateStr string
		exp     time.Time
	}{
		{
			"21-Nov-1997 09:55:06 -0600",
			time.Date(1997, 11, 21, 9, 55, 6, 0, time.FixedZone("", -6*60*60)),
		},
	}
	for _, test := range tests {
		date, err := common.ParseDate(test.dateStr)
		if err != nil {
			t.Errorf("Failed parsing %q: %v", test.dateStr, err)
			continue
		}
		if !date.Equal(test.exp) {
			t.Errorf("Parse of %q: got %+v, want %+v", test.dateStr, date, test.exp)
		}
	}
}

func TestNewBodySectionName(t *testing.T) {
	tests := []struct{
		raw string
		parsed *common.BodySectionName
	}{
		{
			raw: "BODY[]",
			parsed: &common.BodySectionName{BodyPartName: &common.BodyPartName{}},
		},
		{
			raw: "RFC822",
			parsed: &common.BodySectionName{BodyPartName: &common.BodyPartName{}},
		},
		{
			raw: "BODY[HEADER]",
			parsed: &common.BodySectionName{BodyPartName: &common.BodyPartName{Specifier: common.HeaderSpecifier}},
		},
		{
			raw: "BODY.PEEK[]",
			parsed: &common.BodySectionName{BodyPartName: &common.BodyPartName{}, Peek: true},
		},
		{
			raw: "BODY[TEXT]",
			parsed: &common.BodySectionName{BodyPartName: &common.BodyPartName{Specifier: common.TextSpecifier}},
		},
		{
			raw: "RFC822.HEADER",
			parsed: &common.BodySectionName{BodyPartName: &common.BodyPartName{Specifier: common.HeaderSpecifier}, Peek: true},
		},
		{
			raw: "BODY[]<0.512>",
			parsed: &common.BodySectionName{BodyPartName: &common.BodyPartName{}, Partial: []int{0, 512}},
		},
		{
			raw: "BODY[1.2.3]",
			parsed: &common.BodySectionName{BodyPartName: &common.BodyPartName{Path: []int{1, 2, 3}}},
		},
		{
			raw: "BODY[1.2.3.HEADER]",
			parsed: &common.BodySectionName{BodyPartName: &common.BodyPartName{Specifier: common.HeaderSpecifier, Path: []int{1, 2, 3}}},
		},
		{
			raw: "BODY[5.MIME]",
			parsed: &common.BodySectionName{BodyPartName: &common.BodyPartName{Specifier: common.MimeSpecifier, Path: []int{5}}},
		},
		{
			raw: "BODY[HEADER.FIELDS (From To)]",
			parsed: &common.BodySectionName{BodyPartName: &common.BodyPartName{Specifier: common.HeaderSpecifier, Fields: []string{"From", "To"}}},
		},
		{
			raw: "BODY[HEADER.FIELDS.NOT (Content-Id)]",
			parsed: &common.BodySectionName{BodyPartName: &common.BodyPartName{Specifier: common.HeaderSpecifier, Fields: []string{"Content-Id"}, NotFields: true}},
		},
	}

	for i, test := range tests {
		bsn, err := common.NewBodySectionName(test.raw)
		if err != nil {
			t.Errorf("Cannot parse #%v: %v", i, err)
			continue
		}

		if !reflect.DeepEqual(bsn.BodyPartName, test.parsed.BodyPartName) {
			t.Errorf("Invalid body part name for #%v: %#+v", i, bsn.BodyPartName)
		} else if bsn.Peek != test.parsed.Peek {
			t.Errorf("Invalid peek value for #%v: %#+v", i, bsn.Peek)
		} else if !reflect.DeepEqual(bsn.Partial, test.parsed.Partial) {
			t.Errorf("Invalid partial for #%v: %#+v", i, bsn.Partial)
		}
	}
}

var bodyStructureTests = []struct{
	fields []interface{}
	bodyStructure *common.BodyStructure
}{
	{
		fields: []interface{}{"image", "jpeg", []interface{}{}, "<foo4%25foo1@bar.net>", "A picture of cat", "base64", "4242"},
		bodyStructure: &common.BodyStructure{
			MimeType: "image",
			MimeSubType: "jpeg",
			Params: map[string]string{},
			Id: "<foo4%25foo1@bar.net>",
			Description: "A picture of cat",
			Encoding: "base64",
			Size: 4242,
		},
	},
	{
		fields: []interface{}{"text", "plain", []interface{}{"charset", "utf-8"}, nil, nil, "us-ascii", "42", "2"},
		bodyStructure: &common.BodyStructure{
			MimeType: "text",
			MimeSubType: "plain",
			Params: map[string]string{"charset": "utf-8"},
			Encoding: "us-ascii",
			Size: 42,
			Lines: 2,
		},
	},
	{
		fields: []interface{}{
			"message", "rfc822", []interface{}{}, nil, nil, "us-ascii", "42",
			(&common.Envelope{}).Format(),
			(&common.BodyStructure{}).Format(),
			"67",
		},
		bodyStructure: &common.BodyStructure{
			MimeType: "message",
			MimeSubType: "rfc822",
			Params: map[string]string{},
			Encoding: "us-ascii",
			Size: 42,
			Lines: 67,
			Envelope: &common.Envelope{
				From: []*common.Address{},
				Sender: []*common.Address{},
				ReplyTo: []*common.Address{},
				To: []*common.Address{},
				Cc: []*common.Address{},
				Bcc: []*common.Address{},
			},
			BodyStructure: &common.BodyStructure{
				Params: map[string]string{},
			},
		},
	},
	{
		fields: []interface{}{"application", "pdf", []interface{}{}, nil, nil, "base64", "4242",
			"e0323a9039add2978bf5b49550572c7c", "attachment", []interface{}{"en-US"}, []interface{}{}},
		bodyStructure: &common.BodyStructure{
			MimeType: "application",
			MimeSubType: "pdf",
			Params: map[string]string{},
			Encoding: "base64",
			Size: 4242,
			Extended: true,
			Md5: "e0323a9039add2978bf5b49550572c7c",
			Disposition: "attachment",
			Language: []string{"en-US"},
			Location: []string{},
		},
	},
	{
		fields: []interface{}{
			[]interface{}{"text", "plain", []interface{}{}, nil, nil, "us-ascii", "87", "22"},
			[]interface{}{"text", "html", []interface{}{}, nil, nil, "us-ascii", "106", "36"},
			"alternative",
		},
		bodyStructure: &common.BodyStructure{
			MimeType: "multipart",
			MimeSubType: "alternative",
			Params: map[string]string{},
			Parts: []*common.BodyStructure{
				&common.BodyStructure{
					MimeType: "text",
					MimeSubType: "plain",
					Params: map[string]string{},
					Encoding: "us-ascii",
					Size: 87,
					Lines: 22,
				},
				&common.BodyStructure{
					MimeType: "text",
					MimeSubType: "html",
					Params: map[string]string{},
					Encoding: "us-ascii",
					Size: 106,
					Lines: 36,
				},
			},
		},
	},
	{
		fields: []interface{}{
			[]interface{}{"text", "plain", []interface{}{}, nil, nil, "us-ascii", "87", "22"},
			"alternative", []interface{}{"hello", "world"}, "inline", []interface{}{"en-US"}, []interface{}{},
		},
		bodyStructure: &common.BodyStructure{
			MimeType: "multipart",
			MimeSubType: "alternative",
			Params: map[string]string{"hello": "world"},
			Parts: []*common.BodyStructure{
				&common.BodyStructure{
					MimeType: "text",
					MimeSubType: "plain",
					Params: map[string]string{},
					Encoding: "us-ascii",
					Size: 87,
					Lines: 22,
				},
			},
			Extended: true,
			Disposition: "inline",
			Language: []string{"en-US"},
			Location: []string{},
		},
	},
}

func TestBodyStructure_Parse(t *testing.T) {
	for i, test := range bodyStructureTests {
		bs := &common.BodyStructure{}

		if err := bs.Parse(test.fields); err != nil {
			t.Errorf("Cannot parse #%v: %v", i, err)
		} else if !reflect.DeepEqual(bs, test.bodyStructure) {
			t.Errorf("Invalid body structure for #%v: got %v but expected %v", i, bs, test.bodyStructure)
		}
	}
}

func TestBodyStructure_Format(t *testing.T) {
	b := &bytes.Buffer{}
	w := common.NewWriter(b)

	formatFields := func(fields []interface{}) string {
		if _, err := w.WriteList(fields); err != nil {
			t.Fatalf("Cannot format %v: %v", fields, err)
		}

		s := b.String()
		b.Reset()
		return s
	}

	for i, test := range bodyStructureTests {
		fields := test.bodyStructure.Format()
		got := formatFields(fields)

		expected := formatFields(test.fields)

		if got != expected {
			t.Errorf("Invalid body structure for #%v: has %v but expected %v", i, got, expected)
		}
	}
}

func TestAddress_Parse(t *testing.T) {
	addr := &common.Address{}
	fields := []interface{}{"The NSA", nil, "root", "nsa.gov"}

	if err := addr.Parse(fields); err != nil {
		t.Fatal(err)
	}

	if addr.PersonalName != "The NSA" {
		t.Error("Invalid personal name:", addr.PersonalName)
	}
	if addr.AtDomainList != "" {
		t.Error("Invalid at-domain-list:", addr.AtDomainList)
	}
	if addr.MailboxName != "root" {
		t.Error("Invalid mailbox name:", addr.MailboxName)
	}
	if addr.HostName != "nsa.gov" {
		t.Error("Invalid host name:", addr.HostName)
	}
}

func TestAddress_Format(t *testing.T) {
	addr := &common.Address{
		PersonalName: "The NSA",
		MailboxName: "root",
		HostName: "nsa.gov",
	}

	fields := addr.Format()
	if len(fields) != 4 {
		t.Fatal("Invalid fields list length")
	}
	if fields[0] != "The NSA" {
		t.Error("Invalid personal name:", fields[0])
	}
	if fields[1] != nil {
		t.Error("Invalid at-domain-list:", fields[1])
	}
	if fields[2] != "root" {
		t.Error("Invalid mailbox name:", fields[2])
	}
	if fields[3] != "nsa.gov" {
		t.Error("Invalid host name:", fields[3])
	}
}
