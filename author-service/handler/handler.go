package handler

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/shafaalafghany/author-service/service"
	"gitlab.com/shafaalafghany/synapsis-proto/go/author"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthorHandler struct {
	author.UnimplementedAuthorServiceServer
	as  service.AuthorServiceInterface
	log *zap.Logger
}

func NewAuthorHandler(as service.AuthorServiceInterface, log *zap.Logger) *AuthorHandler {
	return &AuthorHandler{
		as:  as,
		log: log,
	}
}

func (h *AuthorHandler) Create(ctx context.Context, body *author.Author) (*author.CommonAuthorResponse, error) {
	userId, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	body.CreatedBy = userId

	return h.as.CreateAuthor(ctx, body)
}

func (h *AuthorHandler) Get(ctx context.Context, body *author.Author) (*author.Author, error) {
	return h.as.GetAuthor(ctx, body)
}

func (h *AuthorHandler) Update(ctx context.Context, body *author.Author) (*author.CommonAuthorResponse, error) {
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	return h.as.UpdateAuthor(ctx, body)
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
