package types

import (
	"fmt"
	"time"
)

const (
	AvgYearDur = time.Duration(60*60*8766) * time.Second // 365.25 days
)

// BlockDurFilter keeps blockTime values up to params.AvgBlockTimeWindow.
type BlockDurFilter struct {
	Values []int64 `json:"values"`
}

// Push pushes a new value back and trims slice if needed.
func (f *BlockDurFilter) Push(blockTime time.Time, window uint16) {
	value := blockTime.UnixNano()
	f.Values = append(f.Values, value)

	fWindow, fLen := int(window), len(f.Values)
	if fLen > fWindow {
		f.Values = f.Values[fLen-fWindow:]
	}
}

// GetAvg returns moving average value of blockTime duration if filter window is full.
func (f *BlockDurFilter) GetAvg(window uint16) (time.Duration, error) {
	fWindow, fLen := int(window), len(f.Values)

	if fLen < 2 {
		return 0, fmt.Errorf("filter length is LT 2: %d", fLen)
	}
	if fLen != fWindow {
		return 0, fmt.Errorf("filter length is LT window size: %d < %d", fLen, fWindow)
	}

	res := time.Duration(0)
	for i := 1; i < fLen; i++ {
		diff := f.Values[i] - f.Values[i-1]
		res += time.Duration(diff) * time.Nanosecond
	}
	res /= time.Duration(window - 1)

	return res, nil
}

// GetBlocksPerYear returns average blockPerYear counter if filter window is full.
func (f *BlockDurFilter) GetBlocksPerYear(window uint16) (uint64, error) {
	avgDur, err := f.GetAvg(window)
	if err != nil {
		return 0, err
	}

	// sanity check
	// this is relevant for tests as they might not increase Header time
	if avgDur == 0 {
		return 0, fmt.Errorf("average duration is zero")
	}

	return uint64(AvgYearDur / avgDur), nil
}

// NewBlockDurFilter creates a new BlockDurFilter object.
func NewBlockDurFilter(window uint16) BlockDurFilter {
	return BlockDurFilter{
		Values: make([]int64, 0, window),
	}
}
