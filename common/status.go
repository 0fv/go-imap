package common

// A status response type.
type StatusRespType string

const (
	OK StatusRespType = "OK"
	NO = "NO"
	BAD = "BAD"
	PREAUTH = "PREAUTH"
	BYE = "BYE"
)

// A status response code.
type StatusRespCode string

const (
	ALERT StatusRespCode = "ALERT"
	BADCHARSET = "BADCHARSET"
	CAPABILITY = "CAPABILITY"
	PARSE = "PARSE"
	PERMANENTFLAGS = "PERMANENTFLAGS"
	READ_ONLY = "READ-ONLY"
	READ_WRITE = "READ-WRITE"
	TRYCREATE = "TRYCREATE"
	UIDNEXT = "UIDNEXT"
	UIDVALIDITY = "UIDVALIDITY"
	UNSEEN = "UNSEEN"
)

// A status response.
// See https://tools.ietf.org/html/rfc3501#section-7.1
type StatusResp struct {
	Tag string
	Type StatusRespType
	Code StatusRespCode
	Arguments []interface{}
	Info string
}

func (r *StatusResp) WriteTo(w io.Writer) (n int64, err error) {
	// TODO
	return
}

func StatusRespFromError(tag string, err error) *StatusResp {
	res := &StatusResp{
		Tag: tag,
	}

	if err == nil {
		res.Type = OK
		res.Info = tag + " completed"
		return res
	}

	res.Info = err.Error()

	//if imapErr, ok := err.(Error); ok {
	//	res.Type = imapErr.Completion()
	//} else {
	res.Type = BAD
	//}

	return res
}
