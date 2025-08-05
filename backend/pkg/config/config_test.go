package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigWithDefaults(t *testing.T) {
	// Unset all relevant environment variables to ensure a clean test
	unsetEnvVars()

	config, err := LoadConfig()

	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Assert default values
	assert.Equal(t, "localhost", config.ServerHost)
	assert.Equal(t, 8888, config.ServerPort)
	assert.Equal(t, "development", config.NodeEnv)
}

func TestLoadConfigWithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("ZHULONG_SERVER_HOST", "testhost")
	os.Setenv("ZHULONG_SERVER_PORT", "9999")
	os.Setenv("NODE_ENV", "production")

	config, err := LoadConfig()

	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Assert values from environment variables
	assert.Equal(t, "testhost", config.ServerHost)
	assert.Equal(t, 9999, config.ServerPort)
	assert.Equal(t, "production", config.NodeEnv)

	// Unset environment variables after the test
	unsetEnvVars()
}

func TestLoadConfigWithEnvFile(t *testing.T) {
	// Create a temporary .env file
	content := []byte("ZHULONG_SERVER_HOST=filehost\nZHULONG_SERVER_PORT=1111")
	tmpfile, err := os.Create(".env")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err)
	assert.NoError(t, tmpfile.Close())

	config, err := LoadConfig()

	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Assert values from .env file
	assert.Equal(t, "filehost", config.ServerHost)
	assert.Equal(t, 1111, config.ServerPort)
}

// unsetEnvVars unsets the environment variables used in the tests.
func unsetEnvVars() {
	os.Unsetenv("ZHULONG_SERVER_HOST")
	os.Unsetenv("ZHULONG_SERVER_PORT")
	os.Unsetenv("NODE_ENV")
}
