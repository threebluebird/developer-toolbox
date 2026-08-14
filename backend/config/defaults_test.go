package config

import (
	"testing"

	"developer-toolbox/backend/models"
)

func TestNormalizeSettingsRepairsUnsupportedValues(t *testing.T) {
	settings := NormalizeSettings(models.Settings{Theme: "blue", Language: "fr", JSONIndent: 8})
	defaults := DefaultSettings()
	if settings.Theme != defaults.Theme || settings.Language != defaults.Language || settings.JSONIndent != defaults.JSONIndent {
		t.Fatalf("settings were not normalized: %#v", settings)
	}
}

func TestNormalizeSettingsKeepsSupportedValues(t *testing.T) {
	settings := NormalizeSettings(models.Settings{Theme: "system", Language: "zh", JSONIndent: 4})
	if settings.Theme != "system" || settings.Language != "zh" || settings.JSONIndent != 4 {
		t.Fatalf("valid settings changed: %#v", settings)
	}
}
