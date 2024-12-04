package repository

import (
	"github.com/shafaalafghany/author-service/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuthorRepositoryInterface interface {
	Create(*model.Author) error
}

type AuthorRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewAuthorRepository(db *gorm.DB, log *zap.Logger) AuthorRepositoryInterface {
	return &AuthorRepository{
		db:  db,
		log: log,
	}
}

func (r *AuthorRepository) Create(data *model.Author) error {
	if err := r.db.Create(&data).Error; err != nil {
		return err
	}
	return nil
}
