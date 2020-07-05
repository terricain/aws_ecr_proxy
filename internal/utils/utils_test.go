package utils

import (
	"github.com/rs/zerolog"
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	if err := os.Setenv("KNOWN_ENV_KEY", "test1234"); err != nil {
		t.Fatalf("Failed to set env: %s", err.Error())
	}

	if result := GetEnv("NONEXISTENT_ENV_KEY", "default"); result != "default" {
		t.Errorf("LogNameToLevel(\"NONEXISTENT_ENV_KEY\") was incorrect, got: %v, want: \"default\"", result)
	}

	if result := GetEnv("KNOWN_ENV_KEY", "default"); result != "test1234" {
		t.Errorf("LogNameToLevel(\"KNOWN_ENV_KEY\") was incorrect, got: %v, want: \"test1234\"", result)
	}
}

func TestLogNameToLevel(t *testing.T) {
	tables := []struct {
		logName  string
		logLevel zerolog.Level
	}{
		{"INFO", zerolog.InfoLevel},
		{"DEBUG", zerolog.DebugLevel},
		{"WARN", zerolog.WarnLevel},
		{"ERROR", zerolog.ErrorLevel},
		{"LALALALA", zerolog.InfoLevel},
	}

	for _, table := range tables {
		expectedLevel := LogNameToLevel(table.logName)
		if expectedLevel != table.logLevel {
			t.Errorf("LogNameToLevel(%s) was incorrect, got: %v, want: %v", table.logName, expectedLevel, table.logLevel)
		}
	}
}
