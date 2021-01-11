package types

// PoolInfo models query result of pool information
type PoolInfo struct {
	Name     string `json:"name"`
	Volumes  int    `json:"volumes"`
	Bytes    int    `json:"files"`
	Prunable bool   `json:"prunable"`
}
