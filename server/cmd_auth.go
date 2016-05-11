package server

import (
	"errors"

	"github.com/emersion/imap/common"
	"github.com/emersion/imap/commands"
	"github.com/emersion/imap/responses"
)

type Select struct {
	commands.Select
}

func (cmd *Select) Handle(conn *Conn) error {
	if conn.User == nil {
		return errors.New("Not authenticated")
	}

	mbox, err := conn.User.GetMailbox(cmd.Mailbox)
	if err != nil {
		return err
	}

	items := []string{"MESSAGES", "RECENT", "UNSEEN", "UIDNEXT", "UIDVALIDITY"}
	status, err := mbox.Status(items)
	if err != nil {
		return err
	}

	conn.Mailbox = mbox
	conn.MailboxReadOnly = cmd.ReadOnly || status.ReadOnly

	flags := make([]interface{}, len(status.Flags))
	for i, f := range status.Flags {
		flags[i] = f
	}
	res := common.NewUntaggedResp([]interface{}{"FLAGS", flags})
	if err := res.WriteTo(conn.Writer); err != nil {
		return err
	}

	res = common.NewUntaggedResp([]interface{}{status.Messages, "EXISTS"})
	if err := res.WriteTo(conn.Writer); err != nil {
		return err
	}

	res = common.NewUntaggedResp([]interface{}{status.Recent, "RECENT"})
	if err := res.WriteTo(conn.Writer); err != nil {
		return err
	}

	statusRes := &common.StatusResp{
		Tag: "*",
		Type: common.OK,
		Code: "UNSEEN",
		Arguments: []interface{}{status.Unseen},
	}
	if err := statusRes.WriteTo(conn.Writer); err != nil {
		return err
	}

	flags = make([]interface{}, len(status.PermanentFlags))
	for i, f := range status.PermanentFlags {
		flags[i] = f
	}
	statusRes = &common.StatusResp{
		Tag: "*",
		Type: common.OK,
		Code: "PERMANENTFLAGS",
		Arguments: []interface{}{flags},
		Info: "Flags permitted.",
	}
	if err := statusRes.WriteTo(conn.Writer); err != nil {
		return err
	}

	statusRes = &common.StatusResp{
		Tag: "*",
		Type: common.OK,
		Code: "UIDNEXT",
		Arguments: []interface{}{status.UidNext},
		Info: "Predicted next UID",
	}
	if err := statusRes.WriteTo(conn.Writer); err != nil {
		return err
	}

	statusRes = &common.StatusResp{
		Tag: "*",
		Type: common.OK,
		Code: "UIDVALIDITY",
		Arguments: []interface{}{status.UidValidity},
		Info: "UIDs valid",
	}
	if err := statusRes.WriteTo(conn.Writer); err != nil {
		return err
	}

	return nil
}

type Create struct {
	commands.Create
}

func (cmd *Create) Handle(conn *Conn) error {
	if conn.User == nil {
		return errors.New("Not authenticated")
	}

	return conn.User.CreateMailbox(cmd.Mailbox)
}

type Delete struct {
	commands.Delete
}

func (cmd *Delete) Handle(conn *Conn) error {
	if conn.User == nil {
		return errors.New("Not authenticated")
	}

	return conn.User.DeleteMailbox(cmd.Mailbox)
}

type Rename struct {
	commands.Rename
}

func (cmd *Rename) Handle(conn *Conn) error {
	if conn.User == nil {
		return errors.New("Not authenticated")
	}

	return conn.User.RenameMailbox(cmd.Existing, cmd.New)
}

type Subscribe struct {
	commands.Subscribe
}

func (cmd *Subscribe) Handle(conn *Conn) error {
	if conn.User == nil {
		return errors.New("Not authenticated")
	}

	mbox, err := conn.User.GetMailbox(cmd.Mailbox)
	if err != nil {
		return err
	}

	return mbox.Subscribe()
}

type Unsubscribe struct {
	commands.Unsubscribe
}

func (cmd *Unsubscribe) Handle(conn *Conn) error {
	if conn.User == nil {
		return errors.New("Not authenticated")
	}

	mbox, err := conn.User.GetMailbox(cmd.Mailbox)
	if err != nil {
		return err
	}

	return mbox.Unsubscribe()
}

type List struct {
	commands.List
}

func (cmd *List) Handle(conn *Conn) error {
	if conn.User == nil {
		return errors.New("Not authenticated")
	}

	done := make(chan error)
	defer close(done)

	ch := make(chan *common.MailboxInfo)
	res := responses.List{Mailboxes: ch, Subscribed: cmd.Subscribed}

	go (func () {
		done <- res.WriteTo(conn.Writer)
	})()

	mailboxes, err := conn.User.ListMailboxes(cmd.Subscribed)
	if err != nil {
		close(ch)
		return err
	}

	for _, mbox := range mailboxes {
		info, err := mbox.Info()
		if err != nil {
			close(ch)
			return err
		}

		// TODO: filter mailboxes with cmd.Reference and cmd.Mailbox
		if cmd.Reference != "" || (cmd.Mailbox != "*" && cmd.Mailbox != "%" && cmd.Mailbox != info.Name) {
			continue
		}

		ch <- info
	}

	close(ch)

	return <-done
}

type Status struct {
	commands.Status
}

func (cmd *Status) Handle(conn *Conn) error {
	if conn.User == nil {
		return errors.New("Not authenticated")
	}

	mbox, err := conn.User.GetMailbox(cmd.Mailbox)
	if err != nil {
		return err
	}

	status, err := mbox.Status(cmd.Items)
	if err != nil {
		return err
	}

	res := responses.Status{Mailbox: status}
	return res.WriteTo(conn.Writer)
}

type Append struct {
	commands.Append
}

func (cmd *Append) Handle(conn *Conn) error {
	if conn.User == nil {
		return errors.New("Not authenticated")
	}

	mbox, err := conn.User.GetMailbox(cmd.Mailbox)
	if err != nil {
		// TODO: add [TRYCREATE] to the NO response
		return err
	}

	return mbox.CreateMessage(cmd.Flags, cmd.Date, cmd.Message.Bytes())
}
