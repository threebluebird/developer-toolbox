package config

import "developer-toolbox/backend/models"

// DefaultSettings returns the configured default application settings.
func DefaultSettings() models.Settings {
	return models.Settings{
		Theme:       DefaultTheme,
		Language:    DefaultLanguage,
		StoragePath: "",
		JSONIndent:  DefaultJSONIndent,
	}
}

// NormalizeSettings 将旧版本、缺字段或非法配置合并为当前版本支持的设置。
// 读取和保存共用此逻辑，避免前后端对合法取值的理解不一致。
func NormalizeSettings(settings models.Settings) models.Settings {
	defaults := DefaultSettings()
	if settings.Theme != "dark" && settings.Theme != "light" && settings.Theme != "system" {
		settings.Theme = defaults.Theme
	}
	if settings.Language != "en" && settings.Language != "zh" {
		settings.Language = defaults.Language
	}
	if settings.JSONIndent != 2 && settings.JSONIndent != 4 {
		settings.JSONIndent = defaults.JSONIndent
	}
	return settings
}
