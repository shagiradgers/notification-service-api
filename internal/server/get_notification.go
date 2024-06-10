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

func (s *server) GetNotificationData(
	ctx context.Context,
	req *desc.GetNotificationDataRequest,
) (*desc.GetNotificationDataResponse, error) {
	h, err := newGetNotificationDataHandler(ctx, s.clients, req)
	if err != nil {
		return nil, err
	}
	err = h.handle()

	return h.response(), err
}

func (h *getNotificationDataHandler) handle() error {
	var err error
	switch h.socialNetwork {
	case desc.SocialNetwork_VK.String():
		err = h.getNotificationFromVk()
	case desc.SocialNetwork_TELEGRAM.String():
		err = h.getNotificationFromTelegram()
	}
	return err
}

func (h *getNotificationDataHandler) getNotificationFromTelegram() error {
	n, err := h.clients.TelegramClient().GetNotification(h.ctx, &telegram.GetNotificationRequest{
		NotificationId: h.notificationId,
	})
	if err != nil {
		return err
	}

	h.notificationStatus = n.GetNotification().GetNotificationStatus().String()
	h.senderId = n.GetNotification().GetSenderId()
	h.receiverIds = n.GetNotification().GetReceiverIds()
	h.message = n.GetNotification().GetMessage()
	h.mediaContent = n.GetNotification().MediaContent

	return nil
}

func (h *getNotificationDataHandler) getNotificationFromVk() error {
	n, err := h.clients.VkClient().GetNotification(h.ctx, &vk.GetNotificationRequest{
		NotificationId: h.notificationId,
	})
	if err != nil {
		return err
	}

	h.notificationStatus = n.GetNotification().GetNotificationStatus().String()
	h.senderId = n.GetNotification().GetSenderId()
	h.receiverIds = n.GetNotification().GetReceiverIds()
	h.message = n.GetNotification().GetMessage()
	h.mediaContent = n.GetNotification().MediaContent

	return nil
}

func (h *getNotificationDataHandler) response() *desc.GetNotificationDataResponse {
	return &desc.GetNotificationDataResponse{
		NotificationStatus: desc.NotificationStatus(desc.NotificationStatus_value[h.notificationStatus]),
		Notification: &desc.Notification{
			SenderId:      h.senderId,
			ReceiverIds:   h.receiverIds,
			Message:       h.message,
			MediaContent:  h.mediaContent,
			SocialNetwork: desc.SocialNetwork(desc.SocialNetwork_value[h.socialNetwork]),
		},
	}
}

type getNotificationDataHandler struct {
	ctx     context.Context
	clients clients.Clients

	notificationId int64
	socialNetwork  string

	notificationStatus string
	senderId           int64
	receiverIds        []int64
	message            string
	mediaContent       *string
}

func newGetNotificationDataHandler(
	ctx context.Context,
	clients clients.Clients,
	req *desc.GetNotificationDataRequest,
) (*getNotificationDataHandler, error) {
	h := getNotificationDataHandler{
		ctx:     ctx,
		clients: clients,
	}
	return h.adapt(req), h.validate()
}

func (h *getNotificationDataHandler) adapt(
	req *desc.GetNotificationDataRequest,
) *getNotificationDataHandler {
	h.socialNetwork = req.GetSocialNetwork().String()
	h.notificationId = req.GetNotificationId()
	return h
}

func (h *getNotificationDataHandler) validate() error {
	if h.notificationId <= 0 {
		return errors.
			NewNetworkError(codes.InvalidArgument, "notification_id must be specified").
			ToGRPCError()
	}
	return nil
}
