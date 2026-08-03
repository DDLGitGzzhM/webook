package grpc

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Server struct {
	UnimplementedUserServiceServer
	Name string
}

func (s *Server) GetById(
	ctx context.Context,
	req *GetByIdReq) (*GetByIdResp, error) {
	ddl, ok := ctx.Deadline()
	if ok {
		rest := ddl.Sub(time.Now())
		fmt.Println(rest.String())
	}
	log.Println("命中服务器", s.Name)
	return &GetByIdResp{
		User: &User{
			Id:   req.Id,
			Name: "abcd, from " + s.Name,
		},
	}, nil
}
