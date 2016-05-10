package server

import (
	"crypto/tls"
	"errors"

	"github.com/emersion/imap/common"
	"github.com/emersion/imap/commands"
	"github.com/emersion/imap/sasl"
)

type StartTLS struct {
	commands.StartTLS
}

func (cmd *StartTLS) Handle(conn *Conn) error {
	if conn.State != common.NotAuthenticatedState {
		return errors.New("Already authenticated")
	}
	if conn.IsTLS() {
		return errors.New("TLS is already enabled")
	}
	if conn.Server.TLSConfig == nil {
		return errors.New("TLS support not enabled")
	}

	upgraded := tls.Server(conn.conn, conn.Server.TLSConfig)

	if err := upgraded.Handshake(); err != nil {
		return err
	}

	conn.conn = upgraded
	return nil
}

type Login struct {
	commands.Login
}

func (cmd *Login) Handle(conn *Conn) error {
	if conn.State != common.NotAuthenticatedState {
		return errors.New("Already authenticated")
	}
	if !conn.CanAuth() {
		return errors.New("Authentication disabled")
	}

	user, err := conn.Server.Backend.Login(cmd.Username, cmd.Password)
	if err != nil {
		return err
	}

	conn.State = common.AuthenticatedState
	conn.User = user
	return nil
}

type Authenticate struct {
	commands.Authenticate

	Mechanisms map[string]sasl.Server
}

func (cmd *Authenticate) Handle(conn *Conn) error {
	if conn.State != common.NotAuthenticatedState {
		return errors.New("Already authenticated")
	}
	if !conn.CanAuth() {
		return errors.New("Authentication disabled")
	}

	user, err := cmd.Authenticate.Handle(cmd.Mechanisms, conn.Reader, conn.Writer)
	if err != nil {
		return err
	}

	conn.State = common.AuthenticatedState
	conn.User = user
	return nil
}
