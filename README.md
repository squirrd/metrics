# metrics
A simple metrics system

## Architecture
The design is to have a simple metrics-agent for pushing data at a simple API.  This will obfuscate all the features of influx db and only expose the features that are required via the API.

## Components

### data-store
With som many options available, it did not make sense to build a data store from scratch. The selection criteia for the store was:
- simple to set up
- easy maintain
- easy to use go client

I settled on influxDB.  It also meets most of the exercise requirements.

#### Testing
1. Ensure the metrics repo has been cloned locally.

1. From the root of the metrics repo change to the data-store directory
   ~~~
   $ cd <metrics-root>/data
   ~~~

1. Run the start script
   ~~~
   ./start-datastore
   ~~~

1. Check that influxDB is running by logging into the admin console
http://localhost:8086/
> user: adnin

>  password: 12345678

### Metrics API (server)
The API has several endpoints.  These endpoints are RESTful by design:
- Send Metrics
  ~~~
  /metrics/add/organisation/{{org}}/bucket/{{bucket}}/measurement/{{measure}}/{{ {json:data} }}
  ~~~
  This will accept HTTP POST requests from clients like metrics-agent and then make another call to influxDB to store the requests/metric  
- Search Metrics
  ~~~
  /metrics/search/organisation/{{org}}/bucket/{{bucket}}/measurement/{{measure}}/{{ {json:search} }}
  ~~~
  This will accept HTTP POST requests from clients like metrics-agent and then make another call to influxDB to search the time series database and return results as JSON.

  The API is configured by reading a YAML file. This file location is supplied to the agent using a --config argument

 #### Testing
 1. Ensure the metrics repo has been cloned locally.

1. From the root of the metrics repo change to the data-store directory
   ~~~
   $ cd <metrics-root>/metrics-api/
   ~~~

1. Start the Metrics Agent
   ~~~
   $ ./start-metrics-api
   ...
   InfluxDB endpoint: http://127.0.0.1:8086/
   Metrics API started at: http://localhost::8080/
   ~~~

1. Send a request to the REST server located at http://localhost:8080/
   ~~~
   curl -X POST http://localhost:8080/metrics/add/organisation/metrics/bucket/metrics/measurement/system_metrics \
   -H "Content-Type: application/json" \
   -d '{"time": "2024-08-01T23:31:00+10:00", "server": "server2", "metric_type": "memory", "value": 75}'
   ~~~
   You should it being processed by the RESTful API server and then sent to the InfluxDB
   ~~~
   2024/08/02 10:09:04 main.go:89: Received metric for org=metrics, bucket=metrics, measurement=system_metrics: {Time:2024-08-01 23:31:00 +1000 AEST Server:server2 MetricType:memory Value:72}
   2024/08/02 10:09:04 main.go:98: Received metric for org=metrics, bucket=metrics, measurement=system_metrics: {Time:2024-08-02 10:09:04.047186 +1000 AEST m=+4795.862145043 Server:server2 MetricType:memory Value:72}
   2024/08/02 10:09:04 main.go:184: Stored metric in influxDB org=metrics, bucket=metrics, measurement=system_metrics: {Time:2024-08-02 10:09:04.047186 +1000 AEST m=+4795.862145043 Server:server2 MetricType:memory Value:72}
   ~~~

1. Open the [Data Explorer of the InfluxDB](http://localhost:8086/orgs/f3adfeb5cb217564/data-explorer)
   
   1. Try using the old data Explorer - Switch on the top right
   3. View the metric that was just added as row in a simple table. The row should be located in:
      - **Bucket** - metrics
      - **Measurement** - system_metrics
      - **Tag** - server - for this inserted metric `server1`
      - **Tag** - metric_type - for this inserted metric `memory`
   4. curl more rows into the database using the same curl above, view additional rows in the data explorer

### Metrics Agent (client)
This agent can be deployed on a node or in pod.  The go process will poll the system it is running on and push the metrics to the metrics API.  It could have pushed it directly to the influxDB but the choice in this design is to obfuscate the influxDB and its features and only draw out the features that are required.  This also provides a chance to build a RESTful API (-:  

I chose to use to use [gopsutil](https://github.com/shirou/gopsutil) to pull metrics from the host.  Only the CPU metric was implemented as a PoC. Other metrics can be developed into the same agent.

The agent is configured by reading a YAML file.  This file location is supplied to the agent using a `--config` argument.  For ease of testing, there is another flag `--hostname` is used to allow multiple agents on the same host.

For this exercise, I have deployed the go binary in a pod, this way multiple pods/agent can be deployed at the same time for testing.

#### Testing

1. Ensure the metrics repo has been cloned locally.

1. From the root of the metrics repo change to the data-store directory
   ~~~
   $ cd <metrics-root>/metrics-agent/
   ~~~

1. Start the metrics agent
   ~~~
   $ ./start-metrics-agent
   ~~~

1. List
   ~~~
   
   ~~~

1. List
   ~~~
   
   ~~~

1. List
   ~~~
   
   ~~~

1. List
   ~~~
   
   ~~~





___
___
1. List
   ~~~
   
   ~~~

