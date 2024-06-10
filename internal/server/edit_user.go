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

func (s *server) EditUser(
	ctx context.Context,
	req *desc.EditUserRequest,
) (*desc.EditUserResponse, error) {
	h, err := newEditUserHandler(ctx, s.clients, req)
	if err != nil {
		return nil, err
	}
	err = h.handle()
	return h.response(), err
}

func (h *editUserHandler) handle() error {
	var err error
	switch h.socialNetwork {
	case desc.SocialNetwork_VK.String():
		err = h.editInVk()
	case desc.SocialNetwork_TELEGRAM.String():
		err = h.editInTelegram()
	}
	return err
}

func (h *editUserHandler) editInVk() error {
	var role *vk.UserRole
	if h.userRole != nil {
		r := vk.UserRole(vk.UserRole_value[*h.userRole])
		role = &r
	}

	u, err := h.clients.VkClient().EditUser(h.ctx, &vk.EditUserRequest{
		UserId:      h.userId,
		VkId:        h.vkId,
		UserRole:    role,
		Group:       h.group,
		Firstname:   h.firstname,
		Surname:     h.surname,
		Patronymic:  h.patronymic,
		MobilePhone: h.mobilePhone,
	})

	if err != nil {
		return err
	}
	if u == nil {
		return errors.NewNetworkError(codes.Internal, "got nil user from vk")
	}
	h.updatedUserVk = u.User
	return nil
}

func (h *editUserHandler) editInTelegram() error {
	var role *telegram.UserRole
	var notificationStatus *telegram.UserNotificationStatus

	if h.userRole != nil {
		r := telegram.UserRole(telegram.UserRole_value[*h.userRole])
		role = &r
	}
	if h.userNotificationStatus != nil {
		n := telegram.UserNotificationStatus(telegram.UserNotificationStatus_value[*h.userNotificationStatus])
		notificationStatus = &n
	}

	u, err := h.clients.TelegramClient().EditUser(h.ctx, &telegram.EditUserRequest{
		UserId:                 h.userId,
		TelegramId:             h.telegramId,
		UserRole:               role,
		UserNotificationStatus: notificationStatus,
		Group:                  h.group,
		Firstname:              h.firstname,
		Surname:                h.surname,
		Patronymic:             h.patronymic,
		MobilePhone:            h.mobilePhone,
	})

	if err != nil {
		return err
	}
	if u == nil {
		return errors.NewNetworkError(codes.Internal, "got nil user from telegram")
	}
	h.updatedUserTelegram = u.User
	return nil
}

func (h *editUserHandler) response() *desc.EditUserResponse {
	if h.updatedUserVk != nil {
		return &desc.EditUserResponse{
			User: &desc.EditUserResponse_VkUser{
				VkUser: &desc.VkUser{
					UserId:   h.updatedUserVk.UserId,
					VkId:     h.updatedUserVk.VkId,
					UserRole: desc.UserRole(desc.UserRole_value[h.updatedUserVk.UserRole.String()]),
					Group:    h.updatedUserVk.Group,
					Fio: &desc.FIO{
						Firstname:  h.updatedUserVk.Fio.Firstname,
						Surname:    h.updatedUserVk.Fio.Surname,
						Patronymic: h.updatedUserVk.Fio.Patronymic,
					},
					MobilePhone: h.updatedUserVk.MobilePhone,
					UserStatus:  desc.UserStatus(desc.UserStatus_value[h.updatedUserVk.UserStatus.String()]),
				},
			},
		}
	}
	if h.updatedUserTelegram != nil {
		return &desc.EditUserResponse{
			User: &desc.EditUserResponse_TelegramUser{
				TelegramUser: &desc.TelegramUser{
					UserId:                 h.updatedUserTelegram.UserId,
					TelegramId:             h.updatedUserTelegram.TelegramId,
					UserRole:               desc.UserRole(desc.UserRole_value[h.updatedUserTelegram.UserRole.String()]),
					UserNotificationStatus: desc.UserNotificationStatus(desc.UserNotificationStatus_value[h.updatedUserTelegram.UserNotificationStatus.String()]),
					Group:                  h.updatedUserTelegram.Group,
					Fio: &desc.FIO{
						Firstname:  h.updatedUserTelegram.Fio.Firstname,
						Surname:    h.updatedUserTelegram.Fio.Surname,
						Patronymic: h.updatedUserTelegram.Fio.Patronymic,
					},
					MobilePhone: h.updatedUserTelegram.MobilePhone,
					UserStatus:  desc.UserStatus(desc.UserStatus_value[h.updatedUserTelegram.UserStatus.String()]),
				},
			},
		}
	}
	return nil
}

type editUserHandler struct {
	ctx     context.Context
	clients clients.Clients

	userId                 int64
	socialNetwork          string
	telegramId             *int64
	vkId                   *int64
	userRole               *string
	userNotificationStatus *string
	group                  *string
	firstname              *string
	surname                *string
	patronymic             *string
	mobilePhone            *string

	updatedUserTelegram *telegram.User
	updatedUserVk       *vk.User
}

func newEditUserHandler(
	ctx context.Context,
	clients clients.Clients,
	req *desc.EditUserRequest,
) (*editUserHandler, error) {
	h := editUserHandler{
		ctx:     ctx,
		clients: clients,
	}
	return h.adapt(req), h.validate()
}

func (h *editUserHandler) adapt(
	req *desc.EditUserRequest,
) *editUserHandler {
	h.userId = req.GetUserId()
	h.socialNetwork = req.GetSocialNetwork().String()

	if req.TelegramId != nil {
		h.telegramId = req.TelegramId
	}
	if req.VkId != nil {
		h.vkId = req.VkId
	}
	if req.UserRole != nil {
		role := req.GetUserRole().String()
		h.userRole = &role
	}
	if req.UserNotificationStatus != nil {
		notification := req.GetUserNotificationStatus().String()
		h.userNotificationStatus = &notification
	}
	if req.Group != nil {
		h.group = req.Group
	}
	if req.Firstname != nil {
		h.firstname = req.Firstname
	}
	if req.Surname != nil {
		h.surname = req.Surname
	}
	if req.Patronymic != nil {
		h.patronymic = req.Patronymic
	}
	if req.MobilePhone != nil {
		h.mobilePhone = req.MobilePhone
	}
	return h
}

func (h *editUserHandler) validate() error {
	if h.userId <= 0 {
		return errors.
			NewNetworkError(codes.InvalidArgument, "user_id must be specified").
			ToGRPCError()
	}
	if h.telegramId != nil && *h.telegramId <= 0 {
		return errors.
			NewNetworkError(codes.InvalidArgument, "telegram_id must be specified").
			ToGRPCError()
	}
	if h.vkId != nil && *h.vkId <= 0 {
		return errors.
			NewNetworkError(codes.InvalidArgument, "vk_id must be specified").
			ToGRPCError()
	}
	if h.group != nil && *h.group == "" {
		return errors.
			NewNetworkError(codes.InvalidArgument, "group must be specified").
			ToGRPCError()
	}
	if h.firstname != nil && *h.firstname == "" {
		return errors.
			NewNetworkError(codes.InvalidArgument, "firstname must be specified").
			ToGRPCError()
	}
	if h.surname != nil && *h.surname == "" {
		return errors.
			NewNetworkError(codes.InvalidArgument, "surname must be specified").
			ToGRPCError()
	}
	if h.patronymic != nil && *h.patronymic == "" {
		return errors.
			NewNetworkError(codes.InvalidArgument, "patronymic must be specified").
			ToGRPCError()
	}
	if h.mobilePhone != nil && *h.mobilePhone == "" {
		return errors.
			NewNetworkError(codes.InvalidArgument, "mobile_phone must be specified").
			ToGRPCError()
	}

	return nil
}
