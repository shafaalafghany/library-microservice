package repository

import (
	"github.com/shafaalafghany/category-service/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CategoryRepositoryInterface interface {
	Create(*model.Category) error
}

type CategoryRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewCategoryRepository(db *gorm.DB, log *zap.Logger) CategoryRepositoryInterface {
	return &CategoryRepository{
		db:  db,
		log: log,
	}
}

func (cr *CategoryRepository) Create(data *model.Category) error {
	if err := cr.db.Create(&data).Error; err != nil {
		return err
	}
	return nil
}
