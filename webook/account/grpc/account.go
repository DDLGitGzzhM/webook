package grpc

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"google.golang.org/grpc"
	"webook/webook/account/domain"
	"webook/webook/account/service"
	accountv1 "webook/webook/api/proto/gen/account/v1"
)

type AccountServiceServer struct {
	accountv1.UnimplementedAccountServiceServer
	svc service.AccountService
}

func NewAccountServiceServer(svc service.AccountService) *AccountServiceServer {
	return &AccountServiceServer{svc: svc}
}

func (a *AccountServiceServer) Credit(ctx context.Context,
	req *accountv1.CreditRequest) (*accountv1.CreditResponse, error) {
	err := a.svc.Credit(ctx, a.toDomain(req))
	return &accountv1.CreditResponse{}, err
}

func (a *AccountServiceServer) toDomain(c *accountv1.CreditRequest) domain.Credit {
	return domain.Credit{
		Biz:   c.Biz,
		BizId: c.BizId,
		Items: slice.Map(c.Items, func(idx int, src *accountv1.CreditItem) domain.CreditItem {
			return a.itemToDomain(src)
		}),
	}
}

func (a *AccountServiceServer) itemToDomain(c *accountv1.CreditItem) domain.CreditItem {
	return domain.CreditItem{
		Account: c.Account,
		Amt:     c.Amt,
		Uid:     c.Uid,
		// 两者取值都是一样的，我偷个懒，直接转
		AccountType: domain.AccountType(c.AccountType),
		Currency:    c.Currency,
	}
}

func (a *AccountServiceServer) Register(server *grpc.Server) {
	accountv1.RegisterAccountServiceServer(server, a)
}
