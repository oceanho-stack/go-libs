package xcfg

import (
	"os"
	"path/filepath"
	"testing"
)

type testConfig struct {
	App struct {
		Name string `mapstructure:"name" validate:"required"`
		Port int    `mapstructure:"port"`
	} `mapstructure:"app"`
	DB struct {
		DSN string `mapstructure:"dsn"`
	} `mapstructure:"db"`
}

func TestLoadWithDefaults(t *testing.T) {
	var cfg testConfig
	err := Load(Options{
		Defaults: map[string]any{
			"app.name": "iam-server",
			"app.port": 8080,
		},
		Target: &cfg,
	})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.App.Name != "iam-server" {
		t.Errorf("app.name = %q, want iam-server", cfg.App.Name)
	}
	if cfg.App.Port != 8080 {
		t.Errorf("app.port = %d, want 8080", cfg.App.Port)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := "app:\n  name: my-app\n  port: 9090\ndb:\n  dsn: postgres://localhost:5432/x\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	var cfg testConfig
	err := Load(Options{
		File:   path,
		Target: &cfg,
	})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.App.Name != "my-app" {
		t.Errorf("app.name = %q, want my-app", cfg.App.Name)
	}
	if cfg.App.Port != 9090 {
		t.Errorf("app.port = %d, want 9090", cfg.App.Port)
	}
	if cfg.DB.DSN != "postgres://localhost:5432/x" {
		t.Errorf("db.dsn = %q", cfg.DB.DSN)
	}
}

func TestLoadEnvOverride(t *testing.T) {
	t.Setenv("TEST_APP_NAME", "env-app")
	t.Setenv("TEST_APP_PORT", "7070")

	var cfg testConfig
	err := Load(Options{
		EnvPrefix: "TEST",
		Defaults: map[string]any{
			"app.name": "default",
			"app.port": 8080,
		},
		Target: &cfg,
	})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.App.Name != "env-app" {
		t.Errorf("app.name = %q, want env-app", cfg.App.Name)
	}
	if cfg.App.Port != 7070 {
		t.Errorf("app.port = %d, want 7070", cfg.App.Port)
	}
}

func TestLoadValidation(t *testing.T) {
	var cfg testConfig
	err := Load(Options{
		Defaults: map[string]any{
			"app.port": 8080, // 缺 app.name，校验应失败
		},
		Target: &cfg,
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}
