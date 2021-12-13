package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4"
)

var connectionString string

var (
	exporterPort     = flag.Int("port", 9625, "Bareos exporter port")
	exporterEndpoint = flag.String("endpoint", "/metrics", "Bareos exporter endpoint")
	databaseURL      = flag.String("dsn", "mysql://bareos@unix()/bareos?parseTime=true", "Bareos database DSN")
	jobDiscoveryDays = flag.Int("job-discovery-days", 7, "Number of days in the past that will be searched for jobs")
)

func splitDsn(dsn string) (string, string, error) {
	var splitDsn = strings.SplitN(dsn, "://", 2)
	if len(splitDsn) != 2 {
		return "", "", fmt.Errorf("Database DSN is incomplete: missing protocol: %s", dsn)
	}
	return splitDsn[0], splitDsn[1], nil
}

func main() {
	flag.Parse()

	dbType, connectionString, err := splitDsn(*databaseURL)
	if err != nil {
		log.Fatal(err)
	}

	connection, err := GetConnection(dbType, connectionString, *jobDiscoveryDays)
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
