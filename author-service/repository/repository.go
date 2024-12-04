package repository

import (
	"time"

	"github.com/shafaalafghany/author-service/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuthorRepositoryInterface interface {
	Create(*model.Author) error
	GetById(string) (*model.Author, error)
	Get(string) ([]*model.Author, error)
	Update(*model.Author, string) error
	Delete(string) error
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

func (r *AuthorRepository) GetById(id string) (*model.Author, error) {
	var author model.Author
	if err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&author).Error; err != nil {
		return nil, err
	}

	return &author, nil
}

func (r *AuthorRepository) Get(search string) ([]*model.Author, error) {
	var authors []*model.Author
	base := r.db.Model(&model.Author{}).Where("deleted_at IS NULL")

	if search != "" {
		base.Where("name ILIKE ?", "%"+search+"%")
	}

	if err := base.Find(&authors).Error; err != nil {
		return nil, err
	}

	return authors, nil
}

func (r *AuthorRepository) Update(data *model.Author, id string) error {
	if err := r.db.Model(&model.Author{}).Where("id = ? AND deleted_at IS NULL", id).Update("name", data.Name).Error; err != nil {
		return err
	}
	return nil
}

func (r *AuthorRepository) Delete(id string) error {
	if err := r.db.Model(&model.Author{}).Where("id = ? AND deleted_at IS NULL", id).Update("deleted_at", time.Now()).Error; err != nil {
		return err
	}
	return nil
}
