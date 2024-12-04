package handler

import (
	"context"

	"github.com/shafaalafghany/user-service/service"
	"gitlab.com/shafaalafghany/synapsis-proto/go/user"
	"go.uber.org/zap"
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
