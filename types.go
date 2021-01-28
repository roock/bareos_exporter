package main

import "time"

type JobType string

const (
	BackupJob      JobType = "B"
	MigratedJob            = "J"
	VerifyJob              = "V"
	RestoreJob             = "R"
	ConsoleJob             = "U"
	SystemJob              = "I"
	AdminJob               = "D"
	ArchiveJob             = "A"
	JobCopyJob             = "C"
	CopyJob                = "c"
	MigrateJob             = "J"
	ScanJob                = "S"
	ConsolidateJob         = "O"
)

type JobLookup struct {
	JobName   string
	clientId  int
	fileSetId int
}

// JobInfo models query results for static values of a job
type JobInfo struct {
	JobLookup
	JobType     JobType
	JobName     string
	ClientName  string
	FileSetName string

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
