package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Book struct {
	ID         string     `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	Name       string     `json:"name" gorm:"not null"`
	AuthorID   string     `json:"author_id" gorm:"not null;index"`
	CategoryID string     `json:"category_id" gorm:"not null;index"`
	IsBorrowed bool       `json:"is_borrowed" gorm:"not null"`
	Borrows    int        `json:"borrows" gorm:"not null"`
	CreatedBy  string     `json:"created_by" gorm:"not null;index"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt  *time.Time `json:"deleted_at" gorm:"index"`
}

func (b *Book) BeforeCreate(tx *gorm.DB) (err error) {
	b.ID = uuid.NewString()
	return
}
