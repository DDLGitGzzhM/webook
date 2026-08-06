package grpc

import (
	"context"

	"github.com/ecodeclub/ekit/slice"
	"google.golang.org/grpc"

	searchv1 "webook/webook/api/proto/gen/search/v1"
	"webook/webook/search/domain"
	"webook/webook/search/service"
)

// SearchServiceServer 搜索 gRPC 服务。
type SearchServiceServer struct {
	searchv1.UnimplementedSearchServiceServer
	svc service.SearchService
}

// NewSearchService 创建搜索 gRPC 服务。
func NewSearchService(svc service.SearchService) *SearchServiceServer {
	return &SearchServiceServer{svc: svc}
}

// Register 注册到 gRPC Server。
func (s *SearchServiceServer) Register(server grpc.ServiceRegistrar) {
	searchv1.RegisterSearchServiceServer(server, s)
}

// Search 执行聚合搜索。
func (s *SearchServiceServer) Search(
	ctx context.Context,
	request *searchv1.SearchRequest,
) (*searchv1.SearchResponse, error) {
	resp, err := s.svc.Search(ctx, request.Uid, request.Expression)
	if err != nil {
		return nil, err
	}
	return &searchv1.SearchResponse{
		User: &searchv1.UserResult{
			Users: slice.Map(resp.Users, func(idx int, src domain.User) *searchv1.User {
				return &searchv1.User{
					Id:       src.Id,
					Nickname: src.Nickname,
					Email:    src.Email,
					Phone:    src.Phone,
				}
			}),
		},
		Article: &searchv1.ArticleResult{
			Articles: slice.Map(resp.Articles, func(idx int, src domain.Article) *searchv1.Article {
				return &searchv1.Article{
					Id:      src.Id,
					Title:   src.Title,
					Status:  src.Status,
					Content: src.Content,
				}
			}),
		},
	}, nil
}
