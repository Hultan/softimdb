package imdb

import (
	"path"
	"testing"
)

func Test_NewApiKeyManager(t *testing.T) {
	manager, err := NewApiKeyManager(testApiKey)
	if err != nil {
		t.Errorf("NewApiKeyManager failed with err = %v", err)
	}
	if manager == nil {
		t.Errorf("NewApiKeyManager is nil")
	}
	if manager.GetApiKey() != testApiKey {
		t.Errorf("NewApiKeyManager has the wrong key = %v, want = %v", manager.GetApiKey(), testApiKey)
	}
}

func Test_NewApiKeyManagerFromStandardPath(t *testing.T) {
	manager, err := NewApiKeyManagerFromStandardPath()
	if err != nil {
		t.Errorf("NewApiKeyManagerFromStandardPath failed with err = %v", err)
	}
	if manager == nil {
		t.Errorf("NewApiKeyManagerFromStandardPath is nil")
	}
	if !validateKey(manager.GetApiKey()) {
		t.Errorf("NewApiKeyManagerFromStandardPath got an invalid key = %v", manager.GetApiKey())
	}
}

func Test_NewApiKeyManagerFromPath(t *testing.T) {
	userHome, err := getUserHome()
	if err != nil {
		t.Errorf("getUserHome failed with err = %v", err)
	}
	dir := path.Join(userHome, apiKeyFileName)
	manager, err := NewApiKeyManagerFromPath(dir)
	if err != nil {
		t.Errorf("NewApiKeyManagerFromPath failed with err = %v", err)
	}
	if manager == nil {
		t.Errorf("NewApiKeyManagerFromPath is nil")
	}
	if !validateKey(manager.GetApiKey()) {
		t.Errorf("NewApiKeyManagerFromPath got an invalid key = %v", manager.GetApiKey())
	}
}
func Test_getUserHome(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{"user home", "/home/per", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getUserHome()
			if (err != nil) != tt.wantErr {
				t.Errorf("getUserHome() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getUserHome() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateKey(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"valid key", args{testApiKey}, true},
		{"empty key", args{""}, false},
		{"short key", args{"k_123"}, false},
		{"long key", args{"k_1234567890"}, false},
		{"short key without k_", args{"123"}, false},
		{"long key without k_", args{"12345678901234567890"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateKey(tt.args.key); got != tt.want {
				t.Errorf("validateKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
