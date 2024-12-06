package handler

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/shafaalafghany/book-service/service"
	"gitlab.com/shafaalafghany/synapsis-proto/go/book"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type BookHandler struct {
	book.UnimplementedBookServiceServer
	s   service.BookServiceInterface
	log *zap.Logger
}

func NewBookHandler(s service.BookServiceInterface, log *zap.Logger) *BookHandler {
	return &BookHandler{
		s:   s,
		log: log,
	}
}

func (h *BookHandler) Create(ctx context.Context, body *book.Book) (*book.CommonBookResponse, error) {
	userId, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	body.CreatedBy = userId

	return h.s.CreateBook(ctx, body)
}

func (h *BookHandler) Get(ctx context.Context, body *book.Book) (*book.Book, error) {
	return h.s.GetBook(ctx, body)
}

func (h *BookHandler) Getlist(ctx context.Context, body *book.BookRequest) (*book.BooksResponse, error) {
	return h.s.GetBooks(ctx, body)
}

func (h *BookHandler) Update(ctx context.Context, body *book.Book) (*book.CommonBookResponse, error) {
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	return h.s.UpdateBook(ctx, body)
}

func (h *BookHandler) Delete(ctx context.Context, body *book.Book) (*book.CommonBookResponse, error) {
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	return h.s.DeleteBook(ctx, body)
}

func (h *BookHandler) BorrowBook(ctx context.Context, body *book.BorrowRecord) (*book.CommonBorrowRecordResponse, error) {
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	return h.s.BorrowBook(ctx, body)
}

func (h *BookHandler) ReturnBook(ctx context.Context, body *book.BorrowRecord) (*book.CommonBorrowRecordResponse, error) {
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	return h.s.ReturnBook(ctx, body)
}

func (h *BookHandler) GetRecommendation(ctx context.Context, body *book.BookRequest) (*book.BooksResponse, error) {
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	return h.s.GetRecommendation(ctx, body)
}

func getUserIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("missing metadata")
	}

	authHeader, ok := md["authorization"]
	if !ok || len(authHeader) == 0 {
		return "", errors.New("missing authorization header")
	}

	tokenStr := strings.TrimPrefix(authHeader[0], "Bearer ")
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(os.Getenv("SECRET_KEY")), nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token claims")
	}

	userID, ok := claims["id"].(string)
	if !ok {
		return "", errors.New("invalid user ID in token claims")
	}

	return userID, nil
}
