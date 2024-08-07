#!/bin/bash

function error_command() {
    echo "No such command $command"
    help
}

function help() {
    echo "usage:"
    echo "  run:"
    echo "    $0 db               - start database service with visualization (influxdb, grafana)"
    echo "    $0 server           - start server (write data to influxdb)"
    echo "  stop:"
    echo "    $0 stop             - stop all project services"
}

function start_database() {
    source ./influx-grafana/.env 2>/dev/null
    echo "Starting influxdb and grafana services"
    docker-compose -f ./influx-grafana/influx-grafana.yml up -d
    echo "----------------------------------------"
    echo "Accessing the services:"
    echo "----------------------------------------"
    echo "InfluxDB:"
    echo "  - Access the InfluxDB at http://localhost:${INFLUXDB_PORT:-8086}"
    echo "  - Username: ${INFLUXDB_USER:-my-user}"
    echo "  - Password: ${INFLUXDB_PASSWORD:-my-password}"
    echo "----------------------------------------"
    echo "Grafana:"
    echo "  - Access Grafana at http://localhost:${GRAFANA_PORT:-3000}"
    echo "----------------------------------------"
}

function start_server() {
    echo "Starting server"
    echo "Data is read from HTTP requests and written to InfluxDB"
    cd server
    npm ci
    npm start
}

function stop_services() {
    echo "Stopping all services"
    docker-compose -f ./influx-grafana/influx-grafana.yml down
}

# Check if the script is run from the git root directory
if [[ $PWD != $(git rev-parse --show-toplevel) ]]; then
    echo "This script must be run from the root directory of the repository"
    exit 1
fi

# Check if command is empty
if [ $# -eq 0 ]; then
    help
    exit 1
fi

for command in "$@"; do
    case $command in
        db) start_database ;;
        server) start_server ;;
        stop) stop_services ;;
        -h) help ;;
        --help) help ;;
        build-docs)  pandoc -f markdown docs/REPORT.md -s -o report.pdf;;
        *) error_command ;;
    esac
done
