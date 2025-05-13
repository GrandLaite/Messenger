package service

import (
	"errors"

	"user-service/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = bcrypt.DefaultCost

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(r *repository.UserRepository) *UserService {
	return &UserService{repo: r}
}

func (s *UserService) CreateUser(username, password, role, email, nickname string) (int, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return 0, err
	}
	u := repository.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
		Email:        email,
		Nickname:     nickname,
	}
	return s.repo.Create(u)
}

func (s *UserService) CheckPassword(username, password string) (repository.User, error) {
	u, err := s.repo.GetByUsername(username)
	if err != nil {
		return repository.User{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return repository.User{}, errors.New("invalid password")
	}
	return u, nil
}

func (s *UserService) GetUserByNickname(nickname string) (repository.User, error) {
	return s.repo.GetByNickname(nickname)
}
