package main

import "time"

// LastJob models query results for job metrics
type LastJob struct {
	Level        string    `json:"level"`
	JobBytes     int       `json:"job-bytes"`
	JobFiles     int       `json:"job-files"`
	JobErrors    int       `json:"job-errors"`
	JobStartDate time.Time `json:"job-start-date"`
	JobEndDate   time.Time `json:"job-end-date"`
}

// PoolInfo models query result of pool information
type PoolInfo struct {
	Name     string `json:"name"`
	Volumes  int    `json:"volumes"`
	Bytes    int    `json:"files"`
	Prunable bool   `json:"prunable"`
	Expired  bool   `json:"expired"`
}

// TotalBytes models query result of saved bytes sum for a server
type TotalBytes struct {
	Bytes int `json:"files"`
}

// TotalFiles models query result of saved files sum for a server
type TotalFiles struct {
	Files int `json:"files"`
}

// ScheduledJob models query result of the time a job is about to be executed
type ScheduledJob struct {
	ScheduledJobs int `json:"scheduled-jobs"`
}
