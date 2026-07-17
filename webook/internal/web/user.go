package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"webook/webook/internal/domain"
	"webook/webook/internal/service"
)

// UserHandler 将所有跟用户有关的路由定义在 Handler 上
type UserHandler struct {
	svc     service.IUserService
	codeSvc service.ICodeService
	jwtHandler
}

func NewUserHandler(svc service.IUserService, codeSvc service.ICodeService) *UserHandler {
	return &UserHandler{
		svc:     svc,
		codeSvc: codeSvc,
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.Login)
	ug.POST("/edit", u.Edit)
	ug.GET("/profile", u.Profile)
	ug.POST("/login_sms/code/send", u.SendLoginSMSCode)
	ug.POST("/login_sms", u.LoginSMS)

}

func (u *UserHandler) LoginSMS(ctx *gin.Context) {
	const biz = "login"
	type Req struct {
		Code  string `json:"code"`
		Phone string `json:"phone"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 400,
			Msg:  "请求参数错误",
		})
		return
	}
	verify, err := u.codeSvc.Verify(ctx, biz, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  fmt.Sprintf("验证码校验失败: %s", err.Error()),
		})
		return
	}
	if !verify {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  fmt.Sprintf("错误的验证码 : %s", req.Code),
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "成功登陆",
	})
}

func (u *UserHandler) SendLoginSMSCode(ctx *gin.Context) {
	const biz = "login"
	type SendLoginSMSCodeReq struct {
		Phone string `json:"phone"`
	}
	var req SendLoginSMSCodeReq
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 400,
			Msg:  "请求参数错误",
		})
		return
	}
	err := u.codeSvc.Send(ctx, biz, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  fmt.Sprintf("发送验证码失败: %s", err.Error()),
		})
		return
	}
	/*
	 这里的 id 从哪里来？ 如果FindByPhone 我这个手机号会不会是一个新用户
	*/
	user, err := u.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  fmt.Sprintf("系统错误: %s", err.Error()),
		})
		return
	}
	if err = u.setJWTToken(ctx, user.Id); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  fmt.Sprintf("设置token失败: %s", err.Error()),
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "发送成功",
	})
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

	if err = u.setJWTToken(ctx, user.Id); err != nil {
		return
	}
	ctx.String(http.StatusOK, "登录成功")
}

func (u *UserHandler) Edit(ctx *gin.Context) {

}

func (u *UserHandler) Profile(ctx *gin.Context) {
	uid := ctx.GetInt64("userId")
	userInfo, err := u.svc.Profile(ctx, uid)
	if err != nil {
		ctx.String(http.StatusOK, "获取用户信息失败 : %s", err.Error())
		return
	}
	userInfoStr, err := json.Marshal(userInfo)
	if err != nil {
		ctx.String(http.StatusOK, "获取用户信息失败 : %s", err.Error())
		return
	}
	ctx.String(http.StatusOK, string(userInfoStr))
}
