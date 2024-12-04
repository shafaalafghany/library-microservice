package handler

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/shafaalafghany/user-service/service"
	"gitlab.com/shafaalafghany/synapsis-proto/go/user"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserHandler struct {
	user.UnimplementedUserServiceServer
	us  service.UserServiceInterface
	log *zap.Logger
}

func NewUserHandler(us service.UserServiceInterface, log *zap.Logger) *UserHandler {
	return &UserHandler{
		us:  us,
		log: log,
	}
}

func (h *UserHandler) Register(ctx context.Context, body *user.RegisterRequest) (*user.RegisterResponse, error) {
	return h.us.Register(ctx, body)
}

func (h *UserHandler) Login(ctx context.Context, body *user.LoginRequest) (*user.LoginResponse, error) {
	return h.us.Login(ctx, body)
}

func (h *UserHandler) GetUser(ctx context.Context, empty *emptypb.Empty) (*user.User, error) {
	userId, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	user, err := h.us.Get(ctx, &user.User{Id: userId})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return user, nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, body *user.User) (*user.CommonUserResponse, error) {
	userId, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	res, err := h.us.Update(ctx, body, userId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, empty *emptypb.Empty) (*user.CommonUserResponse, error) {
	userId, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	res, err := h.us.Delete(ctx, userId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res, nil
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
