package errors

import "strconv"

type Error interface {
	error
	// Code get error code.
	Code() int
}
type ecode int

var (
	ServerErr    ecode = -500
	ParamsErr    ecode = -400
	NothingFound ecode = -404
	NotModified  ecode = -304
	Conflict     ecode = -409
)

func (e ecode) Error() string {
	return strconv.FormatInt(int64(e), 10)
}

func (e ecode) Code() int {
	return int(e)
}

func Code(e error) (i int) {
	i, err := strconv.Atoi(e.Error())
	if err != nil {
		i = -500
	}
	return
}
