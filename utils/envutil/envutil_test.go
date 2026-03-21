package envutil

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	const defaultValue = "default_val"
	tests := []struct {
		name     string
		key      string
		envValue string
		setEnv   bool
		want     string
	}{
		{"Environment variable exists", "TEST_KEY_EXIST", "real_value", true, "real_value"},
		{"Environment variable is empty", "TEST_KEY_EMPTY", "", true, ""},
		{"Environment variable does not exist", "TEST_KEY_NONE", "", false, defaultValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.key, tt.envValue)
			}
			if got := GetEnv(tt.key, defaultValue); got != tt.want {
				t.Errorf("GetEnv(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestGetEnvAsBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		setEnv       bool
		defaultValue bool
		want         bool
	}{
		{"Exists and true", "BOOL_TRUE", "true", true, false, true},
		{"Exists and 1 as true", "BOOL_ONE", "1", true, false, true},
		{"Exists and false", "BOOL_FALSE", "false", true, true, false},
		{"Exists and 0 as false", "BOOL_ZERO", "0", true, true, false},
		{"Invalid value returns default", "BOOL_INVALID", "invalid", true, true, true},
		{"Not exists returns default", "BOOL_NONE", "", false, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.key, tt.envValue)
			} else {
				os.Unsetenv(tt.key)
			}
			if got := GetEnvAsBool(tt.key, tt.defaultValue); got != tt.want {
				t.Errorf("GetEnvAsBool(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		setEnv       bool
		defaultValue int
		want         int
	}{
		{"Exists and valid", "INT_VALID", "123", true, 0, 123},
		{"Exists and negative", "INT_NEG", "-5", true, 0, -5},
		{"Exists and invalid returns default", "INT_INVALID", "abc", true, 42, 42},
		{"Not exists returns default", "INT_NONE", "", false, 99, 99},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.key, tt.envValue)
			} else {
				os.Unsetenv(tt.key)
			}
			if got := GetEnvAsInt(tt.key, tt.defaultValue); got != tt.want {
				t.Errorf("GetEnvAsInt(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}
