package main

import (
	"context"
	"fmt"
	"log"
	"time"
	"os"
	"flag"
	"gopkg.in/yaml.v2"
	"encoding/json"
	"net/http"
	// "net"


	// "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/gorilla/mux"

)

type Config struct {
    Influxdb struct {
        Endpoint    	string  `yaml:"endpoint"`
		Token			string  `yaml:"token"`
		// Organisation	string  `yaml:"organisation"`
		// Bucket			string  `yaml:"bucket"`
    } `yaml:"influxdb"`

    REST_API struct {
        Port 			int 	`yaml:"port"`
    } `yaml:"rest_api"`
}


// Define the structure of the JSON input
type System_Metric struct {
	Time 		time.Time	`json:"time"`
	Server     	string  	`json:"server"`
	MetricType 	string  	`json:"metric_type"`
	Value      	float64 	`json:"value"`
}

func main() {
	// TODO: Add the log flags to the config file
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	
	// configFile variable stores the config file location
    // supplied on the command line with `--config`
	configFile := get_arg_config_file()
	log.Printf("")

    // var config Config
    config := load_config (configFile)
    fmt.Printf("InfluxDB endpoint: %s\n", config.Influxdb.Endpoint)

	appState := &AppState{
        Config: config,
    }

	// Define the REST API server
    r := mux.NewRouter()
    r.HandleFunc("/metrics/add/organisation/{org}/bucket/{bucket}/measurement/{measure}", appState.addMetricHandler).Methods("POST")

	port := fmt.Sprintf(":%d", config.REST_API.Port)
    http.Handle("/", r)
    fmt.Printf("Metrics API started at: http://localhost:%s/\n", port)
    log.Fatal(http.ListenAndServe(port, nil))
}

type AppState struct {
    Config Config
}

func (app *AppState) addMetricHandler (w http.ResponseWriter, r *http.Request) {
	log.Printf("In Handler")
    vars := mux.Vars(r)
    org := vars["org"]
    bucket := vars["bucket"]
    measure := vars["measure"]

	var metric System_Metric
    if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Failed to decode body")
        return
    }

	log.Printf("Received metric for org=%s, bucket=%s, measurement=%s: %+v\n", org, bucket, measure, metric)

	// org := config.Influxdb.Organisation
	// bucket := config.Influxdb.Bucket
	// measure := "system_metrics"

	// TODO: Remove this assignment once the agent/client is function as the time will be set by the agent
	// TODO: But while I am testing with curls, iti s just easier to set time here 
	metric.Time = time.Now()
	log.Printf("Received metric for org=%s, bucket=%s, measurement=%s: %+v\n", org, bucket, measure, metric)
	// TODO: Decouple the incoming API request from the outgoing sed to influxdb
	// This is an ainti pattern that makes the metrics system brittle.
	// If influx becomes un unavailable even for a short period of time incoming messages will just fail.
	// Instead queue the incoming messages, even in a persistent filesystem and use
	// another looping function to consume the queue and `send_to_influxdb()`  
	err := send_to_influxdb(app.Config, org, bucket, measure, metric)
	if err != nil {
		log.Printf("Failed to write point to InfluxDB: %v", err)
	}

}

func get_arg_config_file () (string) {
	// Define the --config flag
    var configFile string
	flag.StringVar(&configFile, "config", "", "Path to the config file")

	// Parse command-line flags
	flag.Parse()

	// Check for a value
	if configFile == "" {
		log.Fatalf("No config file specified. Use `--config`")
	}
    fmt.Printf("Using config file: %s\n", configFile)
    return configFile
}

// Load the config from the supplied configFile location
func load_config(configFile string) (Config) {
    var config Config
    file, err := os.Open(configFile)
    if err != nil {
        log.Fatalf("The config file `%s` does not exist", configFile)
    }
    defer file.Close()

    decoder := yaml.NewDecoder(file)
    err = decoder.Decode(&config)
    if err != nil {
        log.Fatalf("Can not parse the config file `%s` - %s", configFile, err)
    }

    // fmt.Printf("Server API: %s\n", config.API.Endpoint)
    return config
}

func send_to_influxdb(config Config, org string, bucket string, measure string, metric System_Metric) error {
	// Create a custom http.Transport that only uses IPv4
	// transport := &http.Transport{
	// 	DialContext: (&net.Dialer{
	// 		Timeout:   30 * time.Second,
	// 		KeepAlive: 30 * time.Second,
	// 		DualStack: false, // Disable dual-stack (IPv4 and IPv6) support
	// 	}).DialContext,
	// }

	// // Create the InfluxDB client configuration
	// clientOptions := influxdb2.DefaultOptions().
	// 	SetHTTPClient(&http.Client{
	// 		Transport: transport,
	// 	})
	
	// Set up InfluxDB client
	//client := influxdb2.NewClientWithOptions(config.Influxdb.Endpoint, config.Influxdb.Token, clientOptions)
	// client := influxdb2.NewClient("http://127.0.0.1:8086/", config.Influxdb.Token)
	client := influxdb2.NewClient(config.Influxdb.Endpoint, config.Influxdb.Token)
	defer client.Close()

	// Get a write API instance		
	writeAPI := client.WriteAPIBlocking(org, bucket)

	// Write metric to InfluxDB
	p := influxdb2.NewPointWithMeasurement(measure).
		AddTag("server", metric.Server).
		AddTag("metric_type", metric.MetricType).
		AddField("value", metric.Value).
		SetTime(time.Now())
		// SetTime()

	if err := writeAPI.WritePoint(context.Background(), p); err != nil {
		// log.Fatalf("Failed to write point to InfluxDB: %v", err)
		log.Printf("Failed to write point to InfluxDB: %v", err)
		return err
	}
	log.Printf("Stored metric in influxDB org=%s, bucket=%s, measurement=%s: %+v\n", org, bucket, measure, metric)
	return nil

}