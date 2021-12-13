package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var (
	exporterPort     = flag.Int("port", 9625, "Bareos exporter port")
	exporterEndpoint = flag.String("endpoint", "/metrics", "Bareos exporter endpoint")
	databaseURL      = flag.String("dsn", "postgres://bareos", "Bareos database DSN")
	databaseType     = flag.String("dbtype", "pgx", "mysql for MySQL/MariaDB, pgx for PostgreSQL")
	jobDiscoveryDays = flag.Int("job-discovery-days", 7, "Number of days in the past that will be searched for jobs")
)

func main() {
	flag.Parse()

	connection, err := GetConnection(*databaseType, *databaseURL, *jobDiscoveryDays)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()
	collector := bareosCollector(connection)
	prometheus.MustRegister(collector)

	http.Handle(*exporterEndpoint, promhttp.Handler())
	log.Info("Beginning to serve on port ", *exporterPort)

	addr := fmt.Sprintf(":%d", *exporterPort)
	log.Fatal(http.ListenAndServe(addr, nil))
}
