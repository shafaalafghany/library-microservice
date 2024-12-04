package repository

import (
	"github.com/shafaalafghany/user-service/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
	Create(*model.User) error
	GetUserByEmail(*model.User) (*model.User, error)
	GetUserById(*model.User) (*model.User, error)
	UpdateUser(*model.User, string) error
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

func (r *UserRepository) GetUserByEmail(data *model.User) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ? AND deleted_at IS NULL", data.Email).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUserById(data *model.User) (*model.User, error) {
	var user model.User
	if err := r.db.Where("id = ? AND deleted_at IS NULL", data.ID).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) UpdateUser(data *model.User, id string) error {
	updatedData := map[string]interface{}{
		"name":  data.Name,
		"email": data.Email,
	}

	if err := r.db.Model(&model.User{}).Where("id = ?", id).Updates(updatedData).Error; err != nil {
		return err
	}

	return nil
}
