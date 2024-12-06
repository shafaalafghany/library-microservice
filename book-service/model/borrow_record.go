package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BorrowRecord struct {
	ID         string     `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	BookID     string     `gorm:"not null;index"`
	UserID     string     `gorm:"not null;index"`
	BorrowedAt time.Time  `gorm:"not null"`
	ReturnedAt *time.Time `gorm:""`
	CreatedAt  time.Time  `gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime"`
	DeletedAt  *time.Time `gorm:"index"`
}

func (b *BorrowRecord) BeforeCreate(tx *gorm.DB) (err error) {
	b.ID = uuid.NewString()
	return
}
