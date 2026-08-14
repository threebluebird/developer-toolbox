package models

// Settings stores application-level preferences.
type Settings struct {
	Theme       string `json:"theme"`
	Language    string `json:"language"`
	StoragePath string `json:"storagePath"`
	JSONIndent  int    `json:"jsonIndent"`
}
