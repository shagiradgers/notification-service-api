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

func (s *server) CreateUser(
	ctx context.Context,
	req *desc.CreateUserRequest,
) (*desc.CreateUserResponse, error) {
	h, err := newCreateUserHandler(ctx, s.clients, req)
	if err != nil {
		return nil, err
	}
	if err = h.handle(); err != nil {
		return nil, err
	}
	return h.response(), nil
}

func (h *createUserHandler) handle() error {
	var err error
	if h.tgUser != nil {
		err = h.createUserInTelegram()
	}
	if h.vkUser != nil {
		err = h.createUserInVk()
	}
	return err
}

func (h *createUserHandler) createUserInVk() error {
	if h.vkUser == nil {
		return errors.
			NewNetworkError(codes.Internal, "vk user is nil").
			ToGRPCError()
	}

	u, err := h.clients.VkClient().CreateUser(h.ctx, &vk.CreateUserRequest{
		VkId:     h.vkUser.VkId,
		UserRole: vk.UserRole(vk.UserRole_value[h.vkUser.UserRole]),
		Group:    h.vkUser.Group,
		Fio: &vk.FIO{
			Firstname:  h.vkUser.Firstname,
			Surname:    h.vkUser.Surname,
			Patronymic: h.vkUser.Patronymic,
		},
		MobilePhone: h.vkUser.MobilePhone,
	})
	if err != nil {
		return err
	}
	if u == nil {
		return errors.NewNetworkError(codes.Internal, "got nil user from vk")
	}

	h.createUserVk = u.GetUser()
	return nil
}

func (h *createUserHandler) createUserInTelegram() error {
	if h.tgUser == nil {
		return errors.
			NewNetworkError(codes.Internal, "telegram user is nil").
			ToGRPCError()
	}

	u, err := h.clients.TelegramClient().CreateUser(h.ctx, &telegram.CreateUserRequest{
		TelegramId:             h.tgUser.TelegramId,
		UserRole:               telegram.UserRole(telegram.UserRole_value[h.tgUser.UserRole]),
		UserNotificationStatus: telegram.UserNotificationStatus(telegram.UserNotificationStatus_value[h.tgUser.UserNotificationStatus]),
		Group:                  h.tgUser.Group,
		Fio: &telegram.FIO{
			Firstname:  h.tgUser.Firstname,
			Surname:    h.tgUser.Surname,
			Patronymic: h.tgUser.Patronymic,
		},
		MobilePhone: h.tgUser.MobilePhone,
	})

	if err != nil {
		return err
	}
	if u == nil {
		return errors.NewNetworkError(codes.Internal, "got nil user from telegram")
	}
	h.createdUserTg = u.GetUser()
	return err
}

func (h *createUserHandler) response() *desc.CreateUserResponse {
	if h.createUserVk != nil {
		return &desc.CreateUserResponse{
			User: &desc.CreateUserResponse_VkUser{
				VkUser: &desc.VkUser{
					UserId:   h.createUserVk.UserId,
					VkId:     h.createUserVk.VkId,
					UserRole: desc.UserRole(desc.UserRole_value[h.createUserVk.UserRole.String()]),
					Group:    h.createUserVk.Group,
					Fio: &desc.FIO{
						Firstname:  h.createUserVk.Fio.Firstname,
						Surname:    h.createUserVk.Fio.Surname,
						Patronymic: h.createUserVk.Fio.Patronymic,
					},
					MobilePhone: h.createUserVk.MobilePhone,
					UserStatus:  desc.UserStatus(desc.UserStatus_value[h.createUserVk.UserStatus.String()]),
				},
			},
		}
	}
	if h.createdUserTg != nil {
		return &desc.CreateUserResponse{
			User: &desc.CreateUserResponse_TelegramUser{
				TelegramUser: &desc.TelegramUser{
					UserId:                 h.createdUserTg.UserId,
					TelegramId:             h.createdUserTg.TelegramId,
					UserRole:               desc.UserRole(desc.UserRole_value[h.createdUserTg.UserRole.String()]),
					UserNotificationStatus: desc.UserNotificationStatus(desc.UserNotificationStatus_value[h.createdUserTg.UserNotificationStatus.String()]),
					Group:                  h.createdUserTg.Group,
					Fio: &desc.FIO{
						Firstname:  h.createdUserTg.Fio.Firstname,
						Surname:    h.createdUserTg.Fio.Surname,
						Patronymic: h.createdUserTg.Fio.Patronymic,
					},
					MobilePhone: h.createdUserTg.MobilePhone,
					UserStatus:  desc.UserStatus(desc.UserStatus_value[h.createdUserTg.UserStatus.String()]),
				},
			},
		}
	}
	return nil
}

type createUserHandler struct {
	ctx     context.Context
	clients clients.Clients

	tgUser *createUserHandlerTelegram
	vkUser *createUserHandlerVk

	createdUserTg *telegram.User
	createUserVk  *vk.User
}

func newCreateUserHandler(
	ctx context.Context,
	clients clients.Clients,
	req *desc.CreateUserRequest,
) (*createUserHandler, error) {
	h := &createUserHandler{
		ctx:     ctx,
		clients: clients,
	}
	return h.adapt(req), h.validate()
}

func (h *createUserHandler) validate() error {
	var err error
	if h.vkUser != nil {
		err = h.validateVkUser()
	}
	if h.tgUser != nil {
		err = h.validateTelegramUser()
	}
	return err
}

func (h *createUserHandler) validateVkUser() error {
	if h.vkUser == nil {
		return errors.
			NewNetworkError(codes.Internal, "vk user is nil").
			ToGRPCError()
	}

	if h.vkUser.VkId <= 0 {
		return errors.
			NewNetworkError(codes.InvalidArgument, "vk_id must be specified").
			ToGRPCError()
	}
	if h.vkUser.Group == "" {
		return errors.
			NewNetworkError(codes.InvalidArgument, "group must be specified").
			ToGRPCError()
	}
	if h.vkUser.Firstname == "" {
		return errors.
			NewNetworkError(codes.InvalidArgument, "firstname must be specified").
			ToGRPCError()
	}
	if h.vkUser.Surname == "" {
		return errors.
			NewNetworkError(codes.InvalidArgument, "surname must be specified").
			ToGRPCError()
	}
	if h.vkUser.Patronymic != nil && *h.vkUser.Patronymic == "" {
		return errors.
			NewNetworkError(codes.InvalidArgument, "patronymic must be specified").
			ToGRPCError()
	}
	if h.vkUser.MobilePhone == "" {
		return errors.
			NewNetworkError(codes.InvalidArgument, "mobile phone must be specified").
			ToGRPCError()
	}
	return nil
}

func (h *createUserHandler) validateTelegramUser() error {
	if h.tgUser == nil {
		return errors.
			NewNetworkError(codes.Internal, "telegram user is nil").
			ToGRPCError()
	}

	if h.tgUser.TelegramId <= 0 {
		return errors.
			NewNetworkError(codes.InvalidArgument, "telegram id must be specified").
			ToGRPCError()
	}
	if h.tgUser.Group == "" {
		return errors.
			NewNetworkError(codes.InvalidArgument, "group must be specified").
			ToGRPCError()
	}
	if h.tgUser.Firstname == "" {
		return errors.
			NewNetworkError(codes.InvalidArgument, "firstname must be specified").
			ToGRPCError()
	}
	if h.tgUser.Surname == "" {
		return errors.
			NewNetworkError(codes.InvalidArgument, "surname must be specified").
			ToGRPCError()
	}
	if h.tgUser.Patronymic != nil && *h.tgUser.Patronymic == "" {
		return errors.
			NewNetworkError(codes.InvalidArgument, "patronymic must be specified").
			ToGRPCError()
	}
	if h.tgUser.MobilePhone == "" {
		return errors.
			NewNetworkError(codes.InvalidArgument, "mobile phone must be specified").
			ToGRPCError()
	}
	return nil
}

func (h *createUserHandler) adapt(req *desc.CreateUserRequest) *createUserHandler {
	if req.GetVkUser() != nil {
		h.vkUser = &createUserHandlerVk{
			VkId:        req.GetVkUser().GetVkId(),
			UserRole:    req.GetVkUser().GetUserRole().String(),
			Group:       req.GetVkUser().GetGroup(),
			MobilePhone: req.GetVkUser().GetMobilePhone(),
			Firstname:   req.GetVkUser().GetFio().GetFirstname(),
			Surname:     req.GetVkUser().GetFio().GetSurname(),
			Patronymic:  req.GetVkUser().GetFio().Patronymic,
		}
	}
	if req.GetTelegramUser() != nil {
		h.tgUser = &createUserHandlerTelegram{
			TelegramId:             req.GetTelegramUser().GetTelegramId(),
			UserRole:               req.GetTelegramUser().GetUserRole().String(),
			UserNotificationStatus: req.GetTelegramUser().GetUserNotificationStatus().String(),
			Group:                  req.GetTelegramUser().GetGroup(),
			MobilePhone:            req.GetTelegramUser().GetMobilePhone(),
			Firstname:              req.GetTelegramUser().GetFio().GetFirstname(),
			Surname:                req.GetTelegramUser().GetFio().GetSurname(),
			Patronymic:             req.GetTelegramUser().GetFio().Patronymic,
		}
	}
	return h
}

type createUserHandlerTelegram struct {
	TelegramId             int64
	UserRole               string
	UserNotificationStatus string
	Group                  string
	MobilePhone            string
	Firstname              string
	Surname                string
	Patronymic             *string
}

type createUserHandlerVk struct {
	VkId        int64
	UserRole    string
	Group       string
	MobilePhone string
	Firstname   string
	Surname     string
	Patronymic  *string
}
