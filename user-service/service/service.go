package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shafaalafghany/user-service/model"
	"github.com/shafaalafghany/user-service/repository"
	"gitlab.com/shafaalafghany/synapsis-proto/go/user"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServiceInterface interface {
	Register(context.Context, *user.RegisterRequest) (*user.RegisterResponse, error)
}

type UserService struct {
	repo repository.UserRepositoryInterface
	log  *zap.Logger
}

func NewUserService(repo repository.UserRepositoryInterface, log *zap.Logger) UserServiceInterface {
	return &UserService{
		repo: repo,
		log:  log,
	}
}

func (s *UserService) Register(ctx context.Context, body *user.RegisterRequest) (*user.RegisterResponse, error) {
	id := uuid.NewString()
	data := &model.User{
		ID:    id,
		Name:  body.Name,
		Email: body.Email,
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	data.Password = string(hash)

	if err := s.repo.Create(data); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := fmt.Sprintf("register new account successfully with id %s", id)

	return &user.RegisterResponse{Message: response}, nil
}
