package wechat

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"

	"webook/webook/internal/pkg/logger"
	"webook/webook/payment/domain"
	"webook/webook/payment/events"
)

// 下单
func TestNativeService_Prepay(t *testing.T) {
	appid := os.Getenv("WEPAY_APP_ID")
	mchID := os.Getenv("WEPAY_MCH_ID")
	mchKey := os.Getenv("WEPAY_MCH_KEY")
	mchSerialNumber := os.Getenv("WEPAY_MCH_SERIAL_NUM")
	// 使用 utils 提供的函数从本地文件中加载商户私钥，商户私钥会用来生成请求的签名
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath(
		"./config/cert/apiclient_key.pem",
	)
	require.NoError(t, err)
	ctx := context.Background()
	// 使用商户私钥等初始化 client
	client, err := core.NewClient(
		ctx,
		option.WithWechatPayAutoAuthCipher(mchID, mchSerialNumber, mchPrivateKey, mchKey),
	)
	require.NoError(t, err)
	nativeSvc := &native.NativeApiService{
		Client: client,
	}
	svc := NewNativePaymentService(nativeSvc, nil, events.NopProducer{},
		logger.NewNoOpLogger(), appid, mchID)
	codeUrl, err := svc.Prepay(ctx, domain.Payment{
		Amt: domain.Amount{
			Currency: "CNY",
			Total:    1,
		},
		BizTradeNO:  "live_demo_1234",
		Description: "面试官AI",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, codeUrl)
	t.Log(codeUrl)
}
