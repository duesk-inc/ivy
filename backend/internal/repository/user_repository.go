package repository

import (
	"context"
	"fmt"

	"github.com/duesk/ivy/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserRepository ユーザーリポジトリインターフェース
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByCognitoSub(ctx context.Context, cognitoSub string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
}

type userRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewUserRepository ユーザーリポジトリを作成
func NewUserRepository(db *gorm.DB, logger *zap.Logger) UserRepository {
	return &userRepository{db: db, logger: logger}
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, fmt.Errorf("ユーザー取得失敗 (id=%s): %w", id, err)
	}
	return &user, nil
}

func (r *userRepository) GetByCognitoSub(ctx context.Context, cognitoSub string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("cognito_sub = ?", cognitoSub).First(&user).Error; err != nil {
		return nil, fmt.Errorf("ユーザー取得失敗 (cognito_sub=%s): %w", cognitoSub, err)
	}
	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("ユーザー作成失敗: %w", err)
	}
	return nil
}
