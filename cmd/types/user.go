package types

type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"passowrd"`
}

func NewUser(Username, Email, Password string) User {
	return User{
		Username: Username,
		Email:    Email,
		Password: Password,
	}
}
