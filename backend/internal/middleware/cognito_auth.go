package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	cfg "github.com/duesk/ivy/internal/config"
	"github.com/duesk/ivy/internal/model"
	"github.com/duesk/ivy/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// JWK JSON Web Key構造体
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
	Alg string `json:"alg"`
}

// JWKSet JWK Set構造体
type JWKSet struct {
	Keys []JWK `json:"keys"`
}

// JWKCache JWKキャッシュ構造体
type JWKCache struct {
	keys      map[string]*rsa.PublicKey
	expiresAt time.Time
	mutex     sync.RWMutex
}

// CognitoAuthMiddleware Cognito認証ミドルウェア
type CognitoAuthMiddleware struct {
	config       *cfg.Config
	userRepo     repository.UserRepository
	logger       *zap.Logger
	jwkCache     *JWKCache
	httpClient   *http.Client
	cognitoClient *cognitoidentityprovider.Client
}

// NewCognitoAuthMiddleware 新しいCognito認証ミドルウェアを作成
func NewCognitoAuthMiddleware(
	config *cfg.Config,
	userRepo repository.UserRepository,
	logger *zap.Logger,
) *CognitoAuthMiddleware {
	m := &CognitoAuthMiddleware{
		config:   config,
		userRepo: userRepo,
		logger:   logger,
		jwkCache: &JWKCache{
			keys: make(map[string]*rsa.PublicKey),
		},
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// Cognito clientを初期化（JITプロビジョニング用）
	if config.Cognito.Enabled {
		awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
			awsconfig.WithRegion(config.Cognito.Region),
		)
		if err == nil {
			opts := func(o *cognitoidentityprovider.Options) {
				if config.Cognito.Endpoint != "" {
					o.BaseEndpoint = aws.String(config.Cognito.Endpoint)
				}
			}
			m.cognitoClient = cognitoidentityprovider.NewFromConfig(awsCfg, opts)
		}
	}

	return m
}

// AuthRequired 認証が必要なエンドポイント用のミドルウェア
func (m *CognitoAuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.Cognito.Enabled {
			m.setDevelopmentUser(c)
			c.Next()
			return
		}

		token := m.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
			c.Abort()
			return
		}

		claims, err := m.validateToken(token)
		if err != nil {
			m.logger.Error("トークン検証エラー", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "無効なトークンです"})
			c.Abort()
			return
		}

		user, err := m.getOrCreateUser(c.Request.Context(), claims)
		if err != nil {
			if err.Error() == "engineer_role_forbidden" {
				c.JSON(http.StatusForbidden, gin.H{"error": "このアプリは営業担当者向けです。"})
				c.Abort()
				return
			}
			m.logger.Error("ユーザー情報取得エラー", zap.Error(err))
			// Cognito API障害の場合は設計書4.1に従ったメッセージを返す
			if strings.Contains(err.Error(), "cognito") {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "認証サービスに接続できません。しばらく待ってから再度お試しください。"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "ユーザー情報が取得できません"})
			}
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("role", string(user.Role))
		c.Set("cognito_sub", claims["sub"])
		c.Next()
	}
}

// extractToken リクエストからトークンを抽出
func (m *CognitoAuthMiddleware) extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}
	cookie, err := c.Cookie("access_token")
	if err == nil && cookie != "" {
		return cookie
	}
	return ""
}

// validateToken トークンを検証
func (m *CognitoAuthMiddleware) validateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("予期しない署名方法: %v", token.Header["alg"])
		}
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kidヘッダーが見つかりません")
		}
		return m.getPublicKey(kid)
	})
	if err != nil {
		return nil, fmt.Errorf("トークンのパースに失敗しました: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("無効なトークンです")
	}

	if err := m.validateClaims(claims); err != nil {
		return nil, err
	}
	return claims, nil
}

// validateClaims Claimsを検証
func (m *CognitoAuthMiddleware) validateClaims(claims jwt.MapClaims) error {
	iss, ok := claims["iss"].(string)
	if !ok {
		return fmt.Errorf("issuerが見つかりません")
	}
	expectedIssuer := m.config.Cognito.GetIssuer()
	if iss != expectedIssuer {
		return fmt.Errorf("無効なissuer: expected %s, got %s", expectedIssuer, iss)
	}

	tokenUse, ok := claims["token_use"].(string)
	if !ok {
		return fmt.Errorf("token_useが見つかりません")
	}
	if tokenUse != "id" && tokenUse != "access" {
		return fmt.Errorf("無効なtoken_use: %s", tokenUse)
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return fmt.Errorf("有効期限が見つかりません")
	}
	if time.Now().Unix() > int64(exp) {
		return fmt.Errorf("トークンの有効期限が切れています")
	}

	return nil
}

// getOrCreateUser ClaimsからユーザーDBレコードを取得（なければJITプロビジョニング）
func (m *CognitoAuthMiddleware) getOrCreateUser(ctx context.Context, claims jwt.MapClaims) (*model.User, error) {
	cognitoSub, ok := claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("subが見つかりません")
	}

	// DBからユーザーを取得
	user, err := m.userRepo.GetByCognitoSub(ctx, cognitoSub)
	if err == nil {
		return user, nil
	}

	// JITプロビジョニング: Cognitoからユーザー情報を取得してIvy DBに作成
	return m.jitProvision(ctx, cognitoSub)
}

// jitProvision JITプロビジョニング
func (m *CognitoAuthMiddleware) jitProvision(ctx context.Context, cognitoSub string) (*model.User, error) {
	if m.cognitoClient == nil {
		return nil, fmt.Errorf("cognito client not initialized")
	}

	// Cognitoからユーザー情報を取得
	output, err := m.cognitoClient.AdminGetUser(ctx, &cognitoidentityprovider.AdminGetUserInput{
		UserPoolId: aws.String(m.config.Cognito.UserPoolID),
		Username:   aws.String(cognitoSub),
	})
	if err != nil {
		return nil, fmt.Errorf("cognito AdminGetUser failed: %w", err)
	}

	// ユーザー属性を取得
	var email, name, jobCategory string
	var monsteraRole int
	for _, attr := range output.UserAttributes {
		switch aws.ToString(attr.Name) {
		case "email":
			email = aws.ToString(attr.Value)
		case "name":
			name = aws.ToString(attr.Value)
		case "custom:role":
			fmt.Sscanf(aws.ToString(attr.Value), "%d", &monsteraRole)
		case "custom:job_category":
			jobCategory = aws.ToString(attr.Value)
		}
	}

	// ロールマッピング（Monstera → Ivy）
	ivyRole := mapRole(monsteraRole, jobCategory)
	if ivyRole == "" {
		return nil, fmt.Errorf("engineer_role_forbidden")
	}

	if name == "" {
		name = email
	}

	// Ivy usersテーブルに作成
	user := &model.User{
		CognitoSub: cognitoSub,
		Email:      email,
		Name:       name,
		Role:       model.Role(ivyRole),
	}
	if err := m.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("user creation failed: %w", err)
	}

	m.logger.Info("JITプロビジョニング完了",
		zap.String("cognito_sub", cognitoSub),
		zap.String("email", email),
		zap.String("role", ivyRole),
	)

	return user, nil
}

// mapRole Monsteraのロールをivyのロールにマッピング
func mapRole(monsteraRole int, jobCategory string) string {
	// Monstera Role=3(admin) → Ivy admin
	if monsteraRole == 3 {
		return "admin"
	}
	// Monstera JobCategory="engineer" → Ivyアクセス拒否
	if jobCategory == "engineer" {
		return ""
	}
	// Monstera JobCategory="sales" → Ivy sales
	if jobCategory == "sales" {
		return "sales"
	}
	// デフォルト → sales
	return "sales"
}

// getPublicKey Kid（Key ID）に対応する公開鍵を取得
func (m *CognitoAuthMiddleware) getPublicKey(kid string) (*rsa.PublicKey, error) {
	m.jwkCache.mutex.RLock()
	if key, exists := m.jwkCache.keys[kid]; exists && time.Now().Before(m.jwkCache.expiresAt) {
		m.jwkCache.mutex.RUnlock()
		return key, nil
	}
	m.jwkCache.mutex.RUnlock()

	return m.fetchAndCacheJWK(kid)
}

// fetchAndCacheJWK JWKSetを取得してキャッシュに保存
func (m *CognitoAuthMiddleware) fetchAndCacheJWK(kid string) (*rsa.PublicKey, error) {
	m.jwkCache.mutex.Lock()
	defer m.jwkCache.mutex.Unlock()

	if key, exists := m.jwkCache.keys[kid]; exists && time.Now().Before(m.jwkCache.expiresAt) {
		return key, nil
	}

	jwkURL := m.config.Cognito.GetJWKURL()
	resp, err := m.httpClient.Get(jwkURL)
	if err != nil {
		return nil, fmt.Errorf("JWKSetの取得に失敗しました: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKSetの取得に失敗しました: status %d", resp.StatusCode)
	}

	var jwkSet JWKSet
	if err := json.NewDecoder(resp.Body).Decode(&jwkSet); err != nil {
		return nil, fmt.Errorf("JWKSetのパースに失敗しました: %w", err)
	}

	m.jwkCache.keys = make(map[string]*rsa.PublicKey)
	for _, jwk := range jwkSet.Keys {
		if jwk.Kty != "RSA" {
			continue
		}
		publicKey, err := parseRSAPublicKey(jwk)
		if err != nil {
			m.logger.Error("RSA公開鍵のパースに失敗", zap.String("kid", jwk.Kid), zap.Error(err))
			continue
		}
		m.jwkCache.keys[jwk.Kid] = publicKey
	}

	m.jwkCache.expiresAt = time.Now().Add(m.config.Cognito.JWKCacheDuration)

	if key, exists := m.jwkCache.keys[kid]; exists {
		return key, nil
	}
	return nil, fmt.Errorf("指定されたkid '%s' の公開鍵が見つかりません", kid)
}

// parseRSAPublicKey JWKからRSA公開鍵をパース
func parseRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("nのデコードに失敗: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("eのデコードに失敗: %w", err)
	}

	var eInt int
	for _, b := range eBytes {
		eInt = eInt*256 + int(b)
	}

	return &rsa.PublicKey{
		N: big.NewInt(0).SetBytes(nBytes),
		E: eInt,
	}, nil
}

// setDevelopmentUser 開発環境用のダミーユーザーを設定
func (m *CognitoAuthMiddleware) setDevelopmentUser(c *gin.Context) {
	// 開発環境ではDBからユーザーを検索、なければデフォルト管理者
	token := m.extractToken(c)
	if token != "" && strings.HasPrefix(token, "dev.") {
		parts := strings.Split(token, ".")
		if len(parts) >= 2 {
			userID := parts[1]
			user, err := m.userRepo.GetByID(c.Request.Context(), userID)
			if err == nil {
				c.Set("user", user)
				c.Set("user_id", user.ID)
				c.Set("role", string(user.Role))
				c.Set("cognito_sub", user.ID)
				return
			}
		}
	}

	// デフォルト開発ユーザー
	devUser := &model.User{
		ID:         "00000000-0000-0000-0000-000000000001",
		CognitoSub: "00000000-0000-0000-0000-000000000001",
		Email:      "admin@duesk.co.jp",
		Name:       "開発ユーザー",
		Role:       model.RoleAdmin,
	}
	c.Set("user", devUser)
	c.Set("user_id", devUser.ID)
	c.Set("role", string(devUser.Role))
	c.Set("cognito_sub", devUser.ID)
}
