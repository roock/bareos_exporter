package main

import (
	"github.com/prometheus/client_golang/prometheus"

	log "github.com/sirupsen/logrus"
	"strconv"
)

type bareosMetrics struct {
	TotalJobs             *prometheus.Desc
	TotalFiles            *prometheus.Desc
	TotalBytes            *prometheus.Desc
	LastJobBytes          *prometheus.Desc
	LastJobFiles          *prometheus.Desc
	LastJobErrors         *prometheus.Desc
	LastJobStartTimestamp *prometheus.Desc
	LastJobEndTimestamp   *prometheus.Desc
	LastJobStatus         *prometheus.Desc

	PoolBytes   *prometheus.Desc
	PoolVolumes *prometheus.Desc

	connection *Connection
}

type jobLabels []string

var jobInfoLabelNames = jobLabels{"jobname", "jobtype", "client", "fileset"}

func (jobInfo JobInfo) toLabels() jobLabels {
	return jobLabels{jobInfo.JobName, string(jobInfo.JobType), jobInfo.ClientName, jobInfo.FileSetName}
}

func (s jobLabels) newAppend(label string) jobLabels {
	cpy := make(jobLabels, len(s), len(s)+1)
	copy(cpy, s)
	cpy = append(cpy, label)
	return cpy
}

var jobInfoAndWhichLabelNames = jobInfoLabelNames.newAppend("last_selector")

var poolInfoLabelNames = jobLabels{"pool", "prunable", "expired"}

func (poolInfo PoolInfo) toLabels() jobLabels {
	return jobLabels{poolInfo.Name, strconv.FormatBool(poolInfo.Prunable), strconv.FormatBool(poolInfo.Expired)}
}

func bareosCollector(conn *Connection) *bareosMetrics {
	return &bareosMetrics{
		TotalJobs: prometheus.NewDesc("bareos_jobs_run",
			"Total backup jobs for jobname",
			jobInfoLabelNames, nil,
		),
		TotalFiles: prometheus.NewDesc("bareos_files_saved",
			"Total files saved for during all jobs combined",
			jobInfoLabelNames, nil,
		),
		TotalBytes: prometheus.NewDesc("bareos_bytes_saved",
			"Total bytes saved for during all jobs combined",
			jobInfoLabelNames, nil,
		),
		LastJobBytes: prometheus.NewDesc("bareos_last_job_bytes_saved",
			"Bytes saved during last backup",
			jobInfoAndWhichLabelNames, nil,
		),
		LastJobFiles: prometheus.NewDesc("bareos_last_job_files_saved",
			"Files saved during last job",
			jobInfoAndWhichLabelNames, nil,
		),
		LastJobErrors: prometheus.NewDesc("bareos_last_job_errors",
			"Errors occurred during last job",
			jobInfoAndWhichLabelNames, nil,
		),
		LastJobStartTimestamp: prometheus.NewDesc("bareos_last_job_start_unix_timestamp",
			"Execution start timestamp of last job",
			jobInfoAndWhichLabelNames, nil,
		),
		LastJobEndTimestamp: prometheus.NewDesc("bareos_last_job_end_unix_timestamp",
			"Execution end timestamp of last job",
			jobInfoAndWhichLabelNames, nil,
		),
		LastJobStatus: prometheus.NewDesc("bareos_last_job_status",
			"Status of the last job",
			jobInfoAndWhichLabelNames.newAppend("status"), nil,
		),
		PoolBytes: prometheus.NewDesc("bareos_pool_bytes",
			"Total bytes saved in a pool",
			poolInfoLabelNames, nil,
		),
		PoolVolumes: prometheus.NewDesc("bareos_pool_volumes",
			"Total volumes in a pool",
			poolInfoLabelNames, nil,
		),
		connection: conn,
	}
}

func (collector *bareosMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.TotalJobs
	ch <- collector.TotalFiles
	ch <- collector.TotalBytes

	ch <- collector.LastJobBytes
	ch <- collector.LastJobFiles
	ch <- collector.LastJobErrors
	ch <- collector.LastJobStartTimestamp
	ch <- collector.LastJobEndTimestamp
	ch <- collector.LastJobStatus

	ch <- collector.PoolBytes
	ch <- collector.PoolVolumes
}

func (collector *bareosMetrics) Collect(ch chan<- prometheus.Metric) {

	var jobs, jobsErr = collector.connection.JobList()

	if jobsErr != nil {
		log.WithFields(log.Fields{
			"method": "JobList",
		}).Error(jobsErr)
	} else {

		for _, jobInfo := range jobs {
			ch <- prometheus.MustNewConstMetric(collector.TotalJobs, prometheus.CounterValue, float64(jobInfo.TotalCount), jobInfo.toLabels()...)
			ch <- prometheus.MustNewConstMetric(collector.TotalBytes, prometheus.CounterValue, float64(jobInfo.TotalBytes), jobInfo.toLabels()...)
			ch <- prometheus.MustNewConstMetric(collector.TotalFiles, prometheus.CounterValue, float64(jobInfo.TotalFiles), jobInfo.toLabels()...)

			lastJob, lastJobErr := collector.connection.LastJob(&jobInfo)
			lastSuccessfulJob, lastSuccessfulJobErr := collector.connection.LastSuccessfulJob(&jobInfo)
			lastSuccessfulFullJob, lastSuccessfulFullJobErr := collector.connection.LastSuccessfulFullJob(&jobInfo)

			if lastJobErr != nil {
				log.WithFields(log.Fields{
					"job":    jobInfo,
					"method": "LastJob",
				}).Error(lastJobErr)
			} else {
				collector.collectLastJob(ch, jobInfo, lastJob, "last")
			}

			if lastSuccessfulJobErr != nil {
				log.WithFields(log.Fields{
					"job":    jobInfo,
					"method": "LastSuccessfulJob",
				}).Error(lastSuccessfulJobErr)
			} else {
				collector.collectLastJob(ch, jobInfo, lastSuccessfulJob, "last_successful")
			}

			if lastSuccessfulFullJobErr != nil {
				log.WithFields(log.Fields{
					"job":    jobInfo,
					"method": "LastSuccessfulFullJob",
				}).Error(lastSuccessfulFullJobErr)
			} else {
				collector.collectLastJob(ch, jobInfo, lastSuccessfulFullJob, "last_successful_full")
			}
		}
	}

	var poolInfoList, poolInfoErr = collector.connection.PoolInfo()

	if poolInfoErr != nil {
		log.WithFields(log.Fields{
			"method": "PoolInfo",
		}).Error(poolInfoErr)
	} else {

		for _, poolInfo := range poolInfoList {
			ch <- prometheus.MustNewConstMetric(collector.PoolBytes, prometheus.CounterValue, float64(poolInfo.Bytes), poolInfo.toLabels()...)
			ch <- prometheus.MustNewConstMetric(collector.PoolVolumes, prometheus.CounterValue, float64(poolInfo.Volumes), poolInfo.toLabels()...)
		}
	}
}

func (collector *bareosMetrics) collectLastJob(ch chan<- prometheus.Metric, jobInfo JobInfo, lastJob *LastJob, whichLabel string) {

	labels := jobInfo.toLabels().newAppend(whichLabel)

	ch <- prometheus.MustNewConstMetric(collector.LastJobBytes, prometheus.CounterValue, float64(lastJob.JobBytes), labels...)
	ch <- prometheus.MustNewConstMetric(collector.LastJobFiles, prometheus.CounterValue, float64(lastJob.JobFiles), labels...)
	ch <- prometheus.MustNewConstMetric(collector.LastJobErrors, prometheus.CounterValue, float64(lastJob.JobErrors), labels...)
	ch <- prometheus.MustNewConstMetric(collector.LastJobStartTimestamp, prometheus.CounterValue, float64(lastJob.JobStartDate.Unix()), labels...)
	ch <- prometheus.MustNewConstMetric(collector.LastJobEndTimestamp, prometheus.CounterValue, float64(lastJob.JobEndDate.Unix()), labels...)

	var bareosTerminationStates, bareosTerminationStatesErr = collector.connection.JobStates()

	if bareosTerminationStatesErr != nil {
		log.WithFields(log.Fields{
			"job":    jobInfo,
			"method": "JobStates",
		}).Error(bareosTerminationStatesErr)
	} else {
		for _, terminationState := range bareosTerminationStates {
			var state = float64(0)
			if lastJob != nil {
				if terminationState == lastJob.JobStatus {
					state = 1
				}
			}

			ch <- prometheus.MustNewConstMetric(collector.LastJobStatus, prometheus.CounterValue, state, labels.newAppend(terminationState)...)

		}
	}
}
