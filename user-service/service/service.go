package service

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
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
	Login(context.Context, *user.LoginRequest) (*user.LoginResponse, error)
	Get(context.Context, *user.User) (*user.User, error)
	Update(context.Context, *user.User, string) (*user.CommonUserResponse, error)
	Delete(context.Context, string) (*user.CommonUserResponse, error)
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

func (s *UserService) Login(ctx context.Context, body *user.LoginRequest) (*user.LoginResponse, error) {
	req := model.User{Email: body.Email}

	existsData, err := s.repo.GetUserByEmail(&req)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	if existsData == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(existsData.Password), []byte(body.Password))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "account not found")
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  existsData.ID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	token, err := claims.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	userData := &user.User{
		Id:        existsData.ID,
		Email:     body.Email,
		Name:      existsData.Name,
		CreatedAt: existsData.CreatedAt.String(),
		UpdatedAt: existsData.UpdatedAt.String(),
	}

	res := &user.LoginResponse{
		User:  userData,
		Token: token,
	}

	return res, nil
}

func (s *UserService) Get(ctx context.Context, body *user.User) (*user.User, error) {
	existsUser, err := s.repo.GetUserById(&model.User{ID: body.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if existsUser == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	existsUser.Password = ""

	userData := &user.User{
		Id:        existsUser.ID,
		Email:     existsUser.Email,
		Name:      existsUser.Name,
		CreatedAt: existsUser.CreatedAt.String(),
		UpdatedAt: existsUser.UpdatedAt.String(),
	}

	return userData, nil
}

func (s *UserService) Update(ctx context.Context, body *user.User, id string) (*user.CommonUserResponse, error) {
	data := &model.User{
		Name:  body.Name,
		Email: body.Email,
	}

	existsUser, err := s.repo.GetUserById(&model.User{ID: id})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if existsUser == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	if err := s.repo.UpdateUser(data, id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.CommonUserResponse{Message: "update user successfully"}, nil
}

func (s *UserService) Delete(ctx context.Context, id string) (*user.CommonUserResponse, error) {
	existsUser, err := s.repo.GetUserById(&model.User{ID: id})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if existsUser == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	if err := s.repo.DeleteUser(id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.CommonUserResponse{Message: "delete user successfully"}, nil
}
