#!/bin/env sh
#set -x

# Wait for DB (default 0)
sleep ${WAIT_FOR_DB}

# Run Dockerfile CMD
if [ "$#" -eq 1 -a "$1" == './bareos_exporter' ]; then
  if [ "${DB_TYPE}" == 'postgres' ]; then
    exec ./bareos_exporter -endpoint ${ENDPOINT} -job-discovery-days ${JOB_DAYS} -port ${PORT} -dsn "${DB_TYPE}://host=${DB_HOST} dbname=${DB_NAME} sslmode=${SSL_MODE} user=${DB_USER} password=${DB_PASSWORD} port=${DB_PORT} ${EXTRA_OPTS}"
  else
    exec ./bareos_exporter -endpoint ${ENDPOINT} -job-discovery-days ${JOB_DAYS} -port ${PORT} -dsn "${DB_TYPE}://${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}?parseTime=true"
  fi
else
  exec "$@"
fi
