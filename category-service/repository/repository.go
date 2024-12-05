package repository

import (
	"time"

	"github.com/shafaalafghany/category-service/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CategoryRepositoryInterface interface {
	Create(*model.Category) error
	GetById(string) (*model.Category, error)
	Get(string) ([]*model.Category, error)
	Update(*model.Category, string) error
	Delete(string) error
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

func (cr *CategoryRepository) GetById(id string) (*model.Category, error) {
	var category model.Category
	if err := cr.db.Where("id = ? AND deleted_at IS NULL", id).First(&category).Error; err != nil {
		return nil, err
	}

	return &category, nil
}

func (cr *CategoryRepository) Get(search string) ([]*model.Category, error) {
	var categories []*model.Category
	base := cr.db.Model(&model.Category{}).Where("deleted_at IS NULL")

	if search != "" {
		base.Where("name ILIKE ?", "%"+search+"%")
	}

	if err := base.Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

func (cr *CategoryRepository) Update(data *model.Category, id string) error {
	if err := cr.db.Model(&model.Category{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("name", data.Name).Error; err != nil {
		return err
	}
	return nil
}

func (cr *CategoryRepository) Delete(id string) error {
	if err := cr.db.Model(&model.Category{}).Where("id = ? AND deleted_at IS NULL", id).Update("deleted_at", time.Now()).Error; err != nil {
		return err
	}
	return nil
}
