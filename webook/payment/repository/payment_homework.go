package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"webook/webook/payment/domain"
	"webook/webook/payment/repository/dao"
)

// PaymentGORMRepository 支持本地消息表事务的支付仓储（第十七周作业）
type PaymentGORMRepository struct {
	dao dao.PaymentDAO
	db  *gorm.DB
}

func (p *PaymentGORMRepository) GetPayment(ctx context.Context, bizTradeNO string) (domain.Payment, error) {
	r, err := p.dao.GetPayment(ctx, bizTradeNO)
	return p.toDomain(r), err
}

func (p *PaymentGORMRepository) FindExpiredPayment(ctx context.Context, offset int, limit int, t time.Time) ([]domain.Payment, error) {
	pmts, err := p.dao.FindExpiredPayment(ctx, offset, limit, t)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Payment, 0, len(pmts))
	for _, pmt := range pmts {
		res = append(res, p.toDomain(pmt))
	}
	return res, nil
}

func (p *PaymentGORMRepository) AddPayment(ctx context.Context, pmt domain.Payment) error {
	return p.dao.Insert(ctx, p.toEntity(pmt))
}

func (p *PaymentGORMRepository) toDomain(pmt dao.Payment) domain.Payment {
	return domain.Payment{
		Amt: domain.Amount{
			Currency: pmt.Currency,
			Total:    pmt.Amt,
		},
		BizTradeNO:  pmt.BizTradeNO,
		Description: pmt.Description,
		Status:      domain.PaymentStatus(pmt.Status),
		TxnID:       pmt.TxnID.String,
	}
}

func (p *PaymentGORMRepository) toEntity(pmt domain.Payment) dao.Payment {
	return dao.Payment{
		Amt:         pmt.Amt.Total,
		Currency:    pmt.Amt.Currency,
		BizTradeNO:  pmt.BizTradeNO,
		Description: pmt.Description,
		Status:      domain.PaymentStatusInit,
	}
}

func (p *PaymentGORMRepository) UpdatePayment(ctx context.Context, pmt domain.Payment) error {
	return p.dao.UpdateTxnIDAndStatus(ctx, pmt.BizTradeNO, pmt.TxnID, pmt.Status)
}

// Transaction 目前只会控制两个 dao，所以可以在 cb 里面直接传入两个 DAO
func (p *PaymentGORMRepository) Transaction(ctx context.Context, cb func(pmt *PaymentGORMRepository, msg *LocalMsgGORMRepository) error) error {
	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return cb(NewPaymentGORMRepository(tx), NewLocalMsgGORMRepository(tx))
	})
}

// NewPaymentGORMRepository 创建支持事务的支付仓储
func NewPaymentGORMRepository(db *gorm.DB) *PaymentGORMRepository {
	return &PaymentGORMRepository{
		dao: dao.NewPaymentGORMDAO(db),
		db:  db,
	}
}
