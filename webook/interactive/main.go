package main

import (
	"log"
	"net"

	grpc2 "google.golang.org/grpc"

	intrv1 "webook/webook/api/proto/gen/intr/v1"
	"webook/webook/interactive/grpc"
)

func main() {
	server := grpc2.NewServer()
	// 这里暂时随便搞一下
	intrSvc := &grpc.InteractiveServiceServer{}
	intrv1.RegisterInteractiveServiceServer(server, intrSvc)
	// 监听 8090 端口，你可以随便写
	l, err := net.Listen("tcp", ":8090")
	if err != nil {
		panic(err)
	}
	// 这边会阻塞，类似于 gin.Run
	err = server.Serve(l)
	log.Println(err)
}
