package rest

//nolint:deadcode,unused
type (
	// SearchTxsResult copy of sdk.SearchTxsResult
	SearchTxsResult []struct {
		TotalCount int `json:"total_count"` // Count of all txs
		Count      int `json:"count"`       // Count of txs in current page
		PageNumber int `json:"page_number"` // Index of current page, start from 1
		PageTotal  int `json:"page_total"`  // Count of total pages
		Limit      int `json:"limit"`       // Max count txs per page
		Txs        []struct {
			Height    int64       `json:"height"`
			TxHash    string      `json:"txhash"`
			Codespace string      `json:"codespace,omitempty"`
			Code      uint32      `json:"code,omitempty"`
			Data      string      `json:"data,omitempty"`
			RawLog    string      `json:"raw_log,omitempty"`
			Logs      interface{} `json:"logs,omitempty"`
			Info      string      `json:"info,omitempty"`
			GasWanted int64       `json:"gas_wanted,omitempty"`
			GasUsed   int64       `json:"gas_used,omitempty"`
			Tx        interface{} `json:"tx,omitempty"`
			Timestamp string      `json:"timestamp,omitempty"`
		} `json:"txs"` // List of txs in current page
	}
)
