package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"webook/webook/payment/domain"
	"webook/webook/payment/integration/startup"
	"webook/webook/payment/repository/dao"
	"webook/webook/payment/service/wechat"
)

type WechatNativeServiceTestSuite struct {
	suite.Suite
	svc *wechat.NativePaymentService
	db  *gorm.DB
}

func TestWechatNativeService(t *testing.T) {
	suite.Run(t, new(WechatNativeServiceTestSuite))
}

func (s *WechatNativeServiceTestSuite) SetupSuite() {
	s.svc = startup.InitWechatNativeService()
	s.db = startup.InitTestDB()
}

func (s *WechatNativeServiceTestSuite) TearDownTest() {
	s.db.Exec("TRUNCATE TABLE `payments`")
}

// 记得配置各个环境变量
func (s *WechatNativeServiceTestSuite) TestPrepay() {
	bizNo1 := "integration-1235-p"
	testCases := []struct {
		name    string
		pmt     domain.Payment
		after   func(t *testing.T)
		wantErr error
	}{
		{
			name: "获得了code_url",
			pmt: domain.Payment{
				Amt: domain.Amount{
					Total:    1,
					Currency: "CNY",
				},
				BizTradeNO:  bizNo1,
				Description: "我在这边买了一个产品",
			},
			after: func(t *testing.T) {
				var pmt dao.Payment
				err := s.db.Where("biz_trade_no = ?", bizNo1).First(&pmt).Error
				require.NoError(t, err)
				assert.True(t, pmt.Id > 0)
				pmt.Id = 0
				assert.True(t, pmt.Ctime > 0)
				pmt.Ctime = 0
				assert.True(t, pmt.Utime > 0)
				pmt.Utime = 0
				assert.Equal(t, dao.Payment{
					Amt:         1,
					Currency:    "CNY",
					BizTradeNO:  bizNo1,
					Description: "我在这边买了一个产品",
					Status:      domain.PaymentStatusInit,
				}, pmt)
			},
		},
	}
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()
			url, err := s.svc.Prepay(ctx, tc.pmt)
			assert.Equal(t, tc.wantErr, err)
			if tc.wantErr == nil {
				assert.NotEmpty(t, url)
				t.Log(url)
			}
			tc.after(t)
		})
	}
}
