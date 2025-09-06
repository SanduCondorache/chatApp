package types

import "errors"

var (
	ErrorSendingMessage    = errors.New("sending message error")
	ErrorOpeningDB         = errors.New("opening database error")
	ErrorQuery             = errors.New("query error")
	ErrorUpgradeConnection = errors.New("upgrade error")
	ErrorUsernameTaken     = errors.New("username is taken")
	ErrorReadJson          = errors.New("read json error")
	ErrorBadMessage        = errors.New("Unknown message type")
	ErrorReadingMessage    = errors.New("reading error")
	ErrorWriteJson         = errors.New("write json error")
	ErrorUserNotFound      = errors.New("user not found error")
)

type Errors struct {
	appErr error
	svcErr error
}

func NewError(appErr, svcErr error) *Errors {
	return &Errors{
		appErr: appErr,
		svcErr: svcErr,
	}
}

func (e *Errors) Error() string {
	return errors.Join(e.svcErr, e.appErr).Error()
}
