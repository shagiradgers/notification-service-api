package clients

import (
	"context"

	"google.golang.org/grpc"
	"notification-service-api/internal/config"
)

type Clients interface {
	TelegramClient() *TelegramClient
	VkClient() *VkClient
}

type clients struct {
	telegramClient *TelegramClient
	vkClient       *VkClient
}

func (c *clients) TelegramClient() *TelegramClient {
	return c.telegramClient
}

func (c *clients) VkClient() *VkClient {
	return c.vkClient
}

func NewClients(ctx context.Context, config config.Config) Clients {
	c := &clients{
		telegramClient: nil,
		vkClient:       nil,
	}

	c.vkClient = NewVK(mustGetClientConn(ctx, config.MustGetVkEndpoint()))
	c.telegramClient = NewTelegram(mustGetClientConn(ctx, config.MustGetTelegramEndpoint()))

	return c
}

func getClientConn(ctx context.Context, endpoint string) (grpc.ClientConnInterface, error) {
	return grpc.DialContext(ctx, endpoint)
}

func mustGetClientConn(ctx context.Context, endpoint string) grpc.ClientConnInterface {
	cc, err := grpc.DialContext(ctx, endpoint, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return cc
}
