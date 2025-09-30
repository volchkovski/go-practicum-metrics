package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestConstants(t *testing.T) {
	assert.Equal(t, "prod", ProductionEnv)
	assert.Equal(t, "local", LocalEnv)
}

func TestInitialize_ProductionEnvironment(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error", "dpanic", "panic", "fatal"}

	for _, level := range levels {
		t.Run("production_"+level, func(t *testing.T) {
			err := Initialize(level, ProductionEnv)
			require.NoError(t, err)

			assert.NotNil(t, Log)

			assert.NotEqual(t, zap.NewNop().Sugar(), Log)
		})
	}
}

func TestInitialize_LocalEnvironment(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error", "dpanic", "panic", "fatal"}

	for _, level := range levels {
		t.Run("local_"+level, func(t *testing.T) {
			err := Initialize(level, LocalEnv)
			require.NoError(t, err)

			assert.NotNil(t, Log)

			assert.NotEqual(t, zap.NewNop().Sugar(), Log)
		})
	}
}

func TestInitialize_InvalidLogLevel(t *testing.T) {
	invalidLevels := []string{"invalid", "trace", "verbose", "critical"}

	for _, level := range invalidLevels {
		t.Run("invalid_level_"+level, func(t *testing.T) {
			err := Initialize(level, ProductionEnv)
			assert.Error(t, err)
		})
	}

	t.Run("empty_level", func(t *testing.T) {
		err := Initialize("", ProductionEnv)
		_ = err
	})
}

func TestInitialize_InvalidEnvironment(t *testing.T) {
	invalidEnvs := []string{"development", "test", "staging", "invalid", ""}

	for _, env := range invalidEnvs {
		t.Run("invalid_env_"+env, func(t *testing.T) {
			err := Initialize("info", env)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "logger valid environment values")
		})
	}
}

func TestInitialize_DefaultLogger(t *testing.T) {
	originalLog := Log

	Log = zap.NewNop().Sugar()

	err := Initialize("info", ProductionEnv)
	require.NoError(t, err)

	assert.NotNil(t, Log)
	assert.NotEqual(t, zap.NewNop().Sugar(), Log)

	Log = originalLog
}

func TestInitialize_LoggerFunctionality(t *testing.T) {
	err := Initialize("debug", LocalEnv)
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		Log.Debug("debug message")
		Log.Info("info message")
		Log.Warn("warn message")
		Log.Error("error message")
	})

	assert.NotPanics(t, func() {
		Log.Infow("structured message", "key", "value", "number", 42)
		Log.Errorw("structured error", "error", "test error", "code", 500)
	})

	assert.NotPanics(t, func() {
		Log.Infof("formatted message: %s %d", "test", 123)
		Log.Errorf("formatted error: %v", "test error")
	})
}

func TestInitialize_LogLevel_Filtering(t *testing.T) {
	// This test is more complex as it requires capturing log output
	// For now, we just test that initialization works with different levels

	testCases := []struct {
		level string
		env   string
	}{
		{"debug", LocalEnv},
		{"info", LocalEnv},
		{"warn", LocalEnv},
		{"error", LocalEnv},
		{"debug", ProductionEnv},
		{"info", ProductionEnv},
		{"warn", ProductionEnv},
		{"error", ProductionEnv},
	}

	for _, tc := range testCases {
		t.Run(tc.level+"_"+tc.env, func(t *testing.T) {
			err := Initialize(tc.level, tc.env)
			require.NoError(t, err)

			// Verify logger is functional
			assert.NotNil(t, Log)
		})
	}
}

func TestInitialize_CaseInsensitiveLevel(t *testing.T) {
	// Test that log levels are case insensitive
	levels := []string{"INFO", "Info", "iNfO", "DEBUG", "Debug", "ERROR", "Error"}

	for _, level := range levels {
		t.Run("case_"+level, func(t *testing.T) {
			err := Initialize(level, ProductionEnv)
			require.NoError(t, err)
			assert.NotNil(t, Log)
		})
	}
}

func TestInitialize_Concurrent(t *testing.T) {
	const numGoroutines = 10

	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			env := ProductionEnv
			if id%2 == 0 {
				env = LocalEnv
			}

			err := Initialize("info", env)
			results <- err
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(t, err)
	}

	// Verify final logger is valid
	assert.NotNil(t, Log)
}

func TestInitialize_OverwriteExisting(t *testing.T) {
	err := Initialize("info", ProductionEnv)
	require.NoError(t, err)
	firstLogger := Log

	err = Initialize("debug", LocalEnv)
	require.NoError(t, err)
	secondLogger := Log

	assert.NotEqual(t, firstLogger, secondLogger)

	assert.NotNil(t, firstLogger)
	assert.NotNil(t, secondLogger)
}

func TestLogLevels_Zap(t *testing.T) {
	validLevels := []string{
		"debug", "info", "warn", "error", "dpanic", "panic", "fatal",
	}

	for _, level := range validLevels {
		t.Run("zap_level_"+level, func(t *testing.T) {
			_, err := zap.ParseAtomicLevel(level)
			assert.NoError(t, err, "Expected zap to parse level: %s", level)
		})
	}
}

func TestLogLevel_Values(t *testing.T) {
	testCases := []struct {
		level         string
		expectedLevel zapcore.Level
	}{
		{"debug", zapcore.DebugLevel},
		{"info", zapcore.InfoLevel},
		{"warn", zapcore.WarnLevel},
		{"error", zapcore.ErrorLevel},
	}

	for _, tc := range testCases {
		t.Run(tc.level, func(t *testing.T) {
			atomicLevel, err := zap.ParseAtomicLevel(tc.level)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedLevel, atomicLevel.Level())
		})
	}
}
