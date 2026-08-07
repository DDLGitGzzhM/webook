package grpc

import (
	"context"

	"google.golang.org/grpc"

	followv1 "webook/webook/api/proto/gen/follow/v1"
	"webook/webook/follow/domain"
	"webook/webook/follow/service"
)

// FollowServiceServer 关注关系 gRPC 服务
type FollowServiceServer struct {
	followv1.UnimplementedFollowServiceServer
	svc service.FollowRelationService
}

// NewFollowRelationServiceServer 创建关注关系 gRPC 服务
func NewFollowRelationServiceServer(svc service.FollowRelationService) *FollowServiceServer {
	return &FollowServiceServer{
		svc: svc,
	}
}

// Register 注册到 gRPC Server
func (f *FollowServiceServer) Register(server grpc.ServiceRegistrar) {
	followv1.RegisterFollowServiceServer(server, f)
}

func (f *FollowServiceServer) GetFollowee(ctx context.Context, request *followv1.GetFolloweeRequest) (*followv1.GetFolloweeResponse, error) {
	relationList, err := f.svc.GetFollowee(ctx, request.Follower, request.Offset, request.Limit)
	if err != nil {
		return nil, err
	}
	res := make([]*followv1.FollowRelation, 0, len(relationList))
	for _, relation := range relationList {
		res = append(res, f.convertToView(relation))
	}
	return &followv1.GetFolloweeResponse{
		FollowRelations: res,
	}, nil
}

func (f *FollowServiceServer) FollowInfo(ctx context.Context, request *followv1.FollowInfoRequest) (*followv1.FollowInfoResponse, error) {
	info, err := f.svc.FollowInfo(ctx, request.Follower, request.Followee)
	if err != nil {
		return nil, err
	}
	return &followv1.FollowInfoResponse{
		FollowRelation: f.convertToView(info),
	}, nil
}

func (f *FollowServiceServer) Follow(ctx context.Context, request *followv1.FollowRequest) (*followv1.FollowResponse, error) {
	err := f.svc.Follow(ctx, request.Follower, request.Followee)
	return &followv1.FollowResponse{}, err
}

func (f *FollowServiceServer) CancelFollow(ctx context.Context, request *followv1.CancelFollowRequest) (*followv1.CancelFollowResponse, error) {
	err := f.svc.CancelFollow(ctx, request.Follower, request.Followee)
	return &followv1.CancelFollowResponse{}, err
}

func (f *FollowServiceServer) GetFollower(ctx context.Context, request *followv1.GetFollowerRequest) (*followv1.GetFollowerResponse, error) {
	// 写扩散场景粉丝数通常不大，这里一次性拉取
	relationList, err := f.svc.GetFollower(ctx, request.Followee, 0, 10000)
	if err != nil {
		return nil, err
	}
	res := make([]*followv1.FollowRelation, 0, len(relationList))
	for _, relation := range relationList {
		res = append(res, f.convertToView(relation))
	}
	return &followv1.GetFollowerResponse{
		FollowRelations: res,
	}, nil
}

func (f *FollowServiceServer) GetFollowStatic(ctx context.Context, request *followv1.GetFollowStaticRequest) (*followv1.GetFollowStaticResponse, error) {
	static, err := f.svc.GetFollowStatic(ctx, request.Followee)
	if err != nil {
		return nil, err
	}
	return &followv1.GetFollowStaticResponse{
		FollowStatic: &followv1.FollowStatic{
			Followers: static.Followers,
			Followees: static.Followees,
		},
	}, nil
}

func (f *FollowServiceServer) convertToView(relation domain.FollowRelation) *followv1.FollowRelation {
	return &followv1.FollowRelation{
		Followee: relation.Followee,
		Follower: relation.Follower,
	}
}
