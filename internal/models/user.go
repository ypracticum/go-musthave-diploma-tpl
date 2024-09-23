package models

type UnknownUser struct {
	Login    *string `json:"login"`
	Password *string `json:"password"`
}

type User struct {
	ID    string
	Login string
	Hash  string
}
