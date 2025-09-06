package main

import (
	"flag"
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue string
		expected     string
	}{
		{
			name:         "environment variable set",
			key:          "TEST_VAR",
			value:        "test_value",
			defaultValue: "default",
			expected:     "test_value",
		},
		{
			name:         "environment variable not set",
			key:          "NONEXISTENT_VAR",
			value:        "",
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnv(%q, %q) = %q, want %q", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestCommandLineFlags(t *testing.T) {
	// Test default port value
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test default port
	os.Args = []string{"staticsend"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	port := flag.String("port", getEnv("STATICSEND_PORT", "8080"), "Port to listen on")
	help := flag.Bool("help", false, "Show help")
	flag.Parse()

	if *port != "8080" {
		t.Errorf("Expected default port 8080, got %s", *port)
	}
	if *help != false {
		t.Error("Expected help flag to be false by default")
	}

	// Test custom port flag
	os.Args = []string{"staticsend", "-port=3000"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	port = flag.String("port", getEnv("STATICSEND_PORT", "8080"), "Port to listen on")
	help = flag.Bool("help", false, "Show help")
	flag.Parse()

	if *port != "3000" {
		t.Errorf("Expected port 3000, got %s", *port)
	}

	// Test help flag
	os.Args = []string{"staticsend", "-help"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	help = flag.Bool("help", false, "Show help")
	flag.Parse()

	if *help != true {
		t.Error("Expected help flag to be true when set")
	}
}

func TestEnvironmentVariablePrecedence(t *testing.T) {
	// Set environment variable
	os.Setenv("STATICSEND_PORT", "9000")
	defer os.Unsetenv("STATICSEND_PORT")

	// Test that environment variable takes precedence over default
	result := getEnv("STATICSEND_PORT", "8080")
	if result != "9000" {
		t.Errorf("Expected environment variable value 9000, got %s", result)
	}
}
