package grpc

import (
	"context"

	"google.golang.org/grpc"

	searchv1 "webook/webook/api/proto/gen/search/v1"
	"webook/webook/search/domain"
	"webook/webook/search/service"
)

// SyncServiceServer 同步 gRPC 服务。
type SyncServiceServer struct {
	searchv1.UnimplementedSyncServiceServer
	syncSvc service.SyncService
}

// NewSyncServiceServer 创建同步 gRPC 服务。
func NewSyncServiceServer(syncSvc service.SyncService) *SyncServiceServer {
	return &SyncServiceServer{syncSvc: syncSvc}
}

// Register 注册到 gRPC Server。
func (s *SyncServiceServer) Register(server grpc.ServiceRegistrar) {
	searchv1.RegisterSyncServiceServer(server, s)
}

// InputUser 业务专属接口，你可以高度定制化
func (s *SyncServiceServer) InputUser(
	ctx context.Context,
	request *searchv1.InputUserRequest,
) (*searchv1.InputUserResponse, error) {
	err := s.syncSvc.InputUser(ctx, s.toDomainUser(request.GetUser()))
	return &searchv1.InputUserResponse{}, err
}

// InputArticle 同步文章到搜索引擎。
func (s *SyncServiceServer) InputArticle(
	ctx context.Context,
	request *searchv1.InputArticleRequest,
) (*searchv1.InputArticleResponse, error) {
	err := s.syncSvc.InputArticle(ctx, s.toDomainArticle(request.GetArticle()))
	return &searchv1.InputArticleResponse{}, err
}

// InputAny 向任意索引写入文档。
func (s *SyncServiceServer) InputAny(
	ctx context.Context,
	req *searchv1.InputAnyRequest,
) (*searchv1.InputAnyResponse, error) {
	err := s.syncSvc.InputAny(ctx, req.IndexName, req.DocId, req.Data)
	return &searchv1.InputAnyResponse{}, err
}

func (s *SyncServiceServer) toDomainUser(vuser *searchv1.User) domain.User {
	return domain.User{
		Id:       vuser.Id,
		Email:    vuser.Email,
		Nickname: vuser.Nickname,
		Phone:    vuser.Phone,
	}
}

func (s *SyncServiceServer) toDomainArticle(art *searchv1.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Status:  art.Status,
		Content: art.Content,
		Tags:    art.Tags,
	}
}
