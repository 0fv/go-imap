package imap

// A response that can be either accepted or rejected by a handler.
type RespHandling struct {
	Resp    interface{}
	Accepts chan bool
}

// Accept this response. This means that the handler will process it.
func (h *RespHandling) Accept() {
	h.Accepts <- true
}

// Reject this response. The handler cannot process it.
func (h *RespHandling) Reject() {
	h.Accepts <- false
}

// Accept this response if it has the specified name. If not, reject it.
func (h *RespHandling) AcceptNamedResp(name string) (fields []interface{}, accepted bool) {
	res, ok := h.Resp.(*Resp)
	if !ok || len(res.Fields) == 0 {
		h.Reject()
		return
	}

	n, ok := res.Fields[0].(string)
	if !ok || n != name {
		h.Reject()
		return
	}

	h.Accept()

	fields = res.Fields[1:]
	accepted = true
	return
}

// Delivers responses to handlers.
type RespHandler chan *RespHandling

// Handles responses from a handler.
type RespHandlerFrom interface {
	HandleFrom(hdlr RespHandler) error
}
