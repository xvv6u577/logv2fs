package helper

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SignedDetails 是项目内部使用的 JWT Claims。
// 在 v5 中通过内嵌 jwt.RegisteredClaims（替代旧版 jwt.StandardClaims）来携带标准字段：
// iss / aud / iat / nbf / exp / sub / jti。
type SignedDetails struct {
	Email string `json:"email"`
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
	Uid   string `json:"uid"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

const (
	defaultIssuer   = "logv2fs"
	defaultAudience = "logv2fs-web"

	accessTokenTTL  = 24 * time.Hour
	refreshTokenTTL = 7 * 24 * time.Hour
)

// getSecretKey 从环境变量读取签名密钥。
// 若未设置，函数返回空字符串，调用方在签发/校验时会报错——属于"快速失败"。
func getSecretKey() string {
	return os.Getenv("SECRET_KEY")
}

// getIssuer 允许通过环境变量覆盖默认 iss，便于多实例区分。
func getIssuer() string {
	if v := strings.TrimSpace(os.Getenv("JWT_ISSUER")); v != "" {
		return v
	}
	return defaultIssuer
}

// getAudience 允许通过环境变量覆盖默认 aud。
func getAudience() string {
	if v := strings.TrimSpace(os.Getenv("JWT_AUDIENCE")); v != "" {
		return v
	}
	return defaultAudience
}

// GenerateAllTokens 同时生成 access token 与 refresh token。
// access token 内含完整身份字段；refresh token 仅在 sub/uid 上保留必要信息。
// 返回的 error 非 nil 时调用方需自行决定如何上抛，避免在底层做 log.Panic。
func GenerateAllTokens(email, uuidStr, name, userType, uid string) (string, string, error) {
	secret := getSecretKey()
	if secret == "" {
		return "", "", errors.New("SECRET_KEY is not set")
	}

	now := time.Now()
	issuer := getIssuer()
	audience := jwt.ClaimStrings{getAudience()}

	accessClaims := &SignedDetails{
		Email: email,
		UUID:  uuidStr,
		Name:  name,
		Uid:   uid,
		Role:  userType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   uid,
			Audience:  audience,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenTTL)),
		},
	}

	// refresh token 仍然带上 uid/role，以便刷新时能直接重建 access token，
	// 不必再回查数据库。
	refreshClaims := &SignedDetails{
		Email: email,
		UUID:  uuidStr,
		Uid:   uid,
		Role:  userType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   uid,
			Audience:  audience,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshTokenTTL)),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(secret))
	if err != nil {
		return "", "", fmt.Errorf("sign access token: %w", err)
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(secret))
	if err != nil {
		return "", "", fmt.Errorf("sign refresh token: %w", err)
	}

	return token, refreshToken, nil
}

// ValidateToken 校验 JWT 并返回 claims。
// 安全要点：
//  1. 强制校验签名算法白名单（仅允许 HS256），抵御 alg=none / 算法混淆攻击；
//  2. 校验 iss/aud，避免跨服务 token 复用；
//  3. exp/iat/nbf 由 jwt v5 内置自动校验。
//
// 返回的 string 是给前端展示的简明错误，避免泄露内部细节；
// 详细原因仍写日志便于排查。
func ValidateToken(signedToken string) (*SignedDetails, string) {
	secret := getSecretKey()
	if secret == "" {
		log.Printf("ValidateToken: SECRET_KEY is empty")
		return nil, "server misconfiguration"
	}

	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		jwt.WithIssuer(getIssuer()),
		jwt.WithAudience(getAudience()),
	)
	if err != nil {
		log.Printf("ValidateToken parse error: %v", err)
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, "token is expired"
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, "token not active yet"
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			return nil, "invalid token signature"
		default:
			return nil, "invalid token"
		}
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok || !token.Valid {
		return nil, "invalid token"
	}

	return claims, ""
}
