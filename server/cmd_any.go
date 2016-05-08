package server

import (
	imap "github.com/emersion/imap/common"
	"github.com/emersion/imap/commands"
	"github.com/emersion/imap/responses"
)

type Capability struct {
	commands.Capability
}

func (cmd *Capability) Handle(conn *Conn, bkd Backend) error {
	res := &responses.Capability{
		Caps: conn.getCaps(),
	}

	return res.Response().WriteTo(conn.Writer)
}

type Noop struct {
	commands.Noop
}

func (cmd *Noop) Handle(conn *Conn, bkd Backend) error {
	return nil
}

type Logout struct {
	commands.Logout
}

func (cmd *Logout) Handle(conn *Conn, bkd Backend) error {
	res := &imap.StatusResp{
		Tag: "*",
		Type: imap.BYE,
		Info: "Closing connection",
	}

	if err := res.WriteTo(conn.Writer); err != nil {
		return err
	}

	// Request to close the connection
	conn.State = imap.LogoutState
	return nil
}
