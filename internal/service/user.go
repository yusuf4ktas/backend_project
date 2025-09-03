package service

import (
	"context"
	"fmt"
	"time"

	"github.com/yusuf4ktas/backend-project/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo     domain.UserRepository
	auditService AuditLogService
	balanceRepo  domain.BalanceRepository
}

func NewUserService(repo domain.UserRepository, auditService AuditLogService, balanceRepo domain.BalanceRepository) UserService {
	return &userService{
		userRepo:     repo,
		auditService: auditService,
		balanceRepo:  balanceRepo,
	}
}

func (s *userService) Register(ctx context.Context, username, email, password string) (*domain.User, error) {
	user := &domain.User{
		Username:  username,
		Email:     email,
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := user.Validate(password)
	if err != nil {
		return nil, err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = string(hashedPassword)

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	initialBalance := &domain.Balance{
		UserID:        user.ID,
		Amount:        0.00,
		LastUpdatedAt: time.Now(),
	}

	err = s.balanceRepo.Create(ctx, initialBalance)
	if err != nil {
		return nil, fmt.Errorf("failed to create initial balance for user: %w", err)
	}

	details := fmt.Sprintf("User %s registered with email %s", user.Username, user.Email)
	_, _ = s.auditService.Log(ctx, "user", user.ID, "register", details)

	return user, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, err
	}

	details := fmt.Sprintf("User %s registered with email %s", user.Username, user.Email)
	_, _ = s.auditService.Log(ctx, "user", user.ID, "login", details)

	return user, nil
}

func (s *userService) GetByID(ctx context.Context, userID int64) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	users, err := s.userRepo.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *userService) Delete(ctx context.Context, userID int64) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	return s.userRepo.Delete(ctx, user.ID)
}
