package ansimate_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bengarrett/ansimate"
	"github.com/nalgeon/be"
)

func TestGoldenWebP(t *testing.T) {
	t.Parallel()
	const (
		maxStandard = 15 * time.Second
		maxExtended = 25 * time.Second
	)

	tests := []struct {
		name        string
		maxDuration time.Duration
	}{
		{name: "CHAOS2", maxDuration: maxStandard},
		{name: "MAXIMUM3", maxDuration: maxStandard},
		{name: "SC-N0393", maxDuration: maxStandard},
		{name: "SSACID4", maxDuration: maxExtended},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ansPath := filepath.Join("testdata", tc.name+".ANS")
			goldenPath := filepath.Join("testdata", tc.name+".webp")

			ansData, err := os.ReadFile(ansPath)
			if err != nil {
				t.Skip("testdata not found:", err)
			}

			goldenBytes, err := os.ReadFile(goldenPath)
			be.Err(t, err, nil)

			// 1. Verify ANSImation detection
			isAnim := ansimate.DetectBytes(ansData)
			be.True(t, isAnim)

			// 2. Generate WebP into buffer
			var gotBuf bytes.Buffer
			cfg := &ansimate.Config{
				MaxDuration: tc.maxDuration,
			}
			err = ansimate.Animate(&gotBuf, bytes.NewReader(ansData), cfg)
			be.Err(t, err, nil)

			// 3. Verify exact byte-for-byte bitwise match against golden file
			be.Equal(t, gotBuf.Len(), len(goldenBytes))
			be.True(t, bytes.Equal(gotBuf.Bytes(), goldenBytes))
		})
	}
}
