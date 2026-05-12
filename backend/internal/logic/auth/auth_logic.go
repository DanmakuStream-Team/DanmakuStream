package authlogic

import (
	"context"
	"errors"
	"strconv"
	"strings"
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

type LoginReq struct {
	Nickname string `json:"nickname"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterReq struct {
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

type LoginResp struct {
	Token    string          `json:"token"`
	UserInfo *model.UserInfo `json:"userInfo"`
}

func (l *LoginLogic) Login(req *LoginReq) (*LoginResp, error) {
	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" {
		nickname = strings.TrimSpace(req.Username)
	}
	if nickname == "" {
		return nil, errors.New("昵称或密码错误")
	}

	var user model.User
	if err := l.svcCtx.DB.Where("nickname = ?", nickname).First(&user).Error; err != nil {
		return nil, errors.New("昵称或密码错误")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("昵称或密码错误")
	}

	token, err := generateToken(l.svcCtx, &user)
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

func (l *RegisterLogic) Register(req *RegisterReq) (*LoginResp, error) {
	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" || req.Password == "" {
		return nil, errors.New("昵称和密码不能为空")
	}

	var count int64
	l.svcCtx.DB.Model(&model.User{}).Where("nickname = ?", nickname).Count(&count)
	if count > 0 {
		return nil, errors.New("昵称已存在")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := model.User{
		Username: "pending-" + strconv.FormatInt(time.Now().UnixNano(), 10),
		Password: string(hashedPassword),
		Nickname: nickname,
		Role:     "user",
	}
	if err := l.svcCtx.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	user.Username = strconv.FormatUint(uint64(user.ID), 10)
	if err := l.svcCtx.DB.Model(&user).Update("username", user.Username).Error; err != nil {
		return nil, err
	}

	token, err := generateToken(l.svcCtx, &user)
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

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{ctx: ctx, svcCtx: svcCtx}
}

func generateToken(svcCtx *svc.ServiceContext, user *model.User) (string, error) {
	expire := svcCtx.Config.Auth.AccessExpire
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
	return token.SignedString([]byte(svcCtx.Config.Auth.AccessSecret))
}
