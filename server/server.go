// An IMAP server.
package server

import (
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"

	"github.com/emersion/go-imap/common"
	"github.com/emersion/go-imap/backend"
	"github.com/emersion/go-sasl"
)

// A command handler.
type Handler interface {
	common.Parser

	// Handle this command for a given connection.
	Handle(conn *Conn) error
}

// A function that creates handlers.
type HandlerFactory func() Handler

// A function that creates SASL servers.
type SaslServerFactory func(conn *Conn) sasl.Server

// An IMAP server.
type Server struct {
	listener net.Listener
	conns []*Conn

	caps map[string]common.ConnState
	commands map[string]HandlerFactory
	auths map[string]SaslServerFactory

	// This server's backend.
	Backend backend.Backend
	// This server's TLS configuration.
	TLSConfig *tls.Config
	// Allow authentication over unencrypted connections.
	AllowInsecureAuth bool
	// Print all network activity to STDOUT.
	Debug bool
}

// Get this server's address.
func (s *Server) Addr() net.Addr {
	return s.listener.Addr()
}

func (s *Server) listen() error {
	defer s.listener.Close()

	for {
		c, err := s.listener.Accept()
		if err != nil {
			return err
		}

		conn := newConn(s, c)
		if s.Debug {
			conn.SetDebug(true)
		}

		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn *Conn) error {
	s.conns = append(s.conns, conn)
	defer conn.Close()

	// Send greeting
	if err := conn.greet(); err != nil {
		return err
	}

	for {
		if conn.State == common.LogoutState {
			return nil
		}

		conn.Wait()

		fields, err := conn.ReadLine()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			if conn.State != common.LogoutState {
				log.Println("Error reading command:", err)
			}
			return err
		}

		var res common.WriterTo

		cmd := &common.Command{}
		if err := cmd.Parse(fields); err != nil {
			res = &common.StatusResp{
				Tag: "*",
				Type: common.BAD,
				Info: err.Error(),
			}
		} else {
			var err error
			res, err = s.handleCommand(cmd, conn)
			if err != nil {
				res = &common.StatusResp{
					Tag: cmd.Tag,
					Type: common.BAD,
					Info: err.Error(),
				}
			}
		}

		if err := conn.WriteRes(res); err != nil {
			log.Println("Error writing response:", err)
			continue
		}
	}
}

func (s *Server) getCommandHandler(cmd *common.Command) (hdlr Handler, err error) {
	newHandler, ok := s.commands[cmd.Name]
	if !ok {
		err = errors.New("Unknown command")
		return
	}

	hdlr = newHandler()
	err = hdlr.Parse(cmd.Arguments)
	return
}

func (s *Server) handleCommand(cmd *common.Command, conn *Conn) (res common.WriterTo, err error) {
	hdlr, err := s.getCommandHandler(cmd)
	if err != nil {
		return
	}

	if err := hdlr.Handle(conn); err != nil {
		res = &common.StatusResp{
			Tag: cmd.Tag,
			Type: common.NO,
			Info: err.Error(),
		}
	} else {
		res = &common.StatusResp{
			Tag: cmd.Tag,
			Type: common.OK,
			Info: cmd.Name + " completed",
		}
	}

	return
}

func (s *Server) getCaps(currentState common.ConnState) (caps []string) {
	for name, state := range s.caps {
		if currentState & state != 0 {
			caps = append(caps, name)
		}
	}
	return
}

// Stops listening and closes all current connections.
func (s *Server) Close() error {
	if err := s.listener.Close(); err != nil {
		return err
	}

	for _, conn := range s.conns {
		conn.Close()
	}

	return nil
}

// Create a new IMAP server from an existing listener.
func NewServer(l net.Listener, bkd backend.Backend) *Server {
	s := &Server{
		listener: l,
		caps: map[string]common.ConnState{},
		Backend: bkd,
	}

	s.auths = map[string]SaslServerFactory{
		"PLAIN": func(conn *Conn) sasl.Server {
			return sasl.NewPlainServer(func(identity, username, password string) error {
				if identity != "" && identity != username {
					return errors.New("Identities not supported")
				}

				user, err := bkd.Login(username, password)
				if err != nil {
					return err
				}

				conn.State = common.AuthenticatedState
				conn.User = user
				return nil
			})
		},
	}

	s.commands = map[string]HandlerFactory{
		common.Noop: func() Handler { return &Noop{} },
		common.Capability: func() Handler { return &Capability{} },
		common.Logout: func() Handler { return &Logout{} },

		common.StartTLS: func() Handler { return &StartTLS{} },
		common.Login: func() Handler { return &Login{} },
		common.Authenticate: func() Handler { return &Authenticate{} },

		common.Select: func() Handler { return &Select{} },
		common.Examine: func() Handler {
			hdlr := &Select{}
			hdlr.ReadOnly = true
			return hdlr
		},
		common.Create: func() Handler { return &Create{} },
		common.Delete: func() Handler { return &Delete{} },
		common.Rename: func() Handler { return &Rename{} },
		common.Subscribe: func() Handler { return &Subscribe{} },
		common.Unsubscribe: func() Handler { return &Unsubscribe{} },
		common.List: func() Handler { return &List{} },
		common.Lsub: func() Handler {
			hdlr := &List{}
			hdlr.Subscribed = true
			return hdlr
		},
		common.Status: func() Handler { return &Status{} },
		common.Append: func() Handler { return &Append{} },

		common.Check: func() Handler { return &Check{} },
		common.Close: func() Handler { return &Close{} },
		common.Expunge: func() Handler { return &Expunge{} },
		common.Search: func() Handler { return &Search{} },
		common.Fetch: func() Handler { return &Fetch{} },
		common.Store: func() Handler { return &Store{} },
		common.Copy: func() Handler { return &Copy{} },
		common.Uid: func() Handler { return &Uid{} },
	}

	go s.listen()
	return s
}

func Listen(addr string, bkd backend.Backend) (s *Server, err error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}

	s = NewServer(l, bkd)
	return
}

func ListenTLS(addr string, bkd backend.Backend, tlsConfig *tls.Config) (s *Server, err error) {
	l, err := tls.Listen("tcp", addr, tlsConfig)
	if err != nil {
		return
	}

	s = NewServer(l, bkd)
	return
}
