package client

import (
	"errors"

	imap "github.com/emersion/imap/common"
	"github.com/emersion/imap/commands"
	"github.com/emersion/imap/responses"
)

func (c *Client) Select(name string, readOnly bool) (mbox *imap.MailboxStatus, err error) {
	if c.State != imap.AuthenticatedState && c.State != imap.SelectedState {
		err = errors.New("Not logged in")
		return
	}

	mbox = &imap.MailboxStatus{Name: name}

	cmd := &commands.Select{
		Mailbox: name,
		ReadOnly: readOnly,
	}
	res := &responses.Select{
		Mailbox: mbox,
	}

	status, err := c.execute(cmd, res)
	if err != nil {
		return
	}

	err = status.Err()
	if err != nil {
		return
	}

	mbox.ReadOnly = (status.Code == "READ-ONLY")

	c.State = imap.SelectedState
	c.Selected = mbox
	return
}

// TODO: CREATE, DELETE, RENAME, SUBSCRIBE, UNSUBSCRIBE

func (c *Client) List(ref, mbox string, ch chan<- *imap.MailboxInfo) (err error) {
	defer close(ch)

	if c.State != imap.AuthenticatedState && c.State != imap.SelectedState {
		err = errors.New("Not logged in")
		return
	}

	cmd := &commands.List{
		Reference: ref,
		Mailbox: mbox,
	}
	res := &responses.List{Mailboxes: ch}

	status, err := c.execute(cmd, res)
	if err != nil {
		return
	}

	err = status.Err()
	return
}

// TODO: LSUB

func (c *Client) Status(name string, items []string) (mbox *imap.MailboxStatus, err error) {
	if c.State != imap.AuthenticatedState && c.State != imap.SelectedState {
		err = errors.New("Not logged in")
		return
	}

	mbox = &imap.MailboxStatus{}

	cmd := &commands.Status{
		Mailbox: name,
		Items: items,
	}
	res := &responses.Status{
		Mailbox: mbox,
	}

	status, err := c.execute(cmd, res)
	if err != nil {
		return
	}

	err = status.Err()
	return
}

// TODO: APPEND
