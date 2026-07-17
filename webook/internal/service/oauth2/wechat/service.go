package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	uuid "github.com/lithammer/shortuuid/v4"

	"webook/webook/internal/domain"
)

var redirectURL = url.PathEscape("http://127.0.0.1:8080/wechat/callback")

type Service interface {
	AuthURL(ctx context.Context) (string, error)
	VerifyCode(ctx context.Context, code, state string) (domain.WeChatInfo, error)
}

type WeChatService struct {
	appId     string
	appSecret string
	client    *http.Client
}

func NewWeChatService(appId string, appSecret string) *WeChatService {
	return &WeChatService{
		appId:     appId,
		appSecret: appSecret,
		client:    http.DefaultClient,
	}
}

func (w *WeChatService) AuthURL(ctx context.Context) (string, error) {
	const urlPattern = "https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect"
	state := uuid.New()
	return fmt.Sprintf(urlPattern, "appid", redirectURL, state), nil
}

func (w *WeChatService) VerifyCode(ctx context.Context, code, state string) (domain.WeChatInfo, error) {
	const targetPattern = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"
	target := fmt.Sprintf(targetPattern, w.appId, w.appSecret, code)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return domain.WeChatInfo{}, err
	}
	resp, err := w.client.Do(req)
	if err != nil {
		return domain.WeChatInfo{}, err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var res Result
	err = decoder.Decode(&res)
	if err != nil {
		return domain.WeChatInfo{}, err
	}
	// Unmarshal 会再读一遍预计两遍
	// body, err := io.ReadALl
	// err ;= json.Unmarshal(body, &res)
	if res.ErrCode != 0 {
		return domain.WeChatInfo{}, fmt.Errorf("微信返回错误: %s", res.ErrMsg)
	}
	return domain.WeChatInfo{
		OpenId:       res.OpenId,
		AccessToken:  res.AccessToken,
		ExpiresIn:    res.ExpiresIn,
		RefreshToken: res.RefreshToken,
		Scope:        res.Scope,
		UnionId:      res.UnionId,
	}, nil
}

type Result struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`

	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenId       string `json:"openid"`

	Scope   string `json:"scope"`
	UnionId string `json:"unionid"`
}
