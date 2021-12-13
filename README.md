# bareos_exporter

[![Go Report Card](https://goreportcard.com/badge/github.com/b1-systems/bareos_exporter)](https://goreportcard.com/report/github.com/b1-systems/bareos_exporter)

[Prometheus](https://github.com/prometheus) exporter for [bareos](https://github.com/bareos) data recovery system

## [`Dockerfile`](./Dockerfile)

## Usage with [docker](https://hub.docker.com/r/vierbergenlnars/bareos_exporter)

1. Create a file containing your mysql password and mount it inside `/bareos_exporter/pw/auth`
2. **(optional)** [Overwrite](https://docs.docker.com/engine/reference/run/#env-environment-variables) default args using ENV variables
3. Run docker image as follows

```bash
docker run --name bareos_exporter -p 9625:9625 -d b1-systems/bareos_exporter:latest -dbtype mysql -dsn mysql://user:password@host/dbname
```

## Metrics

### Aggregated metrics for all jobs

These metrics are aggregated across all jobs with the same name, type, client and fileset that are in the catalog.

Metrics:

* `bareos_jobs_run`: Total number of jobs that have run with the parameters from the labels
* `bareos_files_saved`: Total number of files saved during all backups with the parameters from the labels
* `bareos_bytes_saved`: Toal number of bytes saved during all backups with the parameters from the labels

Labels:

* `jobname`: Name of the job
* `jobtype`: Type indication of the job (`B` is backup, `O` is consolidate, `R` is restore)
* `client`: Name of the client of the job
* `fileset`: Name of the fileset of the job

### Metrics for the latest job

These metrics are for the latest job with the same name, type, client and fileset.

The `whichjob` label indicates what criterium was used to select the latest job.

Metrics:

* `bareos_last_job_bytes_saved`: Number of bytes saved during the latest job
* `bareos_last_job_files_saved`: Number of files saved during the latest job
* `bareos_last_job_errors`: Number of errors that occured during the latest job
* `bareos_last_job_start_unix_timestamp`: Timestamp of the start time of the latest job
* `bareos_last_job_end_unix_timestamp`: Timestamp of the end time of the latest job
* `bareos_last_job_status`: Current status of the latest job. This is a binary metric (0 or 1) The value will be 1 only for the `status` label that currently applies for the job

Labels:

* `jobname`: Name of the job
* `jobtype`: Type indication of the job (`B` is backup, `O` is consolidate, `R` is restore)
* `client`: Name of the client of the job
* `fileset`: Name of the fileset of the job
* `last_selector`: Selector used for determining the last job
  * `last`: The job with the highest start time
  * `last_successful`: The job with the highest start time that terminated succesfully
  * `last_successful_full`: The job with the highest start time that terminated succesfully and is a full backup

### Metrics for pools

These metrics are aggregates for all volumes in a pool

Metrics:

* `bareos_pool_bytes`: Number of bytes stored in a pool
* `bareos_pool_volumes`: Number of volumes stored in a pool

Labels:

* `pool`: The pool which the volumes are member of
* `prunable`: `true` or `false` depending on whether there are still jobs referencing the counted volumes
* `expired`: `true` or `false` depending on whether the volume has expired or not

## Flags

| Name               | Description                                                                                                                                            | Default       |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------- |
| port               | Bareos exporter port                                                                                                                                   | `9625`        |
| endpoint           | Bareos exporter endpoint.                                                                                                                              | `/metrics`    |
| dbtype             | Database type (`mysql` or `pgx` for PostgreSQL)                                                                                                        | `pgx`         |
| dsn                | Data source name understood by [go-sql-driver/mysql] or [jakec/pgx]                                                                                    | `postgres://` |
| job-discovery-days | Number of days to look in the history to discover active jobs. Only jobs that have last run this number of days will be considered for data reporting. | 7             |

## TODO

* [ ] Write some tests.
