package main

import (

    "os"
    "fmt"
    "flag"
    "log"
    "time"
    "encoding/json"
    // "net/http"
    
    "gopkg.in/yaml.v2"
    "github.com/shirou/gopsutil/cpu"
)

type Config struct {
    API struct {
        Endpoint    string  `yaml:"endpoint"`
    } `yaml:"metrics_api"`

    Metrics struct {
        CPU struct {
            Interval    time.Duration     `yaml:"interval"`
        }
    } `yaml:"metrics"`
}

type Metric struct {
    Measurement string                 `json:"measurement"`
    Tags         map[string]string      `json:"tags,omitempty"`
    Fields       map[string]interface{} `json:"fields"`
    Time         time.Time              `json:"time,omitempty"`
}

func main() {
    // configFile variable stores the config file location
    // supplied on the command line with `--config`
	configFile := get_arg_config_file()

    // var config Config
    config := load_config (configFile)
    // fmt.Printf("Server API: %s\n", config.API.Endpoint)

    ticker := time.NewTicker(config.Metrics.CPU.Interval * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            cpuPercentages, err := cpu.Percent(0, false)
            if err != nil {
                log.Fatalf("Error getting CPU percentage: %v", err)
            }
            fmt.Printf("CPU: %7.2f\n", cpuPercentages)

            metric := Metric{
                Measurement: "cpu_usage",
                Fields: map[string]interface{}{
                    "cpu_percent": cpuPercentages[0],
                },
                Time: time.Now(),
            }
            s, _ := json.MarshalIndent(metric, "", "\t")
            fmt.Printf(string(s))

            // data, err := json.Marshal(metric)
            // if err != nil {
            //     log.Fatalf("Error marshaling JSON: %v", err)
            // }

            // fmt.Println("Hello")

            // resp, err := http.Post("http://localhost:8186", "application/json", bytes.NewBuffer(data))
            // if err != nil {
            //     log.Fatalf("Error sending data to Telegraf: %v", err)
            // }
            // resp.Body.Close()
        }
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