package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AlwaysFailedServer struct {
	UnimplementedUserServiceServer
	Name string
}

var _ UserServiceServer = &AlwaysFailedServer{}

func (s *AlwaysFailedServer) GetById(
	ctx context.Context, request *GetByIdReq) (*GetByIdResp, error) {
	fmt.Println("进来了 failed 服务端")
	return &GetByIdResp{
		User: &User{
			Id:   123,
			Name: "来自永远失败的服务端节点 " + s.Name,
		},
	}, status.Errorf(codes.Unavailable, "模拟服务端异常")
}
