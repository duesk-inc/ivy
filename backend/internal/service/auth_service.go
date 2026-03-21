package service

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/duesk/ivy/internal/config"
	"github.com/duesk/ivy/internal/dto"
	"github.com/duesk/ivy/internal/repository"
	"go.uber.org/zap"
)

// AuthService 認証サービスインターフェース
type AuthService interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error)
	Logout(ctx context.Context, accessToken string) error
}

type authService struct {
	cfg           *config.Config
	userRepo      repository.UserRepository
	cognitoClient *cognitoidentityprovider.Client
	logger        *zap.Logger
}

// NewAuthService 認証サービスを作成
func NewAuthService(cfg *config.Config, userRepo repository.UserRepository, logger *zap.Logger) AuthService {
	s := &authService{
		cfg:      cfg,
		userRepo: userRepo,
		logger:   logger,
	}

	if cfg.Cognito.Enabled {
		awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
			awsconfig.WithRegion(cfg.Cognito.Region),
		)
		if err == nil {
			opts := func(o *cognitoidentityprovider.Options) {
				if cfg.Cognito.Endpoint != "" {
					o.BaseEndpoint = aws.String(cfg.Cognito.Endpoint)
				}
			}
			s.cognitoClient = cognitoidentityprovider.NewFromConfig(awsCfg, opts)
		}
	}

	return s
}

// Login ログイン
func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	if !s.cfg.Cognito.Enabled {
		return s.devLogin(ctx, req)
	}

	output, err := s.cognitoClient.InitiateAuth(ctx, &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeUserPasswordAuth,
		ClientId: aws.String(s.cfg.Cognito.ClientID),
		AuthParameters: map[string]string{
			"USERNAME": req.Email,
			"PASSWORD": req.Password,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("認証に失敗しました: %w", err)
	}

	if output.AuthenticationResult == nil {
		return nil, fmt.Errorf("認証結果がありません")
	}

	return &dto.LoginResponse{
		AccessToken:  aws.ToString(output.AuthenticationResult.AccessToken),
		RefreshToken: aws.ToString(output.AuthenticationResult.RefreshToken),
		ExpiresIn:    int(output.AuthenticationResult.ExpiresIn),
	}, nil
}

// devLogin 開発環境用ログイン
func (s *authService) devLogin(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	token := fmt.Sprintf("dev.00000000-0000-0000-0000-000000000001.%d", time.Now().Unix())

	return &dto.LoginResponse{
		AccessToken:  token,
		RefreshToken: "dev-refresh-token",
		ExpiresIn:    3600,
		User: dto.UserResponse{
			ID:    "00000000-0000-0000-0000-000000000001",
			Email: req.Email,
			Name:  "開発ユーザー",
			Role:  "admin",
		},
	}, nil
}

// RefreshToken トークンリフレッシュ
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error) {
	if !s.cfg.Cognito.Enabled {
		return &dto.LoginResponse{
			AccessToken:  fmt.Sprintf("dev.00000000-0000-0000-0000-000000000001.%d", time.Now().Unix()),
			RefreshToken: refreshToken,
			ExpiresIn:    3600,
		}, nil
	}

	output, err := s.cognitoClient.InitiateAuth(ctx, &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeRefreshTokenAuth,
		ClientId: aws.String(s.cfg.Cognito.ClientID),
		AuthParameters: map[string]string{
			"REFRESH_TOKEN": refreshToken,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("トークンリフレッシュに失敗しました: %w", err)
	}

	return &dto.LoginResponse{
		AccessToken: aws.ToString(output.AuthenticationResult.AccessToken),
		ExpiresIn:   int(output.AuthenticationResult.ExpiresIn),
	}, nil
}

// Logout ログアウト
func (s *authService) Logout(ctx context.Context, accessToken string) error {
	if !s.cfg.Cognito.Enabled {
		return nil
	}

	_, err := s.cognitoClient.GlobalSignOut(ctx, &cognitoidentityprovider.GlobalSignOutInput{
		AccessToken: aws.String(accessToken),
	})
	if err != nil {
		s.logger.Warn("ログアウト処理でエラー", zap.Error(err))
	}
	return nil
}

