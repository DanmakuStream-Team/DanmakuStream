package authlogic

import (
	"context"
	"errors"
	"time"

	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{ctx: ctx, svcCtx: svcCtx}
}

type LoginResp struct {
	Token    string          `json:"token"`
	UserInfo *model.UserInfo `json:"userInfo"`
}



func (l *LoginLogic) Login(req *LoginReq) (*LoginResp, error) {
	var user model.User
	if err := l.svcCtx.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	token, err := l.generateToken(&user)
	if err != nil {
		return nil, err
	}

	return &LoginResp{
		Token: token,
		UserInfo: &model.UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Role:     user.Role,
		},
	}, nil
}

func (l *LoginLogic) generateToken(user *model.User) (string, error) {
	expire := l.svcCtx.Config.Auth.AccessExpire
	claims := middleware.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(l.svcCtx.Config.Auth.AccessSecret))
}

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *RegisterLogic) Register(req *RegisterReq) (*LoginResp, error) {
	// Check duplicate username
	var count int64
	l.svcCtx.DB.Model(&model.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := model.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
		Role:     "user",
	}
	if err := l.svcCtx.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	token, err := l.generateToken(&user)
	if err != nil {
		return nil, err
	}

	return &LoginResp{
		Token: token,
		UserInfo: &model.UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Nickname: user.Nickname,
			Role:     user.Role,
		},
	}, nil
}

func (l *RegisterLogic) generateToken(user *model.User) (string, error) {
	expire := l.svcCtx.Config.Auth.AccessExpire
	claims := middleware.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(l.svcCtx.Config.Auth.AccessSecret))
}

// Request types (shared between handler and logic)
type LoginReq = struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type RegisterReq = struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}


type MeLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
}

func NewMeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MeLogic {
    return &MeLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *MeLogic) Me() (*model.UserInfo, error) {
    userID, ok := l.ctx.Value(middleware.CtxKeyUserID).(uint)
    if !ok || userID == 0 {
        return nil, errors.New("未登录")
    }

    var user model.User
    if err := l.svcCtx.DB.First(&user, userID).Error; err != nil {
        return nil, err
    }

    return &model.UserInfo{
        ID:       user.ID,
        Username: user.Username,
        Nickname: user.Nickname,
        Avatar:   user.Avatar,
        Role:     user.Role,
    }, nil
}