package grpc

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Server struct {
	UnimplementedUserServiceServer
	name string
}

func (s *Server) GetById(
	ctx context.Context,
	req *GetByIdReq) (*GetByIdResp, error) {
	ddl, ok := ctx.Deadline()
	if ok {
		rest := ddl.Sub(time.Now())
		fmt.Println(rest.String())
	}
	log.Println("命中服务器", s.name)
	return &GetByIdResp{
		User: &User{
			Id:   req.Id,
			Name: s.name,
		},
	}, nil
}
