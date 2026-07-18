package jwt

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// 登录成功返回的 key
	AtKey = "at_secret"
	RtKey = "rt_secret"
)

type RedisJwt struct {
	cmd redis.Cmdable
}

func NewRedisJwt(cmd redis.Cmdable) *RedisJwt {
	return &RedisJwt{
		cmd: cmd,
	}
}

func (r RedisJwt) ExtractToken(ctx *gin.Context) string {
	tokenHeader := ctx.GetHeader("Authorization")
	segs := strings.Split(tokenHeader, " ")
	if len(segs) != 2 || segs[0] != "Bearer" {
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}
	return segs[1]
}

func (r RedisJwt) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := r.SetJWTToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	err = r.SetRefreshToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	return nil
}

func (r RedisJwt) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	userClaims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		UserId:     uid,
		Ssid:       ssid,
		UserAgents: ctx.Request.UserAgent(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims)
	tokenStr, err := token.SignedString([]byte(AtKey))
	if err != nil {
		ctx.String(http.StatusOK, "登录失败 : %s", err.Error())
		return err
	}

	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (r RedisJwt) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	c, _ := ctx.Get("claims")
	claims, ok := c.(*UserClaims)
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return nil
	}
	return r.cmd.Set(ctx, fmt.Sprintf("users:ssid:%s", claims.Ssid), 1, time.Hour*24*7).Err()
}

func (r RedisJwt) CheckSession(ctx *gin.Context, ssid string) error {
	cnt, err := r.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	if err != nil || cnt > 0 {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return err
	}
	return nil
}

func (r RedisJwt) SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	userClaims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
		Uid:  uid,
		Ssid: ssid,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims)
	tokenStr, err := token.SignedString([]byte(RtKey))
	if err != nil {
		ctx.String(http.StatusOK, "登录失败 : %s", err.Error())
		return err
	}

	ctx.Header("x-refresh-token", tokenStr)
	return nil
}
