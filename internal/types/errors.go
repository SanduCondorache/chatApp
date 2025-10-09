package types

import "errors"

var (
	ErrorUsernameTaken     = errors.New("username_is_taken_error")
	ErrorUserNotFound      = errors.New("user_not_found_error")
	ErrorIncorrectPassowrd = errors.New("incorrect_password_error")
	ErrorConnectionClosed  = errors.New("connection closed")
)
