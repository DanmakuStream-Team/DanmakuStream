package authlogic

import (
	"context"
	"testing"
	"time"

	"danmakustream/user-service/internal/config"
	"danmakustream/user-service/internal/middleware"
	model "danmakustream/user-service/internal/model/mysql"
	"danmakustream/user-service/internal/svc"

	"github.com/golang-jwt/jwt/v5"
)

func TestRegisterRejectsInvalidCredentialsWithoutDatabase(t *testing.T) {
	t.Parallel()
	logic := NewRegisterLogic(context.Background(), &svc.ServiceContext{})

	tests := []struct {
		name string
		req  RegisterReq
	}{
		{name: "blank nickname", req: RegisterReq{Nickname: "   ", Password: "secret"}},
		{name: "missing password", req: RegisterReq{Nickname: "alice"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := logic.Register(&test.req); err == nil || err.Error() != "昵称和密码不能为空" {
				t.Fatalf("Register(%+v) error = %v, want 昵称和密码不能为空", test.req, err)
			}
		})
	}
}

func TestLoginRejectsMissingIdentityWithoutDatabase(t *testing.T) {
	t.Parallel()
	logic := NewLoginLogic(context.Background(), &svc.ServiceContext{})

	for _, req := range []LoginReq{
		{},
		{Nickname: "  ", Username: "\t", Password: "secret"},
	} {
		if _, err := logic.Login(&req); err == nil || err.Error() != "昵称或密码错误" {
			t.Fatalf("Login(%+v) error = %v, want 昵称或密码错误", req, err)
		}
	}
}

func TestGenerateTokenCarriesUserIdentityAndExpiry(t *testing.T) {
	t.Parallel()
	const secret = "unit-test-secret"
	testConfig := config.Config{}
	testConfig.Auth.AccessSecret = secret
	testConfig.Auth.AccessExpire = 120
	svcCtx := &svc.ServiceContext{Config: testConfig}
	user := &model.User{Username: "42", Nickname: "alice", Role: "creator"}
	user.ID = 42

	issuedBefore := time.Now()
	tokenString, err := generateToken(svcCtx, user)
	if err != nil {
		t.Fatal(err)
	}

	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		t.Fatalf("generated token invalid: token=%v err=%v", token.Valid, err)
	}
	if claims.UserID != user.ID || claims.Username != user.Username || claims.Role != user.Role {
		t.Fatalf("claims = %+v, want user id=%d username=%q role=%q", claims, user.ID, user.Username, user.Role)
	}
	if claims.IssuedAt == nil || claims.ExpiresAt == nil {
		t.Fatalf("token timestamps missing: issued=%v expires=%v", claims.IssuedAt, claims.ExpiresAt)
	}
	if claims.IssuedAt.Time.Before(issuedBefore.Add(-time.Second)) {
		t.Fatalf("issuedAt %v is earlier than test start %v", claims.IssuedAt.Time, issuedBefore)
	}
	wantExpiry := issuedBefore.Add(120 * time.Second)
	if delta := claims.ExpiresAt.Time.Sub(wantExpiry); delta < -time.Second || delta > time.Second {
		t.Fatalf("expiresAt = %v, want about %v", claims.ExpiresAt.Time, wantExpiry)
	}

	wrongSecretClaims := &middleware.Claims{}
	if _, err := jwt.ParseWithClaims(tokenString, wrongSecretClaims, func(token *jwt.Token) (any, error) {
		return []byte("wrong-secret"), nil
	}); err == nil {
		t.Fatal("token unexpectedly validates with wrong secret")
	}
}
