package server

import (
	"github.com/emersion/imap/common"
	"github.com/emersion/imap/backend"
	"github.com/emersion/imap/commands"
	"github.com/emersion/imap/responses"
)

type Capability struct {
	commands.Capability
}

func (cmd *Capability) Handle(conn *Conn, bkd backend.Backend) error {
	res := &responses.Capability{
		Caps: conn.getCaps(),
	}

	return res.Response().WriteTo(conn.Writer)
}

type Noop struct {
	commands.Noop
}

func (cmd *Noop) Handle(conn *Conn, bkd backend.Backend) error {
	return nil
}

type Logout struct {
	commands.Logout
}

func (cmd *Logout) Handle(conn *Conn, bkd backend.Backend) error {
	res := &common.StatusResp{
		Tag: "*",
		Type: common.BYE,
		Info: "Closing connection",
	}

	if err := res.WriteTo(conn.Writer); err != nil {
		return err
	}

	// Request to close the connection
	conn.State = common.LogoutState
	return nil
}
