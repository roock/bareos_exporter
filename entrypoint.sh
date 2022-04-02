#!/bin/env sh
#set -x

#Wait for DB init
sleep ${WAIT_FOR_DB}

# Run Dockerfile CMD
if [ -z "$@" ]; then
  exec ./bareos_exporter -port ${PORT} -dsn "postgres://host=${DB_HOST} dbname=${DB_NAME} sslmode=${SSL_MODE} user=${DB_USER} password=${DB_PASSWORD}"
else
  exec "$@" 
fi
