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

func (s *server) SendNotification(
	ctx context.Context,
	req *desc.SendNotificationRequest,
) (*desc.SendNotificationResponse, error) {
	h, err := newSendNotificationHandler(ctx, s.clients, req)
	if err != nil {
		return nil, err
	}

	err = h.handle()
	return h.response(), err
}

func (h *sendNotificationHandler) handle() error {
	var err error
	switch h.SocialNetwork {
	case desc.SocialNetwork_TELEGRAM.String():
		err = h.sendToTg()
	case desc.SocialNetwork_VK.String():
		err = h.sendToVk()
	}
	return err
}

func (h *sendNotificationHandler) sendToVk() error {
	notification, err := h.clients.VkClient().SendNotification(h.ctx, &vk.SendNotificationRequest{
		SenderId:     h.SenderId,
		ReceiverIds:  h.ReceiverIds,
		Message:      h.Message,
		MediaContent: h.MediaContent,
	})
	if err != nil {
		return err
	}
	if notification == nil {
		return errors.
			NewNetworkError(codes.Internal, "vk: got null notification").
			ToGRPCError()
	}

	h.notificationStatus = notification.GetMessageStatus().String()
	h.notificationId = notification.GetNotificationId()
	return nil
}

func (h *sendNotificationHandler) sendToTg() error {
	notification, err := h.clients.TelegramClient().SendNotification(h.ctx, &telegram.SendNotificationRequest{
		SenderId:     h.SenderId,
		ReceiverIds:  h.ReceiverIds,
		Message:      h.Message,
		MediaContent: h.MediaContent,
	})
	if err != nil {
		return err
	}
	if notification == nil {
		return errors.
			NewNetworkError(codes.Internal, "telegram: got null notification").
			ToGRPCError()
	}

	h.notificationStatus = notification.GetMessageStatus().String()
	h.notificationId = notification.GetNotificationId()
	return nil
}

func (h *sendNotificationHandler) response() *desc.SendNotificationResponse {
	return &desc.SendNotificationResponse{
		NotificationId: h.notificationId,
		MessageStatus:  desc.NotificationStatus(desc.NotificationStatus_value[h.notificationStatus]),
	}
}

type sendNotificationHandler struct {
	ctx     context.Context
	clients clients.Clients

	SenderId      int64
	ReceiverIds   []int64
	Message       string
	MediaContent  *string
	SocialNetwork string

	notificationId     int64
	notificationStatus string
}

func newSendNotificationHandler(
	ctx context.Context,
	clients clients.Clients,
	req *desc.SendNotificationRequest,
) (*sendNotificationHandler, error) {
	h := &sendNotificationHandler{
		ctx:     ctx,
		clients: clients,
	}
	return h.adapt(req), h.validate()
}

func (h *sendNotificationHandler) adapt(
	req *desc.SendNotificationRequest,
) *sendNotificationHandler {
	h.SenderId = req.GetNotification().GetSenderId()
	h.ReceiverIds = req.GetNotification().GetReceiverIds()
	h.Message = req.GetNotification().GetMessage()
	h.MediaContent = req.GetNotification().MediaContent
	h.SocialNetwork = req.GetNotification().GetSocialNetwork().String()
	return h
}

func (h *sendNotificationHandler) validate() error {
	if h.SenderId <= 0 {
		return errors.
			NewNetworkError(codes.InvalidArgument, "sender_id must be specified").
			ToGRPCError()
	}
	if len(h.ReceiverIds) == 0 {
		return errors.
			NewNetworkError(codes.InvalidArgument, "receiver_ids must be specified").
			ToGRPCError()
	}
	if h.Message == "" {
		return errors.
			NewNetworkError(codes.InvalidArgument, "message must be specified").
			ToGRPCError()
	}
	if h.MediaContent != nil && *h.MediaContent == "" {
		return errors.
			NewNetworkError(codes.InvalidArgument, "wrong media_content").
			ToGRPCError()
	}
	return nil
}
