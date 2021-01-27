package main

import (
	"database/sql"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// Connection to database, and database specific queries
type Connection struct {
	db      *sql.DB
	queries *sqlQueries
}

type sqlQueries struct {
	ServerList    string
	JobTotals     string
	LastJob       string
	LastJobStatus string
	LastFullJob   string
	ScheduledJobs string
	PoolInfo      string
}

var queries map[string]*sqlQueries = map[string]*sqlQueries{
	"mysql": &sqlQueries{
		ServerList:    "SELECT DISTINCT Name FROM Job WHERE SchedTime >= ?",
		JobTotals:     "SELECT COUNT(*), SUM(JobBytes), SUM(JobFiles) FROM Job WHERE Name=? AND PurgedFiles=0",
		LastJob:       "SELECT Level,JobBytes,JobFiles,JobErrors,StartTime,EndTime FROM Job WHERE Name = ? AND JobStatus IN('T', 'W') ORDER BY StartTime DESC LIMIT 1",
		LastJobStatus: "SELECT JobStatus FROM Job WHERE Name = ? ORDER BY StartTime DESC LIMIT 1",
		ScheduledJobs: "SELECT COUNT(SchedTime) AS JobsScheduled FROM Job WHERE Name = ? AND SchedTime >= ?",
		PoolInfo:      "SELECT p.name, sum(m.volbytes) AS bytes, count(*) AS volumes, (not exists(select * from JobMedia jm where jm.mediaid = m.mediaid)) AS prunable, TIMESTAMPADD(SECOND, m.volretention, m.lastwritten) < NOW() AS expired FROM Media m LEFT JOIN Pool p ON m.poolid = p.poolid GROUP BY p.name, prunable, expired",
	},
	"postgres": &sqlQueries{
		ServerList:    "SELECT DISTINCT Name FROM job WHERE SchedTime >= $1",
		JobTotals:     "SELECT COUNT(*), SUM(JobBytes), SUM(JobFiles) FROM job WHERE Name=$1 AND PurgedFiles=0",
		LastJob:       "SELECT Level,JobBytes,JobFiles,JobErrors,StartTime,EndTime FROM job WHERE Name = $1 AND JobStatus IN('T', 'W') ORDER BY StartTime DESC LIMIT 1",
		LastJobStatus: "SELECT JobStatus FROM job WHERE Name = $1 ORDER BY StartTime DESC LIMIT 1",
		ScheduledJobs: "SELECT COUNT(SchedTime) AS JobsScheduled FROM job WHERE Name = $1 AND SchedTime >= $2",
		PoolInfo:      "SELECT p.name, sum(m.volbytes) AS bytes, count(m) AS volumes, (not exists(select * from jobmedia jm where jm.mediaid = m.mediaid)) AS prunable, (m.lastwritten + (m.volretention * interval '1s')) < NOW() as expired FROM media m LEFT JOIN pool p ON m.poolid = p.poolid GROUP BY p.name, prunable, expired",
	},
}

// GetConnection opens a new db connection
func GetConnection(databaseType string, connectionString string) (*Connection, error) {
	selectedQueries := queries[databaseType]

	if selectedQueries == nil {
		return nil, fmt.Errorf("Unknown database type %s", databaseType)
	}

	db, err := sql.Open(databaseType, connectionString)

	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &Connection{
		db:      db,
		queries: selectedQueries,
	}, nil
}

// GetServerList reads all servers with scheduled backups for current date
func (connection Connection) GetServerList() ([]string, error) {
	date := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	results, err := connection.execQuery(connection.queries.ServerList, date)

	if err != nil {
		return nil, err
	}

	defer results.Close()

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

func (connection Connection) JobTotals(server string) (*JobTotals, error) {
	results, err := connection.execQuery(connection.queries.JobTotals, server)
	if err != nil {
		return nil, err
	}
	defer results.Close()

	var jobTotals JobTotals
	if results.Next() {
		err = results.Scan(&jobTotals.Count, &jobTotals.Bytes, &jobTotals.Files)
	}

	return &jobTotals, err

}

// LastJob returns metrics for latest executed server backup
func (connection Connection) LastJob(server string) (*LastJob, error) {
	results, err := connection.execQuery(connection.queries.LastJob, server)
	if err != nil {
		return nil, err
	}
	defer results.Close()

	var lastJob LastJob
	if results.Next() {
		err = results.Scan(&lastJob.Level, &lastJob.JobBytes, &lastJob.JobFiles, &lastJob.JobErrors, &lastJob.JobStartDate, &lastJob.JobEndDate)
	}

	return &lastJob, err
}

// LastJobStatus returns metrics for the status of the latest executed server backup
func (connection Connection) LastJobStatus(server string) (*string, error) {
	results, err := connection.execQuery(connection.queries.LastJobStatus, server)
	if err != nil {
		return nil, err
	}
	defer results.Close()

	var jobStatus string
	if results.Next() {
		err = results.Scan(&jobStatus)
	}
	return &jobStatus, err
}

// ScheduledJobs returns amount of scheduled jobs
func (connection Connection) ScheduledJobs(server string) (*ScheduledJob, error) {
	date := time.Now().Format("2006-01-02")
	results, err := connection.execQuery(connection.queries.ScheduledJobs, server, date)
	if err != nil {
		return nil, err
	}
	defer results.Close()

	var schedJob ScheduledJob
	if results.Next() {
		err = results.Scan(&schedJob.ScheduledJobs)
		results.Close()
	}

	return &schedJob, err
}

type poolInfoState struct {
	Prunable bool
	Expired  bool
}

func (connection Connection) PoolInfo() ([]PoolInfo, error) {
	results, err := connection.execQuery(connection.queries.PoolInfo)
	if err != nil {
		return nil, err
	}
	defer results.Close()

	var poolInfoList []PoolInfo

	poolInfoStates := make(map[string][]poolInfoState)

	for results.Next() {
		var poolInfo PoolInfo
		err = results.Scan(&poolInfo.Name, &poolInfo.Bytes, &poolInfo.Volumes, &poolInfo.Prunable, &poolInfo.Expired)
		if err != nil {
			return nil, err
		}
		poolInfoStates[poolInfo.Name] = append(poolInfoStates[poolInfo.Name], poolInfoState{
			Prunable: poolInfo.Prunable,
			Expired:  poolInfo.Expired,
		})

		poolInfoList = append(poolInfoList, poolInfo)
	}

	for poolName, states := range poolInfoStates {
		missingStates := createMissingStates(states)
		for _, missingState := range missingStates {
			poolInfoList = append(poolInfoList, PoolInfo{
				Name:     poolName,
				Prunable: missingState.Prunable,
				Expired:  missingState.Expired,
				Volumes:  0,
				Bytes:    0,
			})
		}
	}

	return poolInfoList, nil
}

var allPoolInfoStates []poolInfoState = []poolInfoState{
	{
		Prunable: false,
		Expired:  false,
	},
	{
		Prunable: true,
		Expired:  false,
	},
	{
		Prunable: false,
		Expired:  true,
	},
	{
		Prunable: true,
		Expired:  true,
	},
}

func createMissingStates(lst []poolInfoState) []poolInfoState {
	var missingStates []poolInfoState
	for _, infoState := range allPoolInfoStates {
		if !hasState(lst, infoState) {
			missingStates = append(missingStates, infoState)
		}
	}
	return missingStates
}

func hasState(lst []poolInfoState, itm poolInfoState) bool {
	for _, i := range lst {
		if i.Prunable == itm.Prunable && i.Expired == itm.Expired {
			return true
		}
	}
	return false
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
