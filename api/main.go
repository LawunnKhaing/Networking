package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Job struct {
	JobType string  `json:"jobType"`
	Value   float64 `json:"value"`
	Time    string  `json:"time"`
}

type JobService struct {
	logger   *zap.SugaredLogger
	queryAPI api.QueryAPI
}

func NewJobService(logger *zap.SugaredLogger, queryAPI api.QueryAPI) *JobService {
	return &JobService{
		logger:   logger,
		queryAPI: queryAPI,
	}
}

func (js *JobService) JobHandler(w http.ResponseWriter, r *http.Request) {
	jobs, err := js.getJobs()

	if err != nil {
		js.logger.Error("Error getting jobs: ", zap.Error(err))

		errorResponse := map[string]string{
			"error": "Error getting jobs: " + err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	jobsJson, err := json.Marshal(jobs)
	if err != nil {
		js.logger.Error("Error marshalling jobs: ", zap.Error(err))

		errorResponse := map[string]string{
			"error": "Error marshalling jobs: " + err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jobsJson)
}

func (js *JobService) getJobs() ([]Job, error) {
	query := `from(bucket: "my-bucket")
    |> range(start: -1000h)
    |> filter(fn: (r) => r["_measurement"] == "jobs")
    |> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
    |> limit(n: 10)`

	results, err := js.queryAPI.Query(context.Background(), query)
	if err != nil {
		return []Job{}, fmt.Errorf("error querying data: %w", err)
	}
	defer results.Close()

	var jobs_data []Job

	for results.Next() {
		record := results.Record()
		jobType, ok := record.ValueByKey("jobType").(string)
		if !ok {
			return []Job{}, fmt.Errorf("error parsing jobType")
		}
		value, ok := record.ValueByKey("value").(float64)
		if !ok {
			return []Job{}, fmt.Errorf("error parsing value")
		}
		time := record.Time().Format(time.RFC3339)

		jobs_data = append(jobs_data, Job{
			JobType: jobType,
			Value:   value,
			Time:    time,
		})
	}

	return jobs_data, nil
}

func main() {
	// Create a new logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	sugar := logger.Sugar()

	err := godotenv.Load()
	if err != nil {
		sugar.Fatal("Error loading .env file")
	}
	token := os.Getenv("INFLUX_TOKEN")
	url := os.Getenv("INFLUX_URL")
	org := os.Getenv("INFLUX_ORG")
	sugar.Info("Starting server")
	sugar.Infof("Token: %s", token)
	sugar.Infof("URL: %s", url)
	sugar.Infof("Org: %s", org)

	client := influxdb2.NewClient(url, token)
	queryAPI := client.QueryAPI(org)

	js := NewJobService(sugar, queryAPI)

	router := mux.NewRouter()
	router.HandleFunc("/jobs", js.JobHandler)
	sugar.Fatal(http.ListenAndServe(":8000", router))
}
