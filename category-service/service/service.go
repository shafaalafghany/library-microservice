package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shafaalafghany/category-service/model"
	"github.com/shafaalafghany/category-service/repository"
	"gitlab.com/shafaalafghany/synapsis-proto/go/category"
	"gitlab.com/shafaalafghany/synapsis-proto/go/user"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CategoryServiceInterface interface {
	CreateCategory(context.Context, *category.Category) (*category.CommonCategoryResponse, error)
	GetCategory(context.Context, *category.Category) (*category.Category, error)
	GetCategories(context.Context, *category.CategoryRequest) (*category.CategoriesResponse, error)
}

type CategoryService struct {
	repo        repository.CategoryRepositoryInterface
	log         *zap.Logger
	userService user.UserServiceClient
}

func NewCategoryService(repo repository.CategoryRepositoryInterface, log *zap.Logger, userService user.UserServiceClient) CategoryServiceInterface {
	return &CategoryService{
		repo:        repo,
		log:         log,
		userService: userService,
	}
}

func (cs *CategoryService) CreateCategory(ctx context.Context, body *category.Category) (*category.CommonCategoryResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "missing outgoing metadata")
	}
	outbondCtx := metadata.NewOutgoingContext(ctx, md)

	user, err := cs.userService.GetUser(outbondCtx, &emptypb.Empty{})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid user")
	}

	id := uuid.NewString()
	data := &model.Category{
		ID:        id,
		Name:      body.GetName(),
		CreatedBy: user.Id,
	}

	if err := cs.repo.Create(data); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := fmt.Sprintf("create new category %s successfully with id %v", body.GetName(), id)

	return &category.CommonCategoryResponse{Message: response}, nil
}

func (cs *CategoryService) GetCategory(ctx context.Context, body *category.Category) (*category.Category, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "missing outgoing metadata")
	}
	outbondCtx := metadata.NewOutgoingContext(ctx, md)

	_, err := cs.userService.GetUser(outbondCtx, &emptypb.Empty{})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid user")
	}

	data, err := cs.repo.GetById(body.GetId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	body.Name = data.Name
	body.CreatedBy = data.CreatedBy
	body.CreatedAt = data.CreatedAt.String()
	body.UpdatedAt = data.UpdatedAt.String()

	return body, nil
}

func (cs *CategoryService) GetCategories(ctx context.Context, body *category.CategoryRequest) (*category.CategoriesResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "missing outgoing metadata")
	}
	outbondCtx := metadata.NewOutgoingContext(ctx, md)

	_, err := cs.userService.GetUser(outbondCtx, &emptypb.Empty{})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid user")
	}

	data, err := cs.repo.Get(body.GetSearch())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	categories := []*category.Category{}

	if len(data) > 0 {
		for _, v := range data {
			temp := &category.Category{
				Id:        v.ID,
				Name:      v.Name,
				CreatedBy: v.CreatedBy,
				CreatedAt: v.CreatedAt.String(),
				UpdatedAt: v.UpdatedAt.String(),
			}

			categories = append(categories, temp)
		}
	}

	return &category.CategoriesResponse{Categories: categories}, nil
}
