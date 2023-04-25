package main

import "time"

type JobType string

type JobLookup struct {
	JobName     string
	clientId    int
	FileSetName string
}

// JobInfo models query results for static values of a job
type JobInfo struct {
	JobLookup
	JobType    JobType
	JobName    string
	ClientName string

	TotalCount int
	TotalBytes int
	TotalFiles int
}

// LastJob models query results for job metrics
type LastJob struct {
	JobStatus    string
	JobBytes     int
	JobFiles     int
	JobErrors    int
	JobStartDate time.Time
	JobEndDate   time.Time
}

// PoolInfo models query result of pool information
type PoolInfo struct {
	Name     string
	Volumes  int
	Bytes    int
	Prunable bool
	Expired  bool
}
