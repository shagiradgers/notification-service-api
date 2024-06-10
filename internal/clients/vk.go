package clients

import (
	"google.golang.org/grpc"
	"notification-service-api/pb/vk"
)

type VkClient struct {
	vk.VkNotificationsApiClient
}

func NewVK(cc grpc.ClientConnInterface) *VkClient {
	return &VkClient{vk.NewVkNotificationsApiClient(cc)}
}
