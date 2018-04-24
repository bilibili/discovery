package errors

import "strconv"

// Error error.
type Error interface {
	error
	// Code get error code.
	Code() int
}

type ecode int

// ecode error.
var (
	NotModified  ecode = -304
	ParamsErr    ecode = -400
	NothingFound ecode = -404
	Conflict     ecode = -409
	ServerErr    ecode = -500
)

func (e ecode) Error() string {
	return strconv.FormatInt(int64(e), 10)
}

func (e ecode) Code() int {
	return int(e)
}

// Code converts error to ecode.
func Code(e error) (ie Error) {
	if e == nil {
		return
	}
	i, err := strconv.Atoi(e.Error())
	if err != nil {
		i = -500
	}
	ie = ecode(i)
	return
}
