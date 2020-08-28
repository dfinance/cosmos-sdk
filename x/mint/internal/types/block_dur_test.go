package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBlockDurFilter(t *testing.T) {
	input := struct {
		Timestamps []time.Time
		Results    []struct {
			Window uint16
			Avg    time.Duration
		}
	}{
		[]time.Time{
			time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),            // 0s
			time.Date(2000, 1, 1, 0, 0, 6, 0, time.UTC),            // 6s
			time.Date(2000, 1, 1, 0, 0, 11, 500*1000000, time.UTC), // 5.5s
			time.Date(2000, 1, 1, 0, 0, 18, 0, time.UTC),           // 6.5s
			time.Date(2000, 1, 1, 0, 0, 25, 0, time.UTC),           // 7s
		},
		[]struct {
			Window uint16
			Avg    time.Duration
		}{
			{
				Window: 2,
				Avg:    7 * time.Second, // [7.0s] the last one
			},
			{
				Window: 3,
				Avg:    6*time.Second + 750*time.Millisecond, // 5.75s, 6.0s, [6.75s]
			},
			{
				Window: 4,
				Avg:    6*time.Second + 333333333*time.Nanosecond, // 6.0s, [5.33s]
			},
			{
				Window: 5,
				Avg:    6*time.Second + 250*time.Millisecond, // [6.25s]
			},
		},
	}

	for testIdx, testResult := range input.Results {
		f := NewBlockDurFilter(testResult.Window)
		_, err := f.GetAvg(testResult.Window)
		require.Error(t, err, "test #%d", testIdx)

		for valueIdx, value := range input.Timestamps {
			f.Push(value, testResult.Window)

			_, err := f.GetAvg(testResult.Window)
			if uint16(valueIdx)+1 < testResult.Window {
				require.Error(t, err, "test #%d valueIdx: %d", testIdx, valueIdx)
			} else {
				require.NoError(t, err, "test #%d valueIdx: %d", testIdx, valueIdx)
			}
		}

		avg, err := f.GetAvg(testResult.Window)
		require.NoError(t, err, "test #%d", testIdx)
		require.Equal(t, testResult.Avg, avg, "test #%d", testIdx)

		blocks, err := f.GetBlocksPerYear(testResult.Window)
		require.NoError(t, err, "test #%d", testIdx)

		t.Logf("test #%d: avgBlockDur / blocksPerYear: %v / %d", testIdx, avg, blocks)
	}
}
