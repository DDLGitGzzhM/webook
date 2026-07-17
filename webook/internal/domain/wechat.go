package domain

type WeChatInfo struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenId       string `json:"openid"`

	Scope   string `json:"scope"`
	UnionId string `json:"unionid"`
}
