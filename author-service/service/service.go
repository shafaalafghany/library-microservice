package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shafaalafghany/author-service/model"
	"github.com/shafaalafghany/author-service/repository"
	"gitlab.com/shafaalafghany/synapsis-proto/go/author"
	"gitlab.com/shafaalafghany/synapsis-proto/go/user"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthorServiceInterface interface {
	CreateAuthor(context.Context, *author.Author) (*author.CommonAuthorResponse, error)
	GetAuthor(context.Context, *author.Author) (*author.Author, error)
	GetAuthors(context.Context, *author.AuthorRequest) (*author.AuthorsResponse, error)
	UpdateAuthor(context.Context, *author.Author) (*author.CommonAuthorResponse, error)
	DeleteAuthor(context.Context, *author.Author) (*author.CommonAuthorResponse, error)
}

type AuthorService struct {
	repo        repository.AuthorRepositoryInterface
	log         *zap.Logger
	userService user.UserServiceClient
}

func NewAuthorService(repo repository.AuthorRepositoryInterface, log *zap.Logger, userService user.UserServiceClient) AuthorServiceInterface {
	return &AuthorService{
		repo:        repo,
		log:         log,
		userService: userService,
	}
}

func (s *AuthorService) CreateAuthor(ctx context.Context, body *author.Author) (*author.CommonAuthorResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "missing outgoing metadata")
	}
	outbondCtx := metadata.NewOutgoingContext(ctx, md)

	user, err := s.userService.GetUser(outbondCtx, &emptypb.Empty{})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid user")
	}

	id := uuid.NewString()
	data := &model.Author{
		ID:        id,
		Name:      body.Name,
		CreatedBy: user.Id,
	}

	if err := s.repo.Create(data); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := fmt.Sprintf("create new author successfully with id %v", id)

	return &author.CommonAuthorResponse{Message: response}, nil
}

func (s *AuthorService) GetAuthor(ctx context.Context, body *author.Author) (*author.Author, error) {
	if body.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id cannot be empty")
	}

	data, err := s.repo.GetById(body.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &author.Author{
		Id:        data.ID,
		Name:      data.Name,
		CreatedBy: data.CreatedBy,
		CreatedAt: data.CreatedAt.String(),
		UpdatedAt: data.UpdatedAt.String(),
	}

	return res, nil
}

func (s *AuthorService) GetAuthors(ctx context.Context, body *author.AuthorRequest) (*author.AuthorsResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "missing outgoing metadata")
	}
	outbondCtx := metadata.NewOutgoingContext(ctx, md)

	_, err := s.userService.GetUser(outbondCtx, &emptypb.Empty{})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid user")
	}

	data, err := s.repo.Get(body.Search)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	authors := []*author.Author{}

	if len(data) > 0 {
		for _, v := range data {
			temp := &author.Author{
				Id:        v.ID,
				Name:      v.Name,
				CreatedBy: v.CreatedBy,
				CreatedAt: v.CreatedAt.String(),
				UpdatedAt: v.UpdatedAt.String(),
			}

			authors = append(authors, temp)
		}
	}

	return &author.AuthorsResponse{Authors: authors}, nil
}

func (s *AuthorService) UpdateAuthor(ctx context.Context, body *author.Author) (*author.CommonAuthorResponse, error) {
	if body.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id cannot be empty")
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "missing outgoing metadata")
	}
	outbondCtx := metadata.NewOutgoingContext(ctx, md)

	_, err := s.userService.GetUser(outbondCtx, &emptypb.Empty{})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid user")
	}

	_, err = s.repo.GetById(body.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	updateData := &model.Author{Name: body.Name}
	if err := s.repo.Update(updateData, body.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &author.CommonAuthorResponse{Message: "update author successfully"}, nil
}

func (s *AuthorService) DeleteAuthor(ctx context.Context, body *author.Author) (*author.CommonAuthorResponse, error) {
	if body.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id cannot be empty")
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "missing outgoing metadata")
	}
	outbondCtx := metadata.NewOutgoingContext(ctx, md)

	_, err := s.userService.GetUser(outbondCtx, &emptypb.Empty{})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid user")
	}

	_, err = s.repo.GetById(body.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := s.repo.Delete(body.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &author.CommonAuthorResponse{Message: "delete author successfully"}, nil
}
