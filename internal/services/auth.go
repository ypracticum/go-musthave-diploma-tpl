package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/Renal37/go-musthave-diploma-tpl/internal/database"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// Определение пользовательских ошибок
var (
	ErrUserIsAlreadyRegistered = errors.New("пользователь уже зарегистрирован")
	ErrUserIsNotExist          = errors.New("пользователь не существует")
	ErrPasswordIsIncorrect     = errors.New("пароль неверен")
)

// AuthService представляет сервис для аутентификации и управления пользователями
type AuthService struct {
	storage AuthStorage
}

// AuthStorage определяет интерфейс для взаимодействия с хранилищем данных пользователей
type AuthStorage interface {
	CreateUser(ctx context.Context, user database.UserDB) error           // Создание нового пользователя
	FindUser(ctx context.Context, login string) (*database.UserDB, error) // Поиск пользователя по логину
}

// NewAuthService создает новый экземпляр AuthService с заданным хранилищем
func NewAuthService(storage AuthStorage) *AuthService {
	return &AuthService{storage: storage}
}

// Register регистрирует нового пользователя
func (auth *AuthService) Register(ctx context.Context, user models.UnknownUser) error {
	// Проверка валидности входных данных
	if err := validateUser(user); err != nil {
		return err
	}

	// Хэширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("ошибка при хэшировании пароля: %w", err)
	}

	// Создание пользователя в хранилище
	err = auth.storage.CreateUser(ctx, database.UserDB{
		User: models.User{
			Login: *user.Login,
			Hash:  string(hashedPassword),
		},
	})
	if err != nil {
		if errors.Is(err, database.ErrDuplicateUser) {
			return ErrUserIsAlreadyRegistered
		}
		return fmt.Errorf("ошибка при создании пользователя: %w", err)
	}

	return nil
}

// Login выполняет аутентификацию пользователя
func (auth *AuthService) Login(ctx context.Context, user models.UnknownUser) error {
	// Проверка валидности входных данных
	if err := validateUser(user); err != nil {
		return err
	}

	// Поиск пользователя по логину
	u, err := auth.storage.FindUser(ctx, *user.Login)
	if err != nil {
		return fmt.Errorf("ошибка при поиске пользователя: %w", err)
	}

	if u == nil {
		return ErrUserIsNotExist
	}

	// Сравнение пароля
	if err := bcrypt.CompareHashAndPassword([]byte(u.Hash), []byte(*user.Password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrPasswordIsIncorrect
		}
		return fmt.Errorf("ошибка при сравнении паролей: %w", err)
	}

	return nil
}

// GetUser возвращает информацию о пользователе по логину
func (auth *AuthService) GetUser(ctx context.Context, login string) (*models.User, error) {
	// Поиск пользователя по логину
	user, err := auth.storage.FindUser(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("ошибка при поиске пользователя: %w", err)
	}

	if user == nil {
		return nil, ErrUserIsNotExist
	}

	return &user.User, nil
}

// validateUser проверяет валидность входных данных пользователя
func validateUser(user models.UnknownUser) error {
	if user.Login == nil || *user.Login == "" {
		return errors.New("логин не может быть пустым")
	}
	if user.Password == nil || *user.Password == "" {
		return errors.New("пароль не может быть пустым")
	}
	return nil
}
