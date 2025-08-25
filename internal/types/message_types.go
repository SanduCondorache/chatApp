package types

type MessageType string

const (
	Error MessageType = "error"
	Login MessageType = "login"
	Find  MessageType = "find_user"
	Chat  MessageType = "chat"
	Exit  MessageType = "exit"
)
