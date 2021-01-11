package dataaccess

import (
	"database/sql"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vierbergenlars/bareos_exporter/types"
)

// Connection to database, and database specific queries
type Connection struct {
	db      *sql.DB
	queries *sqlQueries
}

type sqlQueries struct {
	ServerList    string
	TotalBytes    string
	TotalFiles    string
	LastJob       string
	LastJobStatus string
	LastFullJob   string
	ScheduledJobs string
	PoolInfo      string
}

var mysqlQueries *sqlQueries = &sqlQueries{
	ServerList:    "SELECT DISTINCT Name FROM Job WHERE SchedTime >= ?",
	TotalBytes:    "SELECT SUM(JobBytes) FROM Job WHERE Name=? AND PurgedFiles=0 AND JobStatus IN('T', 'W')",
	TotalFiles:    "SELECT SUM(JobFiles) FROM Job WHERE Name=? AND PurgedFiles=0 AND JobStatus IN('T', 'W')",
	LastJob:       "SELECT Level,JobBytes,JobFiles,JobErrors,StartTime FROM Job WHERE Name = ? AND JobStatus IN('T', 'W') ORDER BY StartTime DESC LIMIT 1",
	LastJobStatus: "SELECT JobStatus FROM Job WHERE Name = ? ORDER BY StartTime DESC LIMIT 1",
	LastFullJob:   "SELECT Level,JobBytes,JobFiles,JobErrors,StartTime FROM Job WHERE Name = ? AND Level = 'F' AND JobStatus IN('T', 'W') ORDER BY StartTime DESC LIMIT 1",
	ScheduledJobs: "SELECT COUNT(SchedTime) AS JobsScheduled FROM Job WHERE Name = ? AND SchedTime >= ?",
	PoolInfo:      "SELECT p.name, sum(m.volbytes) as bytes, COUNT(m) as volumes, 0 as prunable FROM Media m LEFT JOIN Pool p ON m.poolid = p.poolid GROUP BY p.name",
}

var postgresQueries *sqlQueries = &sqlQueries{
	ServerList:    "SELECT DISTINCT Name FROM job WHERE SchedTime >= $1",
	TotalBytes:    "SELECT SUM(JobBytes) FROM job WHERE Name=$1 AND PurgedFiles=0 AND JobStatus IN('T', 'W')",
	TotalFiles:    "SELECT SUM(JobFiles) FROM job WHERE Name=$1 AND PurgedFiles=0 AND JobStatus IN('T', 'W')",
	LastJob:       "SELECT Level,JobBytes,JobFiles,JobErrors,StartTime FROM job WHERE Name = $1 AND JobStatus IN('T', 'W') ORDER BY StartTime DESC LIMIT 1",
	LastJobStatus: "SELECT JobStatus FROM job WHERE Name = $1 ORDER BY StartTime DESC LIMIT 1",
	LastFullJob:   "SELECT Level,JobBytes,JobFiles,JobErrors,StartTime FROM job WHERE Name = $1 AND Level = 'F' AND JobStatus IN('T', 'W') ORDER BY StartTime DESC LIMIT 1",
	ScheduledJobs: "SELECT COUNT(SchedTime) AS JobsScheduled FROM job WHERE Name = $1 AND SchedTime >= $2",
	PoolInfo:      "SELECT p.name, sum(m.volbytes) AS bytes, count(m) AS volumes, (not exists(select * from jobmedia jm where jm.mediaid = m.mediaid)) AS prunable FROM media m LEFT JOIN pool p ON m.poolid = p.poolid GROUP BY p.name, prunable",
}

// GetConnection opens a new db connection
func GetConnection(databaseType string, connectionString string) (*Connection, error) {
	var queries *sqlQueries
	switch databaseType {
	case "mysql":
		queries = mysqlQueries
	case "postgres":
		queries = postgresQueries
	default:
		return nil, fmt.Errorf("Unknown database type %s", databaseType)
	}

	db, err := sql.Open(databaseType, connectionString)

	if err != nil {
		return nil, err
	}

	return &Connection{
		db:      db,
		queries: queries,
	}, nil
}

// GetServerList reads all servers with scheduled backups for current date
func (connection Connection) GetServerList() ([]string, error) {
	date := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	results, err := connection.execQuery(connection.queries.ServerList, date)
	defer results.Close()

	if err != nil {
		return nil, err
	}

	var servers []string

	for results.Next() {
		var server string
		err = results.Scan(&server)
		if err != nil {
			return nil, err
		}
		servers = append(servers, server)
	}

	return servers, err
}

func (connection Connection) execQuery(query string, args ...interface{}) (*sql.Rows, error) {
	results, err := connection.db.Query(query, args...)
	if err != nil {
		log.WithFields(log.Fields{
			"query": query,
			"args":  args,
		}).Error(err)
	}
	return results, err
}

// TotalBytes returns total bytes saved for a server since the very first backup
func (connection Connection) TotalBytes(server string) (*types.TotalBytes, error) {
	results, err := connection.execQuery(connection.queries.TotalBytes, server)
	defer results.Close()

	if err != nil {
		return nil, err
	}

	var totalBytes types.TotalBytes
	if results.Next() {
		err = results.Scan(&totalBytes.Bytes)
	}

	return &totalBytes, err
}

// TotalFiles returns total files saved for a server since the very first backup
func (connection Connection) TotalFiles(server string) (*types.TotalFiles, error) {
	results, err := connection.execQuery(connection.queries.TotalFiles, server)
	defer results.Close()

	if err != nil {
		return nil, err
	}

	var totalFiles types.TotalFiles
	if results.Next() {
		err = results.Scan(&totalFiles.Files)
	}

	return &totalFiles, err
}

// LastJob returns metrics for latest executed server backup
func (connection Connection) LastJob(server string) (*types.LastJob, error) {
	results, err := connection.execQuery(connection.queries.LastJob, server)
	defer results.Close()

	if err != nil {
		return nil, err
	}

	var lastJob types.LastJob
	if results.Next() {
		err = results.Scan(&lastJob.Level, &lastJob.JobBytes, &lastJob.JobFiles, &lastJob.JobErrors, &lastJob.JobDate)
	}

	return &lastJob, err
}

// LastJobStatus returns metrics for the status of the latest executed server backup
func (connection Connection) LastJobStatus(server string) (*string, error) {
	results, err := connection.execQuery(connection.queries.LastJobStatus, server)
	defer results.Close()

	if err != nil {
		return nil, err
	}

	var jobStatus string
	if results.Next() {
		err = results.Scan(&jobStatus)
	}
	return &jobStatus, err
}

// LastFullJob returns metrics for latest executed server backup with Level F
func (connection Connection) LastFullJob(server string) (*types.LastJob, error) {
	results, err := connection.execQuery(connection.queries.LastFullJob, server)
	defer results.Close()

	if err != nil {
		return nil, err
	}

	var lastJob types.LastJob
	if results.Next() {
		err = results.Scan(&lastJob.Level, &lastJob.JobBytes, &lastJob.JobFiles, &lastJob.JobErrors, &lastJob.JobDate)
	}

	return &lastJob, err
}

// ScheduledJobs returns amount of scheduled jobs
func (connection Connection) ScheduledJobs(server string) (*types.ScheduledJob, error) {
	date := time.Now().Format("2006-01-02")
	results, err := connection.execQuery(connection.queries.ScheduledJobs, server, date)
	defer results.Close()

	if err != nil {
		return nil, err
	}

	var schedJob types.ScheduledJob
	if results.Next() {
		err = results.Scan(&schedJob.ScheduledJobs)
		results.Close()
	}

	return &schedJob, err
}

func (connection Connection) PoolInfo() ([]types.PoolInfo, error) {
	results, err := connection.execQuery(connection.queries.PoolInfo)
	defer results.Close()

	if err != nil {
		return nil, err
	}

	var poolInfoList []types.PoolInfo

	var withPrunable []string
	var withoutPrunable []string

	for results.Next() {
		var poolInfo types.PoolInfo
		err = results.Scan(&poolInfo.Name, &poolInfo.Bytes, &poolInfo.Volumes, &poolInfo.Prunable)
		if err != nil {
			return nil, err
		}
		if poolInfo.Prunable {
			withPrunable = append(withPrunable, poolInfo.Name)
		} else {
			withoutPrunable = append(withoutPrunable, poolInfo.Name)
		}

		poolInfoList = append(poolInfoList, poolInfo)
	}

	for _, poolInfo := range poolInfoList {
		if !hasItem(withPrunable, poolInfo.Name) {
			poolInfoList = append(poolInfoList, types.PoolInfo{
				Name:     poolInfo.Name,
				Prunable: true,
				Volumes:  0,
				Bytes:    0,
			})
		}

		if !hasItem(withoutPrunable, poolInfo.Name) {
			poolInfoList = append(poolInfoList, types.PoolInfo{
				Name:     poolInfo.Name,
				Prunable: false,
				Volumes:  0,
				Bytes:    0,
			})
		}
	}

	return poolInfoList, nil

}

func hasItem(lst []string, itm string) bool {
	for _, i := range lst {
		if i == itm {
			return true
		}
	}
	return false
}

// Close the database connection
func (connection Connection) Close() error {
	return connection.db.Close()
}
