package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/shafaalafghany/book-service/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BookRepositoryInterface interface {
	Create(*model.Book) error
	GetById(context.Context, *model.Book) (*model.Book, error)
	Get(context.Context, string) ([]*model.Book, error)
	Update(context.Context, *model.Book, string) error
	Delete(context.Context, string) error
	Borrow(context.Context, *model.BorrowRecord) error
	ReturnBook(context.Context, *model.BorrowRecord) error
	MostBorrows(string) ([]*model.Book, error)
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

func (r *BookRepository) Borrow(ctx context.Context, data *model.BorrowRecord) error {
	lockKey := fmt.Sprintf("lock:book:%s", data.BookID)
	if ok := r.redis.SetNX(ctx, lockKey, data.UserID, 30*time.Second).Val(); !ok {
		return fmt.Errorf("book currently is still borrowed, try again later")
	}
	defer r.redis.Del(ctx, lockKey)

	var book model.Book
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&book, "id = ? AND deleted_at IS NULL", data.BookID).Error; err != nil {
		return err
	}

	if book.IsBorrowed {
		return errors.New("book is still borrowed")
	}

	if err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&data).Error; err != nil {
			return err
		}

		book.IsBorrowed = true
		book.Borrows++
		if err := tx.Save(&book).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (r *BookRepository) ReturnBook(ctx context.Context, data *model.BorrowRecord) error {
	var book model.Book
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&book, "id = ? AND deleted_at IS NULL", data.BookID).Error; err != nil {
		return err
	}

	if !book.IsBorrowed {
		return errors.New("book is not borrowed")
	}

	var borrowRecord model.BorrowRecord
	if err := r.db.First(&borrowRecord, "book_id = ? AND user_id = ? AND returned_at IS NULL", data.BookID, data.UserID).Error; err != nil {
		return err
	}

	now := time.Now()
	borrowRecord.ReturnedAt = &now

	if err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&borrowRecord).Error; err != nil {
			return err
		}

		book.IsBorrowed = false
		if err := tx.Save(&book).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (r *BookRepository) MostBorrows(search string) ([]*model.Book, error) {
	base := r.db.Order("borrows DESC").Limit(5)

	if search != "" {
		base.Where("name ILIKE ?", "%"+search+"%")
	}
	var books []*model.Book
	if err := base.Find(&books).Error; err != nil {
		return nil, err
	}

	return books, nil
}
