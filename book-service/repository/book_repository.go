package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/shafaalafghany/book-service/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type BookRepositoryInterface interface {
	Create(*model.Book) error
	GetById(context.Context, *model.Book) (*model.Book, error)
	Get(context.Context, string) ([]*model.Book, error)
	Update(context.Context, *model.Book, string) error
	Delete(context.Context, string) error
}

type BookRepository struct {
	db     *gorm.DB
	logger *zap.Logger
	redis  *redis.Client
}

func NewBookRepository(db *gorm.DB, logger *zap.Logger, redis *redis.Client) BookRepositoryInterface {
	return &BookRepository{
		db:     db,
		logger: logger,
		redis:  redis,
	}
}

func (r *BookRepository) Create(data *model.Book) error {
	if err := r.db.Create(&data).Error; err != nil {
		return err
	}
	return nil
}

func (r *BookRepository) GetById(ctx context.Context, data *model.Book) (*model.Book, error) {
	var book model.Book
	key := fmt.Sprintf("book:%s", data.ID)
	bookRedis, err := r.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		if err := r.db.First(&book).Error; err != nil {
			return nil, err
		}

		bookJson, err := json.Marshal(book)
		if err != nil {
			return nil, err
		}

		if err = r.redis.Set(ctx, key, bookJson, 5*time.Minute).Err(); err != nil {
			return nil, err
		}

		return &book, nil
	} else if err != nil {
		return nil, err
	}

	if err = json.Unmarshal([]byte(bookRedis), &book); err != nil {
		return nil, err
	}

	return &book, nil
}

func (r *BookRepository) Get(ctx context.Context, search string) ([]*model.Book, error) {
	var books []*model.Book
	base := r.db.Model(&model.Book{}).Where("deleted_at IS NULL")

	if search != "" {
		if err := base.Where("name ILIKE ?", "%"+search+"%").Find(&books).Error; err != nil {
			return nil, err
		}
		return books, nil
	}

	key := "books"
	booksRedis, err := r.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		if err := base.Find(&books).Error; err != nil {
			return nil, err
		}

		booksJson, err := json.Marshal(books)
		if err != nil {
			return nil, err
		}

		if err = r.redis.Set(ctx, key, booksJson, 5*time.Minute).Err(); err != nil {
			return nil, err
		}

		return books, nil
	} else if err != nil {
		return nil, err
	}

	if err = json.Unmarshal([]byte(booksRedis), &books); err != nil {
		return nil, err
	}

	return books, nil
}

func (r *BookRepository) Update(ctx context.Context, data *model.Book, id string) error {
	updatedData := map[string]interface{}{
		"name":        data.Name,
		"author_id":   data.AuthorID,
		"category_id": data.CategoryID,
	}

	exist, err := r.redis.Exists(ctx, fmt.Sprintf("book:%s", id)).Result()
	if err != nil {
		return err
	}

	if exist == 1 {
		if err = r.redis.Del(ctx, fmt.Sprintf("book:%s", id)).Err(); err != nil {
			return err
		}
	}

	exist, err = r.redis.Exists(ctx, "books").Result()
	if err != nil {
		return err
	}

	if exist == 1 {
		if err = r.redis.Del(ctx, "books").Err(); err != nil {
			return err
		}
	}

	if err = r.db.Model(&model.Book{}).Where("id = ? AND deleted_at IS NULL", id).Updates(updatedData).Error; err != nil {
		return err
	}

	return nil
}

func (r *BookRepository) Delete(ctx context.Context, id string) error {
	exist, err := r.redis.Exists(ctx, fmt.Sprintf("book:%s", id)).Result()
	if err != nil {
		return err
	}

	if exist == 1 {
		if err = r.redis.Del(ctx, fmt.Sprintf("book:%s", id)).Err(); err != nil {
			return err
		}
	}

	exist, err = r.redis.Exists(ctx, "books").Result()
	if err != nil {
		return err
	}

	if exist == 1 {
		if err = r.redis.Del(ctx, "books").Err(); err != nil {
			return err
		}
	}

	if err := r.db.Model(&model.Book{}).Where("id = ? AND deleted_at IS NULL", id).Update("deleted_at", time.Now()).Error; err != nil {
		return err
	}

	return nil
}
