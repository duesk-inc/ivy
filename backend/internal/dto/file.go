package dto

// FileParseResponse ファイルパースレスポンス
type FileParseResponse struct {
	Text          string   `json:"text"`
	FileKey       string   `json:"file_key"`
	FileName      string   `json:"file_name"`
	ParseWarnings []string `json:"parse_warnings"`
}
