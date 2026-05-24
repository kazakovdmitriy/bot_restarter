package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	validCfg := Config{
		TelegramToken:  "test-token",
		AllowedUserIDs: []int64{123, 456},
	}

	tests := []struct {
		name    string
		content []byte
		want    *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			content: mustMarshal(validCfg),
			want:    &validCfg,
			wantErr: false,
		},
		{
			name:    "empty allowed_user_ids",
			content: mustMarshal(Config{TelegramToken: "token"}),
			want:    &Config{TelegramToken: "token", AllowedUserIDs: nil},
			wantErr: false,
		},
		{
			name:    "missing token",
			content: mustMarshal(Config{AllowedUserIDs: []int64{1}}),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "missing file",
			content: nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid json",
			content: []byte(`{bad json`),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			if tt.content != nil {
				dir := t.TempDir()
				path = filepath.Join(dir, "config.json")
				if err := os.WriteFile(path, tt.content, 0o644); err != nil {
					t.Fatalf("write test config: %v", err)
				}
			} else {
				path = filepath.Join(t.TempDir(), "nonexistent.json")
			}

			got, err := LoadConfig(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !cfgEqual(got, tt.want) {
				t.Errorf("LoadConfig() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func mustMarshal(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

func cfgEqual(a, b *Config) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.TelegramToken != b.TelegramToken {
		return false
	}
	if len(a.AllowedUserIDs) != len(b.AllowedUserIDs) {
		return false
	}
	for i := range a.AllowedUserIDs {
		if a.AllowedUserIDs[i] != b.AllowedUserIDs[i] {
			return false
		}
	}
	return true
}
