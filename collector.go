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

	LastFullJobBytes     *prometheus.Desc
	LastFullJobFiles     *prometheus.Desc
	LastFullJobErrors    *prometheus.Desc
	LastFullJobTimestamp *prometheus.Desc

	ScheduledJob *prometheus.Desc

	PoolBytes   *prometheus.Desc
	PoolVolumes *prometheus.Desc

	connection *Connection
}

func bareosCollector(conn *Connection) *bareosMetrics {
	return &bareosMetrics{
		TotalJobs: prometheus.NewDesc("bareos_jobs_run_total",
			"Total backup jobs for hostname combined",
			[]string{"hostname"}, nil,
		),
		TotalFiles: prometheus.NewDesc("bareos_files_saved_total",
			"Total files saved for server during all backups for hostname combined",
			[]string{"hostname"}, nil,
		),
		TotalBytes: prometheus.NewDesc("bareos_bytes_saved_total",
			"Total bytes saved for server during all backups for hostname combined",
			[]string{"hostname"}, nil,
		),
		LastJobBytes: prometheus.NewDesc("bareos_last_backup_job_bytes_saved_total",
			"Total bytes saved during last backup for hostname",
			[]string{"hostname", "level"}, nil,
		),
		LastJobFiles: prometheus.NewDesc("bareos_last_backup_job_files_saved_total",
			"Total files saved during last backup for hostname",
			[]string{"hostname", "level"}, nil,
		),
		LastJobErrors: prometheus.NewDesc("bareos_last_backup_job_errors_occurred_while_saving_total",
			"Total errors occurred during last backup for hostname",
			[]string{"hostname", "level"}, nil,
		),
		LastJobStartTimestamp: prometheus.NewDesc("bareos_last_backup_job_unix_timestamp",
			"Execution start timestamp of last backup for hostname",
			[]string{"hostname", "level"}, nil,
		),
		LastJobEndTimestamp: prometheus.NewDesc("bareos_last_backup_job_end_unix_timestamp",
			"Execution end timestamp of last backup for hostname",
			[]string{"hostname", "level"}, nil,
		),
		LastJobStatus: prometheus.NewDesc("bareos_last_backup_job_status",
			"Termination status of the last backup for hostname",
			[]string{"hostname", "status"}, nil,
		),
		ScheduledJob: prometheus.NewDesc("bareos_scheduled_jobs_total",
			"Probable execution timestamp of next backup for hostname",
			[]string{"hostname"}, nil,
		),
		PoolBytes: prometheus.NewDesc("bareos_pool_bytes",
			"Total bytes saved in a pool",
			[]string{"pool", "prunable", "expired"}, nil,
		),
		PoolVolumes: prometheus.NewDesc("bareos_pool_volumes",
			"Total volumes in a pool",
			[]string{"pool", "prunable", "expired"}, nil,
		),
		connection: conn,
	}
}

func (collector *bareosMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.TotalFiles
	ch <- collector.TotalBytes
	ch <- collector.LastJobBytes
	ch <- collector.LastJobFiles
	ch <- collector.LastJobErrors
	ch <- collector.LastJobStartTimestamp
	ch <- collector.LastJobEndTimestamp
	ch <- collector.LastJobStatus
	ch <- collector.ScheduledJob
	ch <- collector.PoolBytes
	ch <- collector.PoolVolumes
}

var bareosTerminationStates = []string{
	"C", "R", "B", "T", "E", "e", "f", "D", "A", "I", "L", "W", "l", "q", "F", "S", "m", "M", "s", "j", "c", "d", "t", "p", "i", "a",
}

func (collector *bareosMetrics) Collect(ch chan<- prometheus.Metric) {

	var servers, getServerListErr = collector.connection.GetServerList()

	if getServerListErr != nil {
		log.WithFields(log.Fields{
			"method": "GetServerList",
		}).Error(getServerListErr)
	} else {

		for _, server := range servers {
			jobTotals, jobTotalsErr := collector.connection.JobTotals(server)
			lastServerJob, jobErr := collector.connection.LastJob(server)
			scheduledJob, scheduledJobErr := collector.connection.ScheduledJobs(server)
			lastJobStatus, lastJobStatusErr := collector.connection.LastJobStatus(server)

			if jobTotalsErr != nil || jobErr != nil || scheduledJobErr != nil || lastJobStatusErr != nil {
				log.Info(server)
			}

			if jobTotalsErr != nil {
				log.WithFields(log.Fields{
					"method": "JobTotals",
				}).Error(jobTotalsErr)
			}

			if jobErr != nil {
				log.WithFields(log.Fields{
					"method": "LastJob",
				}).Error(jobErr)
			}

			if scheduledJobErr != nil {
				log.WithFields(log.Fields{
					"method": "ScheduledJobs",
				}).Error(scheduledJobErr)
			}

			if lastJobStatusErr != nil {
				log.WithFields(log.Fields{
					"method": "LastJobStatus",
				}).Error(lastJobStatusErr)
			}

			ch <- prometheus.MustNewConstMetric(collector.TotalJobs, prometheus.CounterValue, float64(jobTotals.Count), server)
			ch <- prometheus.MustNewConstMetric(collector.TotalBytes, prometheus.CounterValue, float64(jobTotals.Bytes), server)
			ch <- prometheus.MustNewConstMetric(collector.TotalFiles, prometheus.CounterValue, float64(jobTotals.Files), server)

			ch <- prometheus.MustNewConstMetric(collector.LastJobBytes, prometheus.CounterValue, float64(lastServerJob.JobBytes), server, lastServerJob.Level)
			ch <- prometheus.MustNewConstMetric(collector.LastJobFiles, prometheus.CounterValue, float64(lastServerJob.JobFiles), server, lastServerJob.Level)
			ch <- prometheus.MustNewConstMetric(collector.LastJobErrors, prometheus.CounterValue, float64(lastServerJob.JobErrors), server, lastServerJob.Level)
			ch <- prometheus.MustNewConstMetric(collector.LastJobStartTimestamp, prometheus.CounterValue, float64(lastServerJob.JobStartDate.Unix()), server, lastServerJob.Level)
			ch <- prometheus.MustNewConstMetric(collector.LastJobEndTimestamp, prometheus.CounterValue, float64(lastServerJob.JobEndDate.Unix()), server, lastServerJob.Level)
			for _, terminationState := range bareosTerminationStates {
				var state = float64(0)
				if lastJobStatus != nil {
					if terminationState == *lastJobStatus {
						state = 1
					}
				}

				ch <- prometheus.MustNewConstMetric(collector.LastJobStatus, prometheus.CounterValue, state, server, terminationState)

			}

			ch <- prometheus.MustNewConstMetric(collector.ScheduledJob, prometheus.CounterValue, float64(scheduledJob.ScheduledJobs), server)

		}
	}

	var poolInfoList, poolInfoErr = collector.connection.PoolInfo()

	if poolInfoErr != nil {
		log.WithFields(log.Fields{
			"method": "PoolInfo",
		}).Error(poolInfoErr)
	} else {

		for _, poolInfo := range poolInfoList {
			ch <- prometheus.MustNewConstMetric(collector.PoolBytes, prometheus.CounterValue, float64(poolInfo.Bytes), poolInfo.Name, strconv.FormatBool(poolInfo.Prunable), strconv.FormatBool(poolInfo.Expired))
			ch <- prometheus.MustNewConstMetric(collector.PoolVolumes, prometheus.CounterValue, float64(poolInfo.Volumes), poolInfo.Name, strconv.FormatBool(poolInfo.Prunable), strconv.FormatBool(poolInfo.Expired))
		}
	}
}
