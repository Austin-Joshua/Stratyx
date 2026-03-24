package security

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"uid"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret               []byte
	accessTokenTTL       time.Duration
	refreshTokenTTL      time.Duration
	refreshTokenTTLDays  int
	accessTokenTTLMinute int
}

func NewJWTManager(secret string, accessTTLMinutes, refreshTTLDays int) *JWTManager {
	return &JWTManager{
		secret:               []byte(secret),
		accessTokenTTL:       time.Duration(accessTTLMinutes) * time.Minute,
		refreshTokenTTL:      time.Duration(refreshTTLDays) * 24 * time.Hour,
		refreshTokenTTLDays:  refreshTTLDays,
		accessTokenTTLMinute: accessTTLMinutes,
	}
}

func (j *JWTManager) GenerateAccessToken(userID string) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTManager) GenerateRefreshToken(userID string) (string, time.Time, error) {
	expiresAt := time.Now().Add(j.refreshTokenTTL)
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	raw, err := token.SignedString(j.secret)
	return raw, expiresAt, err
}

func (j *JWTManager) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(_ *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}
