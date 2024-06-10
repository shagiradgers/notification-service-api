package server

import (
	"context"
	"google.golang.org/grpc/codes"
	desc "notification-service-api/api"
	"notification-service-api/internal/clients"
	"notification-service-api/internal/errors"
	telegram "notification-service-api/pb/tg"
	"notification-service-api/pb/vk"
)

func (s *server) GetUser(
	ctx context.Context,
	req *desc.GetUserRequest,
) (*desc.GetUserResponse, error) {
	h, err := newGetUserHandler(ctx, s.clients, req)
	if err != nil {
		return nil, err
	}
	if err = h.handle(); err != nil {
		return nil, err
	}
	return h.response(), nil
}

func (h *getUserHandler) handle() error {
	var err error
	switch h.socialNetwork {
	case desc.SocialNetwork_VK.String():
		err = h.getUserFromVk()
	case desc.SocialNetwork_TELEGRAM.String():
		err = h.getUserFromTelegram()
	}
	return err
}

func (h *getUserHandler) getUserFromVk() error {
	u, err := h.clients.VkClient().GetUser(h.ctx, &vk.GetUserRequest{UserId: h.userId})
	if err != nil {
		return err
	}
	if u == nil {
		return errors.NewNetworkError(codes.Internal, "got nil user from vk")
	}
	h.vkUser = u.GetUser()
	return nil
}

func (h *getUserHandler) getUserFromTelegram() error {
	u, err := h.clients.TelegramClient().GetUser(h.ctx, &telegram.GetUserRequest{UserId: h.userId})
	if err != nil {
		return err
	}
	if u == nil {
		return errors.NewNetworkError(codes.Internal, "got nil user from telegram")
	}
	h.telegramUser = u.GetUser()
	return nil
}

func (h *getUserHandler) response() *desc.GetUserResponse {
	if h.vkUser != nil {
		return &desc.GetUserResponse{
			User: &desc.GetUserResponse_VkUser{
				VkUser: &desc.VkUser{
					UserId:   h.vkUser.UserId,
					VkId:     h.vkUser.VkId,
					UserRole: desc.UserRole(desc.UserRole_value[h.vkUser.UserRole.String()]),
					Group:    h.vkUser.Group,
					Fio: &desc.FIO{
						Firstname:  h.vkUser.Fio.Firstname,
						Surname:    h.vkUser.Fio.Surname,
						Patronymic: h.vkUser.Fio.Patronymic,
					},
					MobilePhone: h.vkUser.MobilePhone,
					UserStatus:  desc.UserStatus(desc.UserStatus_value[h.vkUser.UserStatus.String()]),
				},
			},
		}
	}
	if h.telegramUser != nil {
		return &desc.GetUserResponse{
			User: &desc.GetUserResponse_TelegramUser{
				TelegramUser: &desc.TelegramUser{
					UserId:                 h.telegramUser.UserId,
					TelegramId:             h.telegramUser.TelegramId,
					UserRole:               desc.UserRole(desc.UserRole_value[h.telegramUser.UserRole.String()]),
					UserNotificationStatus: desc.UserNotificationStatus(desc.UserNotificationStatus_value[h.telegramUser.UserNotificationStatus.String()]),
					Group:                  h.telegramUser.Group,
					Fio: &desc.FIO{
						Firstname:  h.telegramUser.Fio.Firstname,
						Surname:    h.telegramUser.Fio.Surname,
						Patronymic: h.telegramUser.Fio.Patronymic,
					},
					MobilePhone: h.telegramUser.MobilePhone,
					UserStatus:  desc.UserStatus(desc.UserStatus_value[h.telegramUser.UserStatus.String()]),
				},
			},
		}
	}
	return nil
}

type getUserHandler struct {
	ctx     context.Context
	clients clients.Clients

	socialNetwork string
	userId        int64

	vkUser       *vk.User
	telegramUser *telegram.User
}

func newGetUserHandler(
	ctx context.Context,
	clients clients.Clients,
	req *desc.GetUserRequest,
) (*getUserHandler, error) {
	h := &getUserHandler{
		ctx:           ctx,
		clients:       clients,
		socialNetwork: req.GetSocialNetwork().String(),
		userId:        req.GetUserId(),
	}
	return h, h.validate()
}

func (h *getUserHandler) validate() error {
	if h.userId <= 0 {
		return errors.
			NewNetworkError(codes.InvalidArgument, "patronymic must be specified").
			ToGRPCError()
	}

	return nil
}
