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

func (s *server) GetGroups(
	ctx context.Context,
	req *desc.GetGroupsRequest,
) (*desc.GetGroupsResponse, error) {
	h, err := newGetGroupsHandler(ctx, s.clients, req)
	if err != nil {
		return nil, err
	}
	err = h.handle()
	return h.response(), err
}

func (h *getGroupsHandler) handle() error {
	var (
		groups []string
		err    error
	)

	if h.socialNetwork == nil {
		var (
			telegramGroups []string
			vkGroups       []string
		)
		telegramGroups, err = h.getTelegramGroups()
		if err != nil {
			return err
		}
		vkGroups, err = h.getVkGroups()
		if err != nil {
			return err
		}

		groups = append(groups, telegramGroups...)
		groups = append(groups, vkGroups...)
	} else {
		switch *h.socialNetwork {
		case desc.SocialNetwork_TELEGRAM.String():
			groups, err = h.getTelegramGroups()
		case desc.SocialNetwork_VK.String():
			groups, err = h.getVkGroups()
		}
		if err != nil {
			return err
		}
	}

	if int64(len(groups)) < h.offset {
		h.groups = make([]string, 0)
		return nil
	}
	h.groups = groups[h.offset:]
	if int64(len(h.groups)) > h.limit {
		h.groups = h.groups[:h.limit]
	}
	return nil
}

func (h *getGroupsHandler) getTelegramGroups() ([]string, error) {
	groups, err := h.clients.TelegramClient().GetGroups(h.ctx, &telegram.GetGroupsRequest{})
	if err != nil {
		return nil, err
	}
	if groups == nil {
		return nil, errors.
			NewNetworkError(codes.Internal, "telegram: got null groups").
			ToGRPCError()
	}

	return groups.Groups, nil
}

func (h *getGroupsHandler) getVkGroups() ([]string, error) {
	groups, err := h.clients.VkClient().GetGroups(h.ctx, &vk.GetGroupsRequest{})
	if err != nil {
		return nil, err
	}
	if groups == nil {
		return nil, errors.
			NewNetworkError(codes.Internal, "vk: got null groups").
			ToGRPCError()
	}

	return groups.Groups, nil
}

func (h *getGroupsHandler) response() *desc.GetGroupsResponse {
	return &desc.GetGroupsResponse{
		Groups: h.groups,
		Limit:  h.limit,
		Offset: h.offset,
	}
}

type getGroupsHandler struct {
	ctx     context.Context
	clients clients.Clients

	limit         int64
	offset        int64
	socialNetwork *string

	groups []string
}

func newGetGroupsHandler(
	ctx context.Context,
	clients clients.Clients,
	req *desc.GetGroupsRequest,
) (*getGroupsHandler, error) {
	h := &getGroupsHandler{
		ctx:     ctx,
		clients: clients,
	}
	return h.adapt(req), h.validate()
}

func (h *getGroupsHandler) adapt(req *desc.GetGroupsRequest) *getGroupsHandler {
	h.limit = req.GetLimit()
	h.offset = req.GetOffset()

	if req.SocialNetwork != nil {
		socialNetwork := req.GetSocialNetwork().String()
		h.socialNetwork = &socialNetwork
	}
	return h
}

func (h *getGroupsHandler) validate() error {
	if h.limit <= 0 {
		return errors.
			NewNetworkError(codes.InvalidArgument, "limit must be specified").
			ToGRPCError()
	}

	if h.offset < 0 {
		return errors.
			NewNetworkError(codes.InvalidArgument, "offset must be greater or equal 0").
			ToGRPCError()
	}
	return nil
}
