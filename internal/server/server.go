package server

import (
	desc "notification-service-api/api"
	"notification-service-api/internal/clients"
)

type server struct {
	clients clients.Clients

	desc.UnimplementedNotificationServiceServer
}

func NewServer(clients clients.Clients) desc.NotificationServiceServer {
	return &server{
		clients: clients,
	}
}
