package main

import (

    "os"
    "fmt"
    "flag"
    "log"
    "time"
    "encoding/json"
    "net/http"
    "bytes"
    
    "gopkg.in/yaml.v2"
    "github.com/shirou/gopsutil/cpu"
    "github.com/shirou/gopsutil/mem"
)

type Config struct {
    API struct {
        Endpoint        string  `yaml:"endpoint"`
        Organisation    string  `yaml:"organisation"`
        Bucket          string  `yaml:"bucket"`
        Measurement     string  `yaml:"measurement"`
    } `yaml:"metrics_api"`

    Metrics struct {
        Interval        time.Duration   `yaml:"interval"`
    } `yaml:"metrics"`
}

type Metric struct {
    Measurement string                 `json:"measurement"`
    Tags         map[string]string      `json:"tags,omitempty"`
    Fields       map[string]interface{} `json:"fields"`
    Time         time.Time              `json:"time,omitempty"`
}

func main() {
    // Parse command line arguments
    configFile, hostname := get_args()

    // var config Config
    config := load_config (configFile)
    // fmt.Printf("Server API: %s\n", config.API.Endpoint)

    ticker := time.NewTicker(config.Metrics.Interval * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            // Fetch cpu statistics
            cpuPercentages, err := cpu.Percent(0, false)
            if err != nil {
                log.Fatalf("Error getting CPU percentage: %v", err)
            }
            cpu := cpuPercentages[0]

            // Fetch memory statistics
            v, err := mem.VirtualMemory()
            if err != nil {
                log.Fatalf("Failed to get memory usage: %v", err)
            }
            mem := v.UsedPercent

            // TODO: Decouple the collection of metrics from the sending of metrics
            // This is an ainti pattern that makes the metrics system brittle.
            // If the metrics API becomes un unavailable even for a short period of time storage of metrics will just fail.
            // This lead to lost metrics
            // Instead queue the metrics in memory and then use 
            // another looping function to consume the queue and the metrics to the metrics API server`  

            err = send_metics(config, hostname, cpu, mem)          
            if err != nil {
                // TODO: Some of these errors can be recovered from 
                // TODO: Consider improving error handling
                log.Fatalf("Error sending request: %v", err)
            }
        }
    }
}

func get_args () (string, string) {
	// Define the --config flag
    var configFile string
	flag.StringVar(&configFile, "config", "", "Path to the config file")

    // Define the --hostname flag
    var hostname string
    flag.StringVar(&hostname, "hostname", "", "The hostname to use")


	// Parse command-line flags
	flag.Parse()

	// Check for a value
	if configFile == "" {
		log.Fatalf("No config file specified. Use `--config`")
	}

    if hostname == "" {
		log.Fatalf("No hostname specified. Use `--hostname`")
	}
    return configFile, hostname
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

func send_metics (config Config, hostname string, cpu float64, mem float64) error {
    host := config.API.Endpoint
    org := config.API.Organisation
    bucket := config.API.Bucket
    measure := config.API.Measurement

    url := fmt.Sprintf("%s/metrics/add/organisation/%s/bucket/%s/measurement/%s", host, org, bucket, measure)

    // Create the data to be sent in the POST request
    cpu_data := map[string]interface{}{
        "time":        time.Now().Format("2006-01-02T15:04:05-07:00"),
        "server":      hostname,
        "metric_type": "cpu",
        "value":       cpu,
    }
    mem_data := map[string]interface{}{
        "time":        time.Now().Format("2006-01-02T15:04:05-07:00"),
        "server":      hostname,
        "metric_type": "mem",
        "value":       mem,
    }

    err := send_to_api(cpu_data, url)
    if err != nil {
        log.Fatalf("Error sending to API: %v", err)
    }

    err = send_to_api(mem_data, url)
    if err != nil {
        log.Fatalf("Error sending to API: %v", err)
    }
    return nil
}

func send_to_api (data map[string]interface{}, url string) error {
    // Convert data to JSON
    jsonData, err := json.Marshal(data)
    if err != nil {
        log.Fatalf("Error marshalling JSON: %v", err)
    }
    log.Printf("Sending: %v \nto endpoint: %s", string(jsonData), url)
    

    // Create a new POST request
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        log.Fatalf("Error creating request: %v", err)
    }

    // Set the Content-Type header
    req.Header.Set("Content-Type", "application/json")

    // Send the request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Fatalf("Error sending request: %v", err)
        return err
    }
    defer resp.Body.Close()

    // Log the response status
    log.Printf("Response Status: %s", resp.Status)
    return nil
}