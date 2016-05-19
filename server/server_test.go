package server_test

import (
	"bufio"
	"net"
	"testing"

	"github.com/emersion/go-imap/backend/memory"
	"github.com/emersion/go-imap/server"
)

func testServer(t *testing.T) (s *server.Server, conn net.Conn) {
	bkd := memory.New()

	s, err := server.Listen("127.0.0.1:0", bkd)
	if err != nil {
		t.Fatal("Cannot start server:", err)
	}

	s.AllowInsecureAuth = true

	conn, err = net.Dial("tcp", s.Addr().String())
	if err != nil {
		t.Fatal("Cannot connect to server:", err)
	}

	return
}

func TestServer_greeting(t *testing.T) {
	s, conn := testServer(t)
	defer conn.Close()
	defer s.Close()

	scanner := bufio.NewScanner(conn)

	scanner.Scan() // Wait for greeting
	greeting := scanner.Text()

	if greeting != "* OK [CAPABILITY IMAP4rev1 AUTH=PLAIN] IMAP4rev1 Service Ready" {
		t.Fatal("Bad greeting:", greeting)
	}
}
