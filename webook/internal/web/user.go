package web

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"webook/webook/internal/domain"
	"webook/webook/internal/service"
)

// UserHandler 将所有跟用户有关的路由定义在 Handler 上
type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		svc: svc,
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.Login)
	ug.POST("/edit", u.Edit)
	ug.GET("/profile", u.Profile)
}
func (u *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		ConfirmPassword string `json:"ConfirmPassword"`
		Password        string `json:"password"`
	}
	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	if req.ConfirmPassword != req.Password {
		ctx.String(http.StatusOK, "密码不一致")
		return
	}
	err := u.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		PassWord: req.Password,
	})
	if err != nil {
		if errors.Is(err, service.ErrUserDuplicateEmail) {
			ctx.String(http.StatusOK, "邮箱冲突")
			return
		}
		ctx.String(http.StatusOK, "注册失败")
		return
	}
	ctx.String(http.StatusOK, "注册成功")
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "登录失败 : %s", err.Error())
		return
	}

	userClaims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
		UserId: user.Id,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims)
	tokenStr, err := token.SignedString([]byte("secret"))
	if err != nil {
		ctx.String(http.StatusOK, "登录失败 : %s", err.Error())
		return
	}

	ctx.Header("x-jwt-token", tokenStr)
	ctx.String(http.StatusOK, "登录成功")
}

func (u *UserHandler) Edit(ctx *gin.Context) {

}

func (u *UserHandler) Profile(ctx *gin.Context) {
	uid := ctx.GetInt64("userId")
	ctx.String(http.StatusOK, strconv.FormatInt(uid, 10))
}

type UserClaims struct {
	jwt.RegisteredClaims
	UserId int64
}
