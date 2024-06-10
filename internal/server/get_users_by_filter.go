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

func (s *server) GetUserByFilter(
	ctx context.Context,
	req *desc.GetUserByFilterRequest,
) (*desc.GetUserByFilterResponse, error) {
	h, err := newGetUserByFilterHandler(ctx, s.clients, req)
	if err != nil {
		return nil, err
	}
	err = h.handle()
	return h.response(), err
}

func (h *getUserByFilterHandler) handle() error {
	var err error
	if h.socialNetwork == nil {
		h.tgUsers, err = h.getUsersFromTelegram()
		if err != nil {
			return err
		}
		h.vkUsers, err = h.getUsersFromVk()
		if err != nil {
			return err
		}
	} else {
		switch *h.socialNetwork {
		case desc.SocialNetwork_TELEGRAM.String():
			h.tgUsers, err = h.getUsersFromTelegram()
		case desc.SocialNetwork_VK.String():
			h.vkUsers, err = h.getUsersFromVk()
		}
	}

	if err != nil {
		return err
	}
	return nil
}

func (h *getUserByFilterHandler) getUsersFromTelegram() ([]*desc.TelegramUser, error) {
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

	resp, err := h.clients.TelegramClient().GetUsersByFilter(h.ctx, &telegram.GetUsersByFilterRequest{
		TelegramId:             h.telegramId,
		UserRole:               role,
		UserNotificationStatus: notificationStatus,
		Group:                  h.group,
		Firstname:              h.firstname,
		Surname:                h.surname,
		Patronymic:             h.patronymic,
		MobilePhone:            h.mobilePhone,
		Limit:                  h.limit,
		Offset:                 h.offset,
	})
	if err != nil {
		return nil, err
	}

	return h.adaptTelegramUsers(resp.Users), nil
}

func (h *getUserByFilterHandler) adaptTelegramUsers(tgUsers []*telegram.User) []*desc.TelegramUser {
	users := make([]*desc.TelegramUser, 0, len(tgUsers))
	for _, tgUser := range tgUsers {
		users = append(users, &desc.TelegramUser{
			UserId:                 tgUser.UserId,
			TelegramId:             tgUser.TelegramId,
			UserRole:               desc.UserRole(desc.UserRole_value[tgUser.UserRole.String()]),
			UserNotificationStatus: desc.UserNotificationStatus(desc.UserNotificationStatus_value[tgUser.UserNotificationStatus.String()]),
			Group:                  tgUser.Group,
			Fio: &desc.FIO{
				Firstname:  tgUser.Fio.Firstname,
				Surname:    tgUser.Fio.Surname,
				Patronymic: tgUser.Fio.Patronymic,
			},
			MobilePhone: tgUser.MobilePhone,
			UserStatus:  desc.UserStatus(desc.UserStatus_value[tgUser.UserStatus.String()]),
		})
	}
	return users
}

func (h *getUserByFilterHandler) getUsersFromVk() ([]*desc.VkUser, error) {
	var role *vk.UserRole
	if h.userRole != nil {
		r := vk.UserRole(vk.UserRole_value[*h.userRole])
		role = &r
	}

	resp, err := h.clients.VkClient().GetUsersByFilter(h.ctx, &vk.GetUsersByFilterRequest{
		VkId:        h.vkId,
		UserRole:    role,
		Group:       h.group,
		MobilePhone: h.mobilePhone,
		Firstname:   h.firstname,
		Surname:     h.surname,
		Patronymic:  h.patronymic,
		Limit:       h.limit,
		Offset:      h.offset,
	})
	if err != nil {
		return nil, err
	}

	return h.adaptVkUsers(resp.GetUsers()), nil
}

func (h *getUserByFilterHandler) adaptVkUsers(vkUsers []*vk.User) []*desc.VkUser {
	users := make([]*desc.VkUser, 0, len(vkUsers))

	for _, vkUser := range vkUsers {
		users = append(users, &desc.VkUser{
			UserId:   vkUser.UserId,
			VkId:     vkUser.VkId,
			UserRole: desc.UserRole(desc.UserRole_value[vkUser.UserRole.String()]),
			Group:    vkUser.Group,
			Fio: &desc.FIO{
				Firstname:  vkUser.Fio.Firstname,
				Surname:    vkUser.Fio.Surname,
				Patronymic: vkUser.Fio.Patronymic,
			},
			MobilePhone: vkUser.MobilePhone,
			UserStatus:  desc.UserStatus(desc.UserStatus_value[vkUser.UserStatus.String()]),
		})
	}
	return users
}

func (h *getUserByFilterHandler) response() *desc.GetUserByFilterResponse {
	return &desc.GetUserByFilterResponse{
		TelegramUser: h.tgUsers,
		VkUser:       h.vkUsers,
		Limit:        h.limit,
		Offset:       h.offset,
	}
}

type getUserByFilterHandler struct {
	ctx     context.Context
	clients clients.Clients

	socialNetwork          *string
	telegramId             *int64
	vkId                   *int64
	userRole               *string
	userNotificationStatus *string
	group                  *string
	firstname              *string
	surname                *string
	patronymic             *string
	mobilePhone            *string
	limit                  int64
	offset                 int64

	tgUsers []*desc.TelegramUser
	vkUsers []*desc.VkUser
}

func newGetUserByFilterHandler(
	ctx context.Context,
	clients clients.Clients,
	req *desc.GetUserByFilterRequest,
) (*getUserByFilterHandler, error) {
	h := &getUserByFilterHandler{
		ctx:     ctx,
		clients: clients,
	}
	return h.adapt(req), h.validate()
}

func (h *getUserByFilterHandler) adapt(
	req *desc.GetUserByFilterRequest,
) *getUserByFilterHandler {
	if req.SocialNetwork != nil {
		socialNetwork := req.GetSocialNetwork().String()
		h.socialNetwork = &socialNetwork
	}
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
	h.limit = req.GetLimit()
	h.offset = req.GetOffset()
	return h
}

func (h *getUserByFilterHandler) validate() error {
	if h.limit <= 0 {
		return errors.
			NewNetworkError(codes.InvalidArgument, "limit must be specified").
			ToGRPCError()
	}
	if h.offset < 0 {
		return errors.
			NewNetworkError(codes.InvalidArgument, "offset must greater or equal zero").
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
