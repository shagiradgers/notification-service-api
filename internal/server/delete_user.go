package server

import (
	"context"

	"google.golang.org/grpc/codes"
	desc "notification-service-api/api"
	"notification-service-api/internal/clients"
	"notification-service-api/internal/errors"
	"notification-service-api/pb/tg"
	"notification-service-api/pb/vk"
)

func (s *server) DeleteUser(
	ctx context.Context,
	req *desc.DeleteUserRequest,
) (*desc.DeleteUserResponse, error) {
	h, err := newDeleteUserHandler(ctx, s.clients, req)
	if err != nil {
		return nil, err
	}
	err = h.handle()
	return &desc.DeleteUserResponse{}, err
}

func (h *deleteUserHandler) handle() error {
	var err error
	switch h.socialNetwork {
	case desc.SocialNetwork_TELEGRAM.String():
		err = h.deleteFromTelegram()
	case desc.SocialNetwork_VK.String():
		err = h.deleteFromVk()
	}
	return err
}

func (h *deleteUserHandler) deleteFromVk() error {
	_, err := h.clients.VkClient().DeleteUser(h.ctx, &vk.DeleteUserRequest{UserId: h.userId})
	return err
}

func (h *deleteUserHandler) deleteFromTelegram() error {
	_, err := h.clients.TelegramClient().DeleteUser(h.ctx, &telegram.DeleteUserRequest{UserId: h.userId})
	return err
}

type deleteUserHandler struct {
	ctx     context.Context
	clients clients.Clients

	userId        int64
	socialNetwork string
}

func newDeleteUserHandler(
	ctx context.Context,
	clients clients.Clients,
	req *desc.DeleteUserRequest,
) (*deleteUserHandler, error) {
	h := &deleteUserHandler{
		ctx:           ctx,
		clients:       clients,
		userId:        req.GetUserId(),
		socialNetwork: req.GetSocialNetwork().String(),
	}
	return h, h.validate()
}

func (h *deleteUserHandler) validate() error {
	if h.userId <= 0 {
		return errors.
			NewNetworkError(codes.InvalidArgument, "user_id must be specified").
			ToGRPCError()
	}
	return nil
}
