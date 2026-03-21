---
paths: backend/internal/service/**/*.go
---

# Service層実装規約

## 基本構造

### Service構造体の定義

```go
type engineerService struct {
    db               *gorm.DB                         // トランザクション用
    engineerRepo     repository.EngineerRepository   // インターフェース型で定義
    userRepo         repository.UserRepository
    weeklyReportRepo repository.WeeklyReportRepository
    cognitoAuth      *CognitoAuthService              // 外部サービス連携
    config           *config.Config
    logger           *zap.Logger
}
```

### フィールドの順序（推奨）

1. `db` (*gorm.DB) - トランザクション管理用
2. Repository群 - インターフェース型
3. 他Service/外部依存
4. `config` (*config.Config)
5. `logger` (*zap.Logger)

---

## コンストラクタ

### 標準パターン

```go
func NewEngineerService(
    db *gorm.DB,
    engineerRepo repository.EngineerRepository,
    userRepo repository.UserRepository,
    weeklyReportRepo repository.WeeklyReportRepository,
    cognitoAuth *CognitoAuthService,
    config *config.Config,
    logger *zap.Logger,
) EngineerService {  // 戻り値はインターフェース型
    return &engineerService{
        db:               db,
        engineerRepo:     engineerRepo,
        userRepo:         userRepo,
        weeklyReportRepo: weeklyReportRepo,
        cognitoAuth:      cognitoAuth,
        config:           config,
        logger:           logger,
    }
}
```

### TransactionManager を使用する場合

```go
func NewLeaveService(
    db *gorm.DB,
    leaveRepo repository.LeaveRepository,
    logger *zap.Logger,
) LeaveService {
    return &leaveService{
        txManager: transaction.NewTransactionManager(db, logger),
        leaveRepo: leaveRepo,
        logger:    logger,
    }
}
```

---

## メソッドシグネチャ

### 基本ルール

- 第1引数は必ず `ctx context.Context`
- 戻り値は `(T, error)` または `([]T, int64, error)`

### パターン一覧

```go
// 単一取得
func (s *engineerService) GetEngineerByID(ctx context.Context, id string) (*model.User, error)

// 一覧取得（ページネーション付き）
func (s *engineerService) GetEngineers(ctx context.Context, filters repository.EngineerFilters) ([]*model.User, int64, error)

// 作成
func (s *engineerService) CreateEngineer(ctx context.Context, input CreateEngineerInput) (*model.User, error)

// 更新
func (s *engineerService) UpdateEngineer(ctx context.Context, id string, input UpdateEngineerInput) (*model.User, error)

// 削除
func (s *engineerService) DeleteEngineer(ctx context.Context, id string) error

// 複合操作（トランザクション）
func (s *WeeklyReportService) Create(ctx context.Context, report *model.WeeklyReport, dailyRecords []*model.DailyRecord) error
```

---

## トランザクション管理

### パターン1: db.Transaction（標準）

```go
func (s *engineerService) CreateEngineer(ctx context.Context, input CreateEngineerInput) (*model.User, error) {
    var user *model.User

    err := s.db.Transaction(func(tx *gorm.DB) error {
        // トランザクション用リポジトリを新規作成
        txEngineerRepo := repository.NewEngineerRepository(tx, s.logger)
        txHistoryRepo := repository.NewStatusHistoryRepository(tx, s.logger)

        // ユーザー作成
        user = &model.User{
            ID:    uuid.New().String(),
            Email: input.Email,
            // ...
        }
        if err := txEngineerRepo.Create(ctx, user); err != nil {
            return fmt.Errorf("ユーザー作成失敗: %w", err)
        }

        // ステータス履歴作成
        history := &model.EngineerStatusHistory{
            UserID: user.ID,
            Status: "active",
        }
        if err := txHistoryRepo.Create(ctx, history); err != nil {
            return fmt.Errorf("履歴作成失敗: %w", err)
        }

        return nil
    })

    if err != nil {
        return nil, err
    }

    return user, nil
}
```

### パターン2: TransactionManager

```go
func (s *leaveService) executeInTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
    return s.txManager.ExecuteInTransaction(ctx, fn)
}

func (s *leaveService) CreateLeaveRequest(ctx context.Context, req dto.LeaveRequestRequest) (dto.LeaveRequestResponse, error) {
    var response dto.LeaveRequestResponse

    err := s.executeInTransaction(ctx, func(tx *gorm.DB) error {
        txLeaveRepo := repository.NewLeaveRepository(tx, s.logger)

        createdRequest, err := txLeaveRepo.CreateLeaveRequest(ctx, request)
        if err != nil {
            return logger.LogAndWrapError(s.logger, err, "休暇申請の作成に失敗しました")
        }

        response = mapToResponse(createdRequest)
        return nil
    })

    return response, err
}
```

### 重要: トランザクション内のリポジトリ

```go
// ❌ 禁止: 既存リポジトリの使用（別接続になる）
err := s.db.Transaction(func(tx *gorm.DB) error {
    return s.engineerRepo.Create(ctx, user)  // これは別接続！
})

// ✅ 正しい: トランザクション用リポジトリを新規作成
err := s.db.Transaction(func(tx *gorm.DB) error {
    txRepo := repository.NewEngineerRepository(tx, s.logger)
    return txRepo.Create(ctx, user)
})
```

---

## エラーハンドリング

### 基本パターン

```go
// シンプルなエラー伝播
user, err := s.engineerRepo.FindEngineerByID(ctx, id)
if err != nil {
    return nil, err
}

// ラッピング + コンテキスト追加
if err != nil {
    return nil, fmt.Errorf("エンジニア取得失敗 (id=%s): %w", id, err)
}
```

### logger.LogAndWrapError（推奨）

```go
// ログ出力 + エラーラッピングを同時に行う
if err != nil {
    return nil, logger.LogAndWrapError(s.logger, err, "休暇種別の取得に失敗しました",
        zap.String("leave_type_id", req.LeaveTypeID),
        zap.String("user_id", req.UserID))
}
```

### gorm.ErrRecordNotFound の処理

```go
if err == gorm.ErrRecordNotFound {
    return nil, fmt.Errorf(message.MsgEngineerNotFound+": %w", err)
}
```

### ビジネスルール検証

```go
// 重複チェック
exists, err := s.engineerRepo.ExistsByEmail(ctx, input.Email)
if err != nil {
    return nil, fmt.Errorf("メールアドレスの確認中にエラー: %w", err)
}
if exists {
    return nil, errors.New("このメールアドレスは既に使用されています")
}

// ステータス検証
if !model.IsValidEngineerStatus(status) {
    return nil, fmt.Errorf("無効なステータスです: %s", status)
}
```

---

## メッセージ定数の活用

```go
import "github.com/duesk/monstera/internal/message"

// エラーメッセージ
return dto.LeaveRequestResponse{}, errors.New(message.MsgReasonRequired)
return nil, fmt.Errorf(message.MsgLeaveBalanceExceededFormat, balance.RemainingDays)

// 成功メッセージ（Handler層で使用）
// message.MsgDeleteSuccess
```

---

## ビジネスロジックの実装パターン

### 検証 → 処理 → 保存

```go
func (s *leaveService) CreateLeaveRequest(ctx context.Context, req dto.LeaveRequestRequest) (dto.LeaveRequestResponse, error) {
    // STEP 1: 関連データ取得
    leaveType, err := s.leaveRepo.GetLeaveTypeByID(ctx, req.LeaveTypeID)
    if err != nil {
        return dto.LeaveRequestResponse{}, logger.LogAndWrapError(s.logger, err, "休暇種別が見つかりません")
    }

    // STEP 2: ビジネスルール検証
    if leaveType.ReasonRequired && req.Reason == "" {
        return dto.LeaveRequestResponse{}, errors.New(message.MsgReasonRequired)
    }

    if req.IsHourlyBased && !leaveType.IsHourlyAvailable {
        return dto.LeaveRequestResponse{}, errors.New(message.MsgHourlyLeaveNotAllowed)
    }

    // STEP 3: 残日数検証
    if leaveType.RequiresBalance {
        balance, err := s.leaveRepo.GetUserLeaveBalanceByType(ctx, req.UserID, req.LeaveTypeID)
        if err != nil {
            return dto.LeaveRequestResponse{}, logger.LogAndWrapError(s.logger, err, "残日数取得失敗")
        }
        if balance.RemainingDays < req.TotalDays {
            return dto.LeaveRequestResponse{}, fmt.Errorf(message.MsgLeaveBalanceExceededFormat, balance.RemainingDays)
        }
    }

    // STEP 4: モデル構築
    request := model.LeaveRequest{
        ID:          uuid.New().String(),
        UserID:      req.UserID,
        LeaveTypeID: leaveType.ID,
        Status:      "pending",
        // ...
    }

    // STEP 5: トランザクション内で保存
    var response dto.LeaveRequestResponse
    err = s.executeInTransaction(ctx, func(tx *gorm.DB) error {
        txRepo := repository.NewLeaveRepository(tx, s.logger)
        created, err := txRepo.CreateLeaveRequest(ctx, request)
        if err != nil {
            return err
        }
        response = mapToResponse(created)
        return nil
    })

    return response, err
}
```

---

## DTO変換

### 単一変換

```go
func (s *projectService) toMinimalDTO(project *model.Project) dto.ProjectMinimalDTO {
    return dto.ProjectMinimalDTO{
        ID:          project.ID,
        Name:        project.Name,
        ClientName:  project.Client.Name,
        // ...
    }
}
```

### スライス変換

```go
results := make([]dto.LeaveTypeResponse, len(leaveTypes))
for i, lt := range leaveTypes {
    results[i] = dto.LeaveTypeResponse{
        ID:   lt.ID,
        Code: lt.Code,
        Name: lt.Name,
    }
}

// nilスライスの初期化（JSONで[]を返すため）
if results == nil {
    results = []dto.LeaveTypeResponse{}
}
```

---

## インターフェース定義

```go
// interfaces.go
type EngineerService interface {
    GetEngineers(ctx context.Context, filters repository.EngineerFilters) ([]*model.User, int64, error)
    GetEngineerByID(ctx context.Context, id string) (*model.User, error)
    CreateEngineer(ctx context.Context, input CreateEngineerInput) (*model.User, error)
    UpdateEngineer(ctx context.Context, id string, input UpdateEngineerInput) (*model.User, error)
    DeleteEngineer(ctx context.Context, id string) error
}
```

---

## チェックリスト

新しいServiceメソッドを作成する際：

- [ ] 第1引数が `ctx context.Context`
- [ ] 戻り値が `(T, error)` または `([]T, int64, error)` 形式
- [ ] Repository はインターフェース型で受け取っている
- [ ] 複合操作は `db.Transaction` 内で実行
- [ ] トランザクション内では新規リポジトリを生成
- [ ] エラーは `fmt.Errorf` または `logger.LogAndWrapError` でラッピング
- [ ] メッセージ定数（`message` パッケージ）を使用
- [ ] ビジネスルール検証を保存前に実施
- [ ] nilスライスは空スライスに初期化
