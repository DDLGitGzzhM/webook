package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestServer(t *testing.T) {
	s := grpc.NewServer()
	// 这个是生成的代码
	RegisterUserServiceServer(s, &Server{})
	l, err := net.Listen("tcp", ":8090")
	assert.NoError(t, err)
	// 启动
	if err = s.Serve(l); err != nil {
		// 启动失败，或者退出了服务器
		t.Log("退出 gRPC 服务", err)
	}
}

func TestClient(t *testing.T) {
	// 早期都是用 WithInsecure 选项，现在已经不用了
	//conn, err := grpc.Dial(":8090", grpc.WithInsecure())
	conn, err := grpc.Dial(":8090",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	client := NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := client.GetById(ctx, &GetByIdReq{
		Id: 123,
	})
	assert.NoError(t, err)
	t.Log(resp.User)
}
