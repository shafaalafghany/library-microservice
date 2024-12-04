package repository

import (
	"github.com/shafaalafghany/user-service/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
	Create(*model.User) error
}

type UserRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewUserRepository(db *gorm.DB, log *zap.Logger) UserRepositoryInterface {
	return &UserRepository{
		db:  db,
		log: log,
	}
}

func (r *UserRepository) Create(data *model.User) error {
	if err := r.db.Create(data).Error; err != nil {
		r.log.Info("failed to create user", zap.Any("users", err))
		return err
	}
	return nil
}
