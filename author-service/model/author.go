package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Author struct {
	ID        string     `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	Name      string     `json:"name" gorm:"not null;index"`
	CreatedBy string     `json:"created_by" gorm:"not null"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

func (a *Author) BeforeCreate(tx *gorm.DB) (err error) {
	a.ID = uuid.NewString()
	return
}
