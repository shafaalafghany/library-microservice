package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shafaalafghany/book-service/model"
	"github.com/shafaalafghany/book-service/repository"
	"gitlab.com/shafaalafghany/synapsis-proto/go/author"
	"gitlab.com/shafaalafghany/synapsis-proto/go/book"
	"gitlab.com/shafaalafghany/synapsis-proto/go/category"
	"gitlab.com/shafaalafghany/synapsis-proto/go/user"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type BookServiceInterface interface {
	CreateBook(context.Context, *book.Book) (*book.CommonBookResponse, error)
	GetBook(context.Context, *book.Book) (*book.Book, error)
	GetBooks(context.Context, *book.BookRequest) (*book.BooksResponse, error)
	UpdateBook(context.Context, *book.Book) (*book.CommonBookResponse, error)
	DeleteBook(context.Context, *book.Book) (*book.CommonBookResponse, error)

	BorrowBook(context.Context, *book.BorrowRecord) (*book.CommonBorrowRecordResponse, error)
	ReturnBook(context.Context, *book.BorrowRecord) (*book.CommonBorrowRecordResponse, error)
}

type BookService struct {
	repo        repository.BookRepositoryInterface
	log         *zap.Logger
	userSvc     user.UserServiceClient
	authorSvc   author.AuthorServiceClient
	categorySvc category.CategoryServiceClient
}

func NewBookService(repo repository.BookRepositoryInterface, log *zap.Logger, userSvc user.UserServiceClient, authorSvc author.AuthorServiceClient, categorySvc category.CategoryServiceClient) BookServiceInterface {
	return &BookService{
		repo:        repo,
		log:         log,
		userSvc:     userSvc,
		authorSvc:   authorSvc,
		categorySvc: categorySvc,
	}
}

func (s *BookService) CreateBook(ctx context.Context, body *book.Book) (*book.CommonBookResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "missing outgoing metadata")
	}
	outbondCtx := metadata.NewOutgoingContext(ctx, md)

	userData, err := s.userSvc.GetUser(outbondCtx, &emptypb.Empty{})
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid user")
	}

	authorData, err := s.authorSvc.Get(outbondCtx, &author.Author{Id: body.GetAuthorId()})
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid author")
	}

	categoryData, err := s.categorySvc.Get(outbondCtx, &category.Category{Id: body.GetCategoryId()})
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid category")
	}

	id := uuid.NewString()
	data := &model.Book{
		ID:         id,
		Name:       body.GetName(),
		CreatedBy:  userData.GetId(),
		AuthorID:   authorData.GetId(),
		CategoryID: categoryData.GetId(),
		IsBorrowed: false,
		Borrows:    0,
	}

	if err := s.repo.Create(data); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := fmt.Sprintf("create new book successfully with id %v", id)
	return &book.CommonBookResponse{Message: response}, nil
}

func (s *BookService) GetBook(ctx context.Context, body *book.Book) (*book.Book, error) {
	if body.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id cannot be empty")
	}

	data, err := s.repo.GetById(ctx, &model.Book{ID: body.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &book.Book{
		Id:         data.ID,
		Name:       data.Name,
		AuthorId:   data.AuthorID,
		CategoryId: data.CategoryID,
		IsBorrowed: data.IsBorrowed,
		Borrows:    int32(data.Borrows),
		CreatedBy:  data.CreatedBy,
		CreatedAt:  data.CreatedAt.String(),
		UpdatedAt:  data.UpdatedAt.String(),
	}

	return res, nil
}

func (s *BookService) GetBooks(ctx context.Context, body *book.BookRequest) (*book.BooksResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "missing outgoing metadata")
	}
	outbondCtx := metadata.NewOutgoingContext(ctx, md)

	_, err := s.userSvc.GetUser(outbondCtx, &emptypb.Empty{})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid user")
	}

	data, err := s.repo.Get(ctx, body.GetSearch())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	books := []*book.Book{}

	if len(data) > 0 {
		for _, v := range data {
			temp := &book.Book{
				Id:         v.ID,
				Name:       v.Name,
				AuthorId:   v.AuthorID,
				CategoryId: v.CategoryID,
				IsBorrowed: v.IsBorrowed,
				Borrows:    int32(v.Borrows),
				CreatedBy:  v.CreatedBy,
				CreatedAt:  v.CreatedAt.String(),
				UpdatedAt:  v.UpdatedAt.String(),
			}

			books = append(books, temp)
		}
	}

	return &book.BooksResponse{Books: books}, nil
}

func (s *BookService) UpdateBook(ctx context.Context, body *book.Book) (*book.CommonBookResponse, error) {
	if body.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id cannot be empty")
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "missing outgoing metadata")
	}
	outbondCtx := metadata.NewOutgoingContext(ctx, md)

	_, err := s.userSvc.GetUser(outbondCtx, &emptypb.Empty{})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid user")
	}

	_, err = s.repo.GetById(ctx, &model.Book{ID: body.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	updateData := &model.Book{
		Name:       body.GetName(),
		AuthorID:   body.GetAuthorId(),
		CategoryID: body.GetCategoryId(),
	}
	if err := s.repo.Update(ctx, updateData, body.GetId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &book.CommonBookResponse{Message: "update book successfully"}, nil
}

func (s *BookService) DeleteBook(ctx context.Context, body *book.Book) (*book.CommonBookResponse, error) {
	if body.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id cannot be empty")
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "missing outgoing metadata")
	}
	outbondCtx := metadata.NewOutgoingContext(ctx, md)

	_, err := s.userSvc.GetUser(outbondCtx, &emptypb.Empty{})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid user")
	}

	_, err = s.repo.GetById(ctx, &model.Book{ID: body.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := s.repo.Delete(ctx, body.GetId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &book.CommonBookResponse{Message: "delete book successfully"}, nil
}

func (s *BookService) BorrowBook(ctx context.Context, body *book.BorrowRecord) (*book.CommonBorrowRecordResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "missing outgoing metadata")
	}
	outbondCtx := metadata.NewOutgoingContext(ctx, md)

	_, err := s.userSvc.GetUser(outbondCtx, &emptypb.Empty{})
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid user")
	}

	if _, err = s.repo.GetById(ctx, &model.Book{ID: body.GetBookId()}); err != nil {
		return nil, status.Error(codes.NotFound, "book not found")
	}

	borrowRecord := &model.BorrowRecord{
		ID:         uuid.NewString(),
		BookID:     body.GetBookId(),
		UserID:     body.GetUserId(),
		BorrowedAt: time.Now(),
	}

	if err = s.repo.Borrow(ctx, borrowRecord); err != nil {
		return nil, err
	}

	return &book.CommonBorrowRecordResponse{Message: "borrow book successfully"}, nil
}

func (s *BookService) ReturnBook(ctx context.Context, body *book.BorrowRecord) (*book.CommonBorrowRecordResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "missing outgoing metadata")
	}
	outbondCtx := metadata.NewOutgoingContext(ctx, md)

	_, err := s.userSvc.GetUser(outbondCtx, &emptypb.Empty{})
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid user")
	}

	if _, err = s.repo.GetById(ctx, &model.Book{ID: body.GetBookId()}); err != nil {
		return nil, status.Error(codes.NotFound, "book not found")
	}

	borrowRecord := &model.BorrowRecord{
		BookID: body.GetBookId(),
		UserID: body.GetUserId(),
	}

	if err = s.repo.ReturnBook(ctx, borrowRecord); err != nil {
		return nil, err
	}

	return &book.CommonBorrowRecordResponse{Message: "return book successfully"}, nil
}
