# 実装パターン集

## バックエンド実装パターン

### モデル定義
```go
type Feature struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
    Name      string    `gorm:"not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### リポジトリ実装
```go
type FeatureRepository interface {
    Create(ctx context.Context, f *model.Feature) error
    FindByID(ctx context.Context, id uuid.UUID) (*model.Feature, error)
    Update(ctx context.Context, f *model.Feature) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

### サービス実装
```go
type FeatureService struct {
    repo   repository.FeatureRepository
    logger *zap.Logger
}

func (s *FeatureService) Create(ctx context.Context, input dto.CreateFeatureInput) (*dto.FeatureResponse, error) {
    // バリデーション
    // ビジネスロジック
    // リポジトリ呼び出し
}
```

### ハンドラー実装
```go
func (h *FeatureHandler) Create(c *gin.Context) {
    var input dto.CreateFeatureInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    // サービス呼び出し
}
```

## フロントエンド実装パターン

### APIクライアント
```typescript
export const createFeature = async (input: CreateFeatureInput) => {
  const client = createPresetApiClient('auth');
  const response = await client.post('/features', input);
  return convertSnakeToCamel<FeatureResponse>(response.data);
};
```

### カスタムフック
```typescript
export const useFeatures = () => {
  return useQuery({
    queryKey: queryKeys.features(),
    queryFn: getFeatures,
  });
};
```

### コンポーネント
```typescript
export const FeatureList: React.FC = () => {
  const { data, isLoading } = useFeatures();
  // UI実装
};
```
