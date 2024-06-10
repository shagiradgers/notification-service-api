package clients

import (
	"google.golang.org/grpc"
	telegram "notification-service-api/pb/tg"
)

type TelegramClient struct {
	telegram.TelegramNotificationServiceClient
}

func NewTelegram(cc grpc.ClientConnInterface) *TelegramClient {
	return &TelegramClient{telegram.NewTelegramNotificationServiceClient(cc)}
}
