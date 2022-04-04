#!/bin/env sh
#set -x

# Wait for DB (default 0)
sleep ${WAIT_FOR_DB}

# Run Dockerfile CMD
if [ "$1" == './bareos_exporter' ]; then
  if [ -z "${DB_TYPE}" ]; then
    echo 'Error: DB_TYPE is empty'
    exit 1
  fi
  if [ -z "${DB_PASSWORD}" ]; then
    echo 'Warning: DB_PASSWORD is empty'
  fi
  if [ "${DB_TYPE}" == 'postgres' ]; then
    exec ./bareos_exporter -endpoint ${ENDPOINT} -job-discovery-days ${JOB_DAYS} -port ${PORT} -dsn "${DB_TYPE}://host=${DB_HOST} dbname=${DB_NAME} sslmode=${SSL_MODE} user=${DB_USER} password=${DB_PASSWORD} port=${DB_PORT}"
  fi
  if [ "${DB_TYPE}" == 'mysql' ]; then
    exec ./bareos_exporter -endpoint ${ENDPOINT} -job-discovery-days ${JOB_DAYS} -port ${PORT} -dsn "${DB_TYPE}://${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}?parseTime=true"
  fi
else
  exec ./bareos_exporter "$@"
fi
