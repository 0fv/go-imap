package client_test

import (
	"io"
	"net"
	"testing"

	"github.com/emersion/imap/client"
)

func TestClient_Capability(t *testing.T) {
	ct := func(c *client.Client) {
		caps, err := c.Capability()
		if err != nil {
			t.Fatal(err)
		}

		if !caps["XTEST"] {
			t.Fatal("Client hasn't advertised capability")
		}
	}

	st := func(c net.Conn) {
		scanner := NewCmdScanner(c)

		tag, cmd := scanner.Scan()
		if cmd != "CAPABILITY" {
			t.Fatal("Bad command:", cmd)
		}

		io.WriteString(c, "* CAPABILITY IMAP4rev1 XTEST\r\n")
		io.WriteString(c, tag + " OK CAPABILITY completed.\r\n")
	}

	testClient(t, ct, st)
}
