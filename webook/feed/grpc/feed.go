package grpc

import (
	"context"
	"encoding/json"
	"time"

	"google.golang.org/grpc"

	feedv1 "webook/webook/api/proto/gen/feed/v1"
	"webook/webook/feed/domain"
	"webook/webook/feed/service"
)

type FeedEventGrpcSvc struct {
	feedv1.UnimplementedFeedSvcServer
	svc service.FeedService
}

func NewFeedEventGrpcSvc(svc service.FeedService) *FeedEventGrpcSvc {
	return &FeedEventGrpcSvc{
		svc: svc,
	}
}

func (f *FeedEventGrpcSvc) Register(server grpc.ServiceRegistrar) {
	feedv1.RegisterFeedSvcServer(server, f)
}

func (f *FeedEventGrpcSvc) CreateFeedEvent(ctx context.Context, request *feedv1.CreateFeedEventRequest) (*feedv1.CreateFeedEventResponse, error) {
	err := f.svc.CreateFeedEvent(ctx, f.convertToDomain(request.GetFeedEvent()))
	return &feedv1.CreateFeedEventResponse{}, err
}

func (f *FeedEventGrpcSvc) FindFeedEvents(ctx context.Context, request *feedv1.FindFeedEventsRequest) (*feedv1.FindFeedEventsResponse, error) {
	eventList, err := f.svc.GetFeedEventList(ctx, request.GetUid(), request.Timestamp, request.Limit)
	if err != nil {
		return &feedv1.FindFeedEventsResponse{}, err
	}
	res := make([]*feedv1.FeedEvent, 0, len(eventList))
	for _, event := range eventList {
		res = append(res, f.convertToView(event))
	}
	return &feedv1.FindFeedEventsResponse{
		FeedEvents: res,
	}, nil
}

func (f *FeedEventGrpcSvc) convertToDomain(event *feedv1.FeedEvent) domain.FeedEvent {
	ext := map[string]string{}
	_ = json.Unmarshal([]byte(event.Content), &ext)
	return domain.FeedEvent{
		ID:    event.Id,
		Ctime: time.Unix(event.Ctime, 0),
		Type:  event.GetType(),
		Ext:   ext,
	}
}

func (f *FeedEventGrpcSvc) convertToView(event domain.FeedEvent) *feedv1.FeedEvent {
	val, _ := json.Marshal(event.Ext)
	return &feedv1.FeedEvent{
		Id:      event.ID,
		Type:    event.Type,
		Ctime:   event.Ctime.Unix(),
		Content: string(val),
	}
}
