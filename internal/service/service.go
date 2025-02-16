package service

import (
	"avito-shop/internal/models"
	"avito-shop/internal/repository"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrSelfTransfer      = errors.New("невозможно перевести монеты самому себе")
	ErrInvalidItem       = errors.New("некорректный товар")
	ErrInsufficientCoins = errors.New("недостаточно монет")
	ErrNotFound          = repository.ErrNotFound
	ErrInvalidInput      = errors.New("некорректные входные данные")
)

type Service struct {
	repo      *repository.Repository
	jwtSecret string
}

func NewService(repo *repository.Repository, jwtSecret string) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *Service) Authenticate(ctx context.Context, username, password string) (string, error) {
	if username == "" || password == "" {
		fmt.Println("Ошибка: пустое имя пользователя или пароль")
		return "", ErrInvalidInput
	}

	fmt.Println("Попытка найти пользователя:", username)
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			fmt.Println("Пользователь не найден, создаём нового:", username)

			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
			if err != nil {
				fmt.Println("Ошибка хеширования пароля:", err)
				return "", fmt.Errorf("ошибка хеширования пароля: %w", err)
			}

			newUser := &models.User{
				Username:     username,
				PasswordHash: string(hashedPassword),
				Coins:        1000,
			}

			if err = s.repo.CreateUser(ctx, newUser); err != nil {
				fmt.Println("Ошибка при создании пользователя:", err)
				return "", fmt.Errorf("ошибка создания пользователя: %w", err)
			}

			fmt.Println("Пользователь успешно создан:", newUser.ID)
			user = newUser
		} else {
			fmt.Println("Ошибка при получении пользователя:", err)
			return "", fmt.Errorf("ошибка получения пользователя: %w", err)
		}
	} else {
		fmt.Println("Пользователь найден:", user.ID)

		if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			fmt.Println("Ошибка: неверный пароль")
			return "", fmt.Errorf("неверный пароль: %w", err)
		}
	}

	if s.jwtSecret == "" {
		fmt.Println("Ошибка: секретный ключ JWT не задан")
		return "", errors.New("секретный ключ JWT не задан")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		fmt.Println("Ошибка генерации токена:", err)
		return "", fmt.Errorf("ошибка генерации токена: %w", err)
	}

	fmt.Println("Аутентификация успешна, токен сгенерирован")
	return tokenString, nil
}

func (s *Service) TransferCoins(ctx context.Context, fromUserID int, toUsername string, amount int) error {
	if amount <= 0 {
		return ErrInvalidInput
	}

	toUser, err := s.repo.GetUserByUsername(ctx, toUsername)
	if err != nil {
		return fmt.Errorf("получатель не найден: %w", ErrNotFound)
	}

	if toUser == nil {
		return ErrNotFound
	}

	if fromUserID == toUser.ID {
		return ErrSelfTransfer
	}

	return s.repo.TransferCoins(ctx, fromUserID, toUser.ID, amount)
}

func (s *Service) BuyItem(ctx context.Context, userID int, itemName string) error {
	if itemName == "" {
		return ErrInvalidInput
	}

	item, err := s.repo.GetItemByName(ctx, itemName)
	if err != nil {
		return fmt.Errorf("%w: товар не найден", ErrNotFound)
	}

	if item.Price <= 0 {
		return fmt.Errorf("%w: цена товара некорректна", ErrInvalidItem)
	}

	user, err := s.repo.GetUserInfo(ctx, userID)
	if err != nil {
		return fmt.Errorf("ошибка получения данных пользователя: %w", err)
	}

	if user.Coins < item.Price {
		return ErrInsufficientCoins
	}

	return s.repo.BuyItem(ctx, userID, item.ID)
}

func (s *Service) GetUserInfo(ctx context.Context, userID int) (*models.InfoResponse, error) {
	info, err := s.repo.GetUserInfo(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения информации: %w", err)
	}
	return info, nil
}
