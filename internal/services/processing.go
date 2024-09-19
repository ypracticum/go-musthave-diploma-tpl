package services

import (
	"fmt"

	"github.com/Renal37/go-musthave-diploma-tpl/internal/users"
)

type Repo interface {
	CreateUser(id int, login, password string) error
	GetUser(login string) (*users.User, error)
}

type BoxService struct {
	db Repo
}

func NewBoxService(db Repo) *BoxService {
	return &BoxService{
		db,
	}
}

func (b *BoxService) GetUser(login string) (*users.User, error) {
	return b.db.GetUser(login)
}

func (b *BoxService) CreateUser(user *users.User) error {
	hashedPassword, err := user.PasswordStringToHash()
	if err != nil {
		return err
	}
	err = b.db.CreateUser(user.ID, user.Login, hashedPassword)
	if err != nil {
		return fmt.Errorf("failed registration user in db  %w", err)
	}
	return nil
}
