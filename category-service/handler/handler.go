package handler

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/shafaalafghany/category-service/service"
	"gitlab.com/shafaalafghany/synapsis-proto/go/category"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type CategoryHandler struct {
	category.UnimplementedCategoryServiceServer
	cs  service.CategoryServiceInterface
	log *zap.Logger
}

func NewCategoryHandler(cs service.CategoryServiceInterface, log *zap.Logger) *CategoryHandler {
	return &CategoryHandler{
		cs:  cs,
		log: log,
	}
}

func (ch *CategoryHandler) Create(ctx context.Context, body *category.Category) (*category.CommonCategoryResponse, error) {
	userId, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	body.CreatedBy = userId

	return ch.cs.CreateCategory(ctx, body)
}

func (ch *CategoryHandler) Get(ctx context.Context, body *category.Category) (*category.Category, error) {
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	return ch.cs.GetCategory(ctx, body)
}

func (ch *CategoryHandler) GetList(ctx context.Context, body *category.CategoryRequest) (*category.CategoriesResponse, error) {
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	return ch.cs.GetCategories(ctx, body)
}

func (ch *CategoryHandler) Update(ctx context.Context, body *category.Category) (*category.CommonCategoryResponse, error) {
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	return ch.cs.UpdateCategory(ctx, body)
}
func (ch *CategoryHandler) Delete(ctx context.Context, body *category.Category) (*category.CommonCategoryResponse, error) {
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	return ch.cs.DeleteCategory(ctx, body)
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
