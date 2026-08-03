package grpc

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestServer(t *testing.T) {
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(first, second))
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
	conn, err := grpc.Dial("localhost:8090",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(clientFirst, clientSecond))
	assert.NoError(t, err)
	client := NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	resp, err := client.GetById(ctx, &GetByIdReq{
		Id: 123,
	})
	assert.NoError(t, err)
	t.Log(resp.User)
}

var first grpc.UnaryServerInterceptor = func(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp any, err error) {
	log.Println("这是第一个前")
	resp, err = handler(ctx, req)
	log.Println("这是第一个后")
	return
}

var second grpc.UnaryServerInterceptor = func(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp any, err error) {
	log.Println("这是第二个前")
	resp, err = handler(ctx, req)
	log.Println("这是第二个后")
	return
}

var clientFirst grpc.UnaryClientInterceptor = func(ctx context.Context,
	method string, req, reply any,
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	log.Println("客户端第一个前")
	err := invoker(ctx, method, req, reply, cc, opts...)
	log.Println("客户端第一个后")
	return err
}

var clientSecond grpc.UnaryClientInterceptor = func(ctx context.Context,
	method string, req, reply any,
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {
	log.Println("客户端第二个前")
	err := invoker(ctx, method, req, reply, cc, opts...)
	log.Println("客户端第二个后")
	return err
}
