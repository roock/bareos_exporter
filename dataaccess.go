package main

import (
	"database/sql"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// Connection to database, and database specific queries
type Connection struct {
	db               *sql.DB
	queries          *sqlQueries
	jobDiscoveryDays int
}

type sqlQueries struct {
	JobList               string
	LastJob               string
	LastSuccessfulJob     string
	LastSuccessfulFullJob string
	PoolInfo              string
	JobStates             string
}

var queries map[string]*sqlQueries = map[string]*sqlQueries{
	"mysql": &sqlQueries{
		JobList:               "SELECT j.Name, j.Type, j.ClientId, COALESCE(c.Name, ''), COALESCE(f.FileSet, ''), COUNT(*), SUM(j.JobBytes), SUM(j.JobFiles) FROM Job j LEFT JOIN Client c ON c.ClientId = j.ClientId LEFT JOIN FileSet f ON f.FileSetId = j.FileSetId GROUP BY j.Name, j.Type, j.ClientId, c.Name, f.FileSet HAVING MAX(j.SchedTime) >= ?",
		LastJob:               "SELECT JobStatus,JobBytes,JobFiles,JobErrors,StartTime,COALESCE(EndTime, NOW()) FROM Job WHERE Name = ? AND ClientId = ? AND FileSetId IN(SELECT f.FileSetId FROM FileSet f WHERE f.FileSet = ?) ORDER BY StartTime DESC LIMIT 1",
		LastSuccessfulJob:     "SELECT JobStatus,JobBytes,JobFiles,JobErrors,StartTime,COALESCE(EndTime, NOW()) FROM Job WHERE Name = ? AND ClientId = ? AND FileSetId IN(SELECT f.FileSetId FROM FileSet f WHERE f.FileSet = ?) AND JobStatus IN('T', 'W') ORDER BY StartTime DESC LIMIT 1",
		LastSuccessfulFullJob: "SELECT JobStatus,JobBytes,JobFiles,JobErrors,StartTime,COALESCE(EndTime, NOW()) FROM Job WHERE Name = ? AND ClientId = ? AND FileSetId IN(SELECT f.FileSetId FROM FileSet f WHERE f.FileSet = ?) AND JobStatus IN('T', 'W') AND Level = 'F' ORDER BY StartTime DESC LIMIT 1",
		PoolInfo:              "SELECT p.name, sum(m.volbytes) AS bytes, count(*) AS volumes, (not exists(select * from JobMedia jm where jm.mediaid = m.mediaid)) AS prunable, COALESCE(TIMESTAMPADD(SECOND, m.volretention, m.lastwritten) < NOW(), false) AS expired FROM Media m LEFT JOIN Pool p ON m.poolid = p.poolid GROUP BY p.name, prunable, expired",
		JobStates:             "SELECT JobStatus FROM Status",
	},
	"postgres": &sqlQueries{
		JobList:               "SELECT j.Name, j.Type, j.ClientId, COALESCE(c.Name, ''), COALESCE(f.FileSet, ''), COUNT(*), SUM(j.JobBytes), SUM(j.JobFiles) FROM job j LEFT JOIN client c ON c.ClientId = j.ClientId LEFT JOIN fileset f ON f.FileSetId = j.FileSetId  GROUP BY j.Name, j.Type, j.ClientId, c.Name, f.FileSet HAVING MAX(j.SchedTime) >= $1",
		LastJob:               "SELECT JobStatus,JobBytes,JobFiles,JobErrors,StartTime::timestamptz,COALESCE(EndTime::timestamptz, NOW()) FROM job WHERE Name = $1 AND ClientId = $2 AND FileSetId IN(SELECT f.FileSetId from FileSet f WHERE f.FileSet = $3) ORDER BY StartTime DESC LIMIT 1",
		LastSuccessfulJob:     "SELECT JobStatus,JobBytes,JobFiles,JobErrors,StartTime::timestamptz,COALESCE(EndTime::timestamptz, NOW()) FROM job WHERE Name = $1 AND ClientId = $2 AND FileSetId IN(SELECT f.FileSetId from FileSet f WHERE f.FileSet = $3) AND JobStatus IN('T', 'W') ORDER BY StartTime DESC LIMIT 1",
		LastSuccessfulFullJob: "SELECT JobStatus,JobBytes,JobFiles,JobErrors,StartTime::timestamptz,COALESCE(EndTime::timestamptz, NOW()) FROM job WHERE Name = $1 AND ClientId = $2 AND FileSetId IN(SELECT f.FileSetId from FileSet f WHERE f.FileSet = $3) AND JobStatus IN('T', 'W') AND Level = 'F' ORDER BY StartTime DESC LIMIT 1",
		PoolInfo:              "SELECT p.name, sum(m.volbytes) AS bytes, count(m) AS volumes, (not exists(select * from jobmedia jm where jm.mediaid = m.mediaid)) AS prunable, COALESCE((m.lastwritten + (m.volretention * interval '1s')) < NOW(), false) as expired FROM media m LEFT JOIN pool p ON m.poolid = p.poolid GROUP BY p.name, prunable, expired",
		JobStates:             "SELECT JobStatus FROM status",
	},
}

// GetConnection opens a new db connection
func GetConnection(databaseType string, connectionString string, jobDiscoveryDays int) (*Connection, error) {
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
		db:               db,
		queries:          selectedQueries,
		jobDiscoveryDays: jobDiscoveryDays,
	}, nil
}

func (connection Connection) JobList() ([]JobInfo, error) {
	date := time.Now().AddDate(0, 0, -connection.jobDiscoveryDays).Format("2006-01-02")
	results, err := connection.execQuery(connection.queries.JobList, date)

	if err != nil {
		return nil, err
	}

	defer results.Close()

	var jobs []JobInfo

	for results.Next() {
		var jobInfo JobInfo
		err = results.Scan(&jobInfo.JobName, &jobInfo.JobType, &jobInfo.clientId, &jobInfo.ClientName, &jobInfo.FileSetName, &jobInfo.TotalCount, &jobInfo.TotalBytes, &jobInfo.TotalFiles)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, jobInfo)
	}

	return jobs, nil
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

func (connection Connection) execJobLookupQuery(query string, lookup *JobInfo) (*sql.Rows, error) {
	return connection.execQuery(query, lookup.JobName, lookup.clientId, lookup.FileSetName)
}

func (connection Connection) execLastJobLookupQuery(query string, lookup *JobInfo) (*LastJob, error) {
	results, err := connection.execJobLookupQuery(query, lookup)
	if err != nil {
		return nil, err
	}
	defer results.Close()

	var lastJob LastJob
	if results.Next() {
		err = results.Scan(&lastJob.JobStatus, &lastJob.JobBytes, &lastJob.JobFiles, &lastJob.JobErrors, &lastJob.JobStartDate, &lastJob.JobEndDate)
	}

	return &lastJob, err
}

// LastJob returns metrics for latest executed server backup
func (connection Connection) LastJob(lookup *JobInfo) (*LastJob, error) {
	return connection.execLastJobLookupQuery(connection.queries.LastJob, lookup)
}

// LastSuccessfulJob returns metrics for latest successfully executed server backup
func (connection Connection) LastSuccessfulJob(lookup *JobInfo) (*LastJob, error) {
	return connection.execLastJobLookupQuery(connection.queries.LastSuccessfulJob, lookup)
}

// LastSuccessfulFullJob returns metrics for latest successfully executed server backup with level Full
func (connection Connection) LastSuccessfulFullJob(lookup *JobInfo) (*LastJob, error) {
	return connection.execLastJobLookupQuery(connection.queries.LastSuccessfulFullJob, lookup)
}

func (connection Connection) JobStates() ([]string, error) {
	results, err := connection.execQuery(connection.queries.JobStates)

	if err != nil {
		return nil, err
	}

	defer results.Close()

	var states []string

	for results.Next() {
		var state string
		err = results.Scan(&state)
		if err != nil {
			return nil, err
		}
		states = append(states, state)
	}

	return states, nil
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
