package message

// 認証系メッセージ
const (
	MsgAuthRequired          = "認証が必要です"
	MsgInvalidToken          = "無効なトークンです"
	MsgUserNotFound          = "ユーザーが見つかりません"
	MsgEngineerRoleForbidden = "このアプリは営業担当者向けです。"
	MsgForbidden             = "権限が不足しています"
	MsgAdminRequired         = "管理者権限が必要です"
	MsgCognitoError          = "認証サービスに接続できません。しばらく待ってから再度お試しください。"
)

// マッチング系メッセージ
const (
	MsgMatchingNotFound     = "マッチング結果が見つかりません"
	MsgMatchingFailed       = "マッチング処理に失敗しました"
	MsgAIServerBusy         = "AIサーバーが混雑しています。しばらく待ってから再度お試しください。"
	MsgInvalidRequest       = "リクエストが不正です"
	MsgEngineerInfoRequired = "エンジニア情報（テキストまたはファイル）が必要です"
)

// ファイル系メッセージ
const (
	MsgFileTooLarge       = "ファイルサイズが大きすぎます（上限10MB）"
	MsgUnsupportedFormat  = "対応していないファイル形式です。Excel(.xlsx/.xls)またはPDF(.pdf)をアップロードしてください。"
	MsgFileParseFailed    = "ファイルの読み取りに失敗しました。対応形式（Excel/PDF）か確認してください。"
	MsgFileCorrupted      = "ファイルが破損しています"
	MsgPasswordProtected  = "パスワード保護されたファイルは処理できません"
)

// 設定系メッセージ
const (
	MsgSettingNotFound    = "設定が見つかりません"
	MsgSettingUpdateFailed = "設定の更新に失敗しました"
)

// 汎用エラーメッセージ
const (
	MsgInternalError  = "サーバー内部エラーが発生しました"
	MsgRateLimited    = "リクエスト数が上限に達しました。しばらく待ってから再度お試しください"
	MsgNetworkError   = "通信エラーが発生しました。ネットワーク接続を確認してください。"
)
