package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShowBuildInfo(t *testing.T) {
	t.Run("prints build info", func(t *testing.T) {
		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		showBuildInfo()

		_ = w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output := buf.String()

		assert.Contains(t, output, "Build version:")
		assert.Contains(t, output, "Build date:")
		assert.Contains(t, output, "Build commit:")
	})
}

