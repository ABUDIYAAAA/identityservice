package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid or malformed token")
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	UserID       string `json:"user_id"`
	Role         string `json:"role"`
	TokenVersion int    `json:"token_version"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	RefreshTokenHash string
	AccessJTI        string
	RefreshJTI       string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

type JWTManager struct {
	accessSecret    []byte
	refreshSecret   []byte
	accessDuration  time.Duration
	refreshDuration time.Duration
	cookieDomain    string
	isProduction    bool
}

func NewJWTManager(accessSec, refreshSec string, accessTTL, refreshTTL time.Duration, cookieDomain string, isProd bool) *JWTManager {
	return &JWTManager{
		accessSecret:    []byte(accessSec),
		refreshSecret:   []byte(refreshSec),
		accessDuration:  accessTTL,
		refreshDuration: refreshTTL,
		cookieDomain:    cookieDomain,
		isProduction:    isProd,
	}
}

func (m *JWTManager) GenerateTokenPair(userID, role string, tokenVersion int) (*TokenPair, error) {
	now := time.Now()
	accessJTI := uuid.New().String()
	refreshJTI := uuid.New().String()

	accessExp := now.Add(m.accessDuration)
	refreshExp := now.Add(m.refreshDuration)

	accessClaims := &Claims{
		UserID:       userID,
		Role:         role,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessJTI,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExp),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessSigned, err := accessToken.SignedString(m.accessSecret)
	if err != nil {
		return nil, err
	}

	refreshClaims := &Claims{
		UserID:       userID,
		Role:         role,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshJTI,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(refreshExp),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshSigned, err := refreshToken.SignedString(m.refreshSecret)
	if err != nil {
		return nil, err
	}

	h := sha256.Sum256([]byte(refreshSigned))
	refreshHash := hex.EncodeToString(h[:])

	return &TokenPair{
		AccessToken:      accessSigned,
		RefreshToken:     refreshSigned,
		RefreshTokenHash: refreshHash,
		AccessJTI:        accessJTI,
		RefreshJTI:       refreshJTI,
		AccessExpiresAt:  accessExp,
		RefreshExpiresAt: refreshExp,
	}, nil
}

func (m *JWTManager) VerifyAccessToken(tokenStr string) (*Claims, error) {
	return m.verifyToken(tokenStr, m.accessSecret)
}

func (m *JWTManager) VerifyRefreshToken(tokenStr string) (*Claims, error) {
	return m.verifyToken(tokenStr, m.refreshSecret)
}

func (m *JWTManager) verifyToken(tokenStr string, secret []byte) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

const RefreshCookieName = "refresh_token"

func (m *JWTManager) SetRefreshCookie(w http.ResponseWriter, refreshToken string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    refreshToken,
		Path:     "/api/v1/auth",
		Domain:   m.cookieDomain,
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true,
		Secure:   m.isProduction,
		SameSite: http.SameSiteStrictMode,
	})
}

func (m *JWTManager) ClearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    "",
		Path:     "/api/v1/auth",
		Domain:   m.cookieDomain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   m.isProduction,
		SameSite: http.SameSiteStrictMode,
	})
}
