package config

import (
	"fmt"
	"net/url"
	"time"
)

// CognitoConfig Cognito認証の設定
type CognitoConfig struct {
	Enabled          bool
	Region           string
	UserPoolID       string
	ClientID         string
	ClientSecret     string
	Endpoint         string
	JWKCacheDuration time.Duration
	TokenExpiration  time.Duration
}

// IsDevelopment 開発環境かどうかを判定
func (c *CognitoConfig) IsDevelopment() bool {
	return c.Endpoint != ""
}

// GetJWKURL JWKエンドポイントURLを取得
func (c *CognitoConfig) GetJWKURL() string {
	if c.Endpoint != "" {
		return c.Endpoint + "/" + c.UserPoolID + "/.well-known/jwks.json"
	}
	return "https://cognito-idp." + c.Region + ".amazonaws.com/" + c.UserPoolID + "/.well-known/jwks.json"
}

// GetIssuer Issuerを取得
func (c *CognitoConfig) GetIssuer() string {
	if c.Endpoint != "" {
		u, err := url.Parse(c.Endpoint)
		if err != nil {
			return "http://0.0.0.0:9229/" + c.UserPoolID
		}
		host := u.Hostname()
		port := u.Port()
		if host == "cognito-local" || host == "localhost" {
			host = "0.0.0.0"
		}
		if port != "" {
			return fmt.Sprintf("http://%s:%s/%s", host, port, c.UserPoolID)
		}
		return fmt.Sprintf("http://%s/%s", host, c.UserPoolID)
	}
	return "https://cognito-idp." + c.Region + ".amazonaws.com/" + c.UserPoolID
}

// SetDefaults デフォルト値を設定
func (c *CognitoConfig) SetDefaults() {
	if c.TokenExpiration == 0 {
		c.TokenExpiration = time.Hour
	}
	if c.JWKCacheDuration == 0 {
		c.JWKCacheDuration = time.Hour
	}
}
