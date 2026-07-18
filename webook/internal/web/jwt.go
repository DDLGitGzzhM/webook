package web

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JwtHandler struct {
	atKey []byte // access_token_key
	rtKey []byte // refresh_token_key
}

func NewJWTHandler() *JwtHandler {
	return &JwtHandler{
		atKey: []byte("at_secret"),
		rtKey: []byte("rt_secret"),
	}
}

func (j *JwtHandler) setLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := j.setJWTToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	err = j.setRefreshToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	return nil
}
func (j *JwtHandler) setJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	userClaims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		UserId:     uid,
		Ssid:       ssid,
		UserAgents: ctx.Request.UserAgent(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims)
	tokenStr, err := token.SignedString(j.atKey)
	if err != nil {
		ctx.String(http.StatusOK, "登录失败 : %s", err.Error())
		return err
	}

	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (j *JwtHandler) setRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	userClaims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
		Uid:  uid,
		Ssid: ssid,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims)
	tokenStr, err := token.SignedString(j.rtKey)
	if err != nil {
		ctx.String(http.StatusOK, "登录失败 : %s", err.Error())
		return err
	}

	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

func ExtractToken(ctx *gin.Context) string {
	tokenHeader := ctx.GetHeader("Authorization")
	segs := strings.Split(tokenHeader, " ")
	if len(segs) != 2 || segs[0] != "Bearer" {
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}
	return segs[1]
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid  int64
	Ssid string
}

type UserClaims struct {
	jwt.RegisteredClaims
	UserId     int64
	UserAgents string
	Ssid       string
}
