package wechat

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"

	"webook/webook/internal/pkg/logger"
	"webook/webook/payment/domain"
	"webook/webook/payment/repository"
)

var errUnknownTransactionState = errors.New("未知的微信事务状态")

type NativePaymentService struct {
	svc       *native.NativeApiService
	appID     string
	mchID     string
	notifyURL string
	repo      repository.PaymentRepository
	l         logger.Logger

	// 在微信 native 里面，分别是
	// SUCCESS：支付成功
	// REFUND：转入退款
	// NOTPAY：未支付
	// CLOSED：已关闭
	// REVOKED：已撤销（付款码支付）
	// USERPAYING：用户支付中（付款码支付）
	// PAYERROR：支付失败(其他原因，如银行返回失败)
	nativeCBTypeToStatus map[string]domain.PaymentStatus
}

func NewNativePaymentService(
	svc *native.NativeApiService,
	repo repository.PaymentRepository,
	l logger.Logger,
	appid, mchid string,
) *NativePaymentService {
	return &NativePaymentService{
		l:     l,
		repo:  repo,
		svc:   svc,
		appID: appid,
		mchID: mchid,
		// 一般来说，这个都是固定的，基本不会变的
		notifyURL: "http://wechat.meoying.com/pay/callback",
		nativeCBTypeToStatus: map[string]domain.PaymentStatus{
			"SUCCESS":  domain.PaymentStatusSuccess,
			"PAYERROR": domain.PaymentStatusFailed,
			"NOTPAY":   domain.PaymentStatusInit,
			"CLOSED":   domain.PaymentStatusFailed,
			"REVOKED":  domain.PaymentStatusFailed,
			"REFUND":   domain.PaymentStatusRefund,
			// 其它状态你都可以加
		},
	}
}

// Prepay 为了拿到扫码支付的二维码
func (n *NativePaymentService) Prepay(ctx context.Context, pmt domain.Payment) (string, error) {
	// 唯一索引冲突
	// 业务方唤起了支付，但是没付，下一次再过来，应该换 BizTradeNO
	err := n.repo.AddPayment(ctx, pmt)
	if err != nil {
		return "", err
	}
	resp, result, err := n.svc.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(n.appID),
		Mchid:       core.String(n.mchID),
		Description: core.String(pmt.Description),
		// 这个地方是有讲究的
		// 选择1：业务方直接给我，我透传，我啥也不干
		// 选择2：业务方给我它的业务标识，我自己生成一个 - 担忧出现重复
		// 注意，不管你是选择 1 还是选择 2，业务方都一定要传给你（webook payment）一个唯一标识
		// Biz + BizTradeNo 唯一， biz + biz_id
		OutTradeNo: core.String(pmt.BizTradeNO),
		NotifyUrl:  core.String(n.notifyURL),
		// 设置三十分钟有效
		TimeExpire: core.Time(time.Now().Add(time.Minute * 30)),
		Amount: &native.Amount{
			Total:    core.Int64(pmt.Amt.Total),
			Currency: core.String(pmt.Amt.Currency),
		},
	})
	n.l.Debug("微信prepay响应",
		logger.Field{Key: "result", Value: result},
		logger.Field{Key: "resp", Value: resp})
	if err != nil {
		return "", err
	}
	return *resp.CodeUrl, nil
}

func (n *NativePaymentService) SyncWechatInfo(ctx context.Context, bizTradeNO string) error {
	txn, _, err := n.svc.QueryOrderByOutTradeNo(ctx, native.QueryOrderByOutTradeNoRequest{
		OutTradeNo: core.String(bizTradeNO),
		Mchid:      core.String(n.mchID),
	})
	if err != nil {
		return err
	}
	return n.updateByTxn(ctx, txn)
}

func (n *NativePaymentService) FindExpiredPayment(
	ctx context.Context, offset, limit int, t time.Time,
) ([]domain.Payment, error) {
	return n.repo.FindExpiredPayment(ctx, offset, limit, t)
}

func (n *NativePaymentService) HandleCallback(ctx context.Context, txn *payments.Transaction) error {
	return n.updateByTxn(ctx, txn)
}

func (n *NativePaymentService) updateByTxn(ctx context.Context, txn *payments.Transaction) error {
	status, ok := n.nativeCBTypeToStatus[*txn.TradeState]
	if !ok {
		return fmt.Errorf("%w, %s", errUnknownTransactionState, *txn.TradeState)
	}
	return n.repo.UpdatePayment(ctx, domain.Payment{
		BizTradeNO: *txn.OutTradeNo,
		TxnID:      *txn.TransactionId,
		Status:     status,
	})
}
