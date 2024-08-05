# metrics
A simple metrics system

## Architecture
The design is to have a simple metrics-agent for pushing data at a simple API.  This will obfuscate all the features of influx db and only expose the features that are required via the API.

## Components

### data-store
With so many options available, it did not make sense to build a data store from scratch. The selection criteria for the store was:
- simple to set up
- easy maintain
- easy to use with go client

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
> user: admin

>  password: RedHatRedHat!!

### Metrics API (server)
The API will have several endpoints.  For example,  the following 2 endpoints are RESTful by design:
- Send Metrics
  ~~~
  POST: /metrics/add/organisation/{{org}}/bucket/{{bucket}}/measurement/{{measure}}
  BODY: {json:data}
  ~~~
  This will accept HTTP POST requests from clients, like metrics-agent, and then make another call to influxDB to store the requests/metric  
- Search Metrics
  ~~~
  POST: /metrics/search/organisation/{{org}}/bucket/{{bucket}}/measurement/{{measure}}
  BODY: {json:search}
  ~~~
  This will accept HTTP POST requests from clients and then make another call to influxDB to search the time series database and return results as JSON.

The API is configured by reading a YAML file. This file location is supplied to the agent using a --config argument.  The only feature implemented is inserting metrics into the database, all other features will follow a similar RESTful URL pattern, similar to the second search URL above

 #### Testing
 1. Ensure the metrics repo has been cloned locally and the Data Store above is running.

1. From the root of the metrics repo change to the data-store directory
   ~~~
   $ cd <metrics-root>/metrics-api/
   ~~~

1. Start the Metrics Agent.  By default this will log to stdout in the same terminal.
   ~~~
   $ ./start-metrics-api
   ...
   InfluxDB endpoint: http://127.0.0.1:8086/
   Metrics API started at: http://localhost::8080/
   ~~~
   This terminal will now be the log console for the api

1. Send a request to the REST server located at http://localhost:8080/
   ~~~
   $ curl -X POST http://localhost:8080/metrics/add/organisation/metrics/bucket/metrics/measurement/system_metrics \
   -H "Content-Type: application/json" \
   -d '{"time": "2024-08-01T23:31:00+10:00", "server": "server2", "metric_type": "memory", "value": 75}'
   ~~~
   You should see to log entries in in the console.
   1. Showing the message received into  the RESTful API server
   2. Showing the message successfully sent to the InfluxDB
   ~~~
   2024/08/02 13:21:26 main.go:98: Received metric for org=metrics, bucket=metrics, measurement=system_metrics: {Time:2024-08-02 13:21:26.811022 +1000 AEST m=+16338.624637501 Server:server1 MetricType:mem Value:73.24810028076172}
   
   2024/08/02 13:21:26 main.go:184: Stored metric in influxDB org=metrics, bucket=metrics, measurement=system_metrics: {Time:2024-08-02 13:21:26.811022 +1000 AEST m=+16338.624637501 Server:server1 MetricType:mem Value:73.24810028076172}
   ~~~

2. Open the [Data Explorer of the InfluxDB](http://localhost:8086/orgs/d85216fa4f7d689c/data-explorer) ( http://localhost:8086/orgs/d85216fa4f7d689c/data-explorer )
   
   1. Try using the old data Explorer window, there is a switch on the top right of the page
   2. View the metric that was just added as row in a simple table. The row should be located in the following silo:
      - **Bucket** - `metrics`
      - **Measurement** - `system_metrics`
      - **Tags* - These will change, but for the inserted metric via the curl above:
        - **server** - `server2`
        - **metric_type** - `memory`
   3. Change to the Simple Data view as its easier to view any added metrics as a row.
   4. curl more rows into the database using the same curl above, and view additional rows in the data explorer

### Metrics Agent (client)

This agent can be deployed on a node or in a pod.  The go process will poll the system it is running on and push the metrics to the metrics API.  It could have pushed it directly to the influxDB but the design choice is to obfuscate the influxDB and its features.  The Metrics API can then be used to draw out only the features that are required.  This design pattern will keep features less confusing for users or securely block them from using inappropriate features.  This also provides a chance to build a RESTful API (-:  

I chose to use to use [gopsutil](https://github.com/shirou/gopsutil) to pull metrics from the host.  CPU and memory metrics have been implemented as a PoC. Because gopsutil was used it will be easy to extend the agent to process more metrics in the future.

The agent is configured by reading a YAML file.  This file location is supplied to the agent using a `--config` argument.  For ease of testing, there is another flag `--hostname` which is is used to allow multiple agents on the same host to have different hostnames.


#### Testing the adding of metrics

1. Ensure the metrics repo has been cloned locally and the Data Store and Metrics API are running.

1. From the root of the metrics repo change to the data-store directory
   ~~~
   $ cd <metrics-root>/metrics-agent/
   ~~~

1. Start the metrics agent
   ~~~
   $ ./start-metrics-agent-1
   ~~~
   Check no errors are streaming from the agent

1. Check the [Data Explorer of the InfluxDB](http://localhost:8086/orgs/d85216fa4f7d689c/data-explorer) ( http://localhost:8086/orgs/d85216fa4f7d689c/data-explorer ) console to view the new rows that are being added by the agent


1. Additional agents can be started in separate terminals
   ~~~
   $ ./start-metrics-agent-2
   ...
   $ ./start-metrics-agent-3
   ~~~

1. Check the Metrics API logs and the Data Explorer to see these new hosts/servers are sending in metrics

#### Testing the searching of metrics
WIP - Work in progress

Development of the second API endpoint would need to be completed.

**Search Metrics**
  ~~~
  POST: /metrics/search/organisation/{{org}}/bucket/{{bucket}}/measurement/{{measure}}
  BODY: {json:search}
  ~~~

  Following a similar RESTful endpoint layout note how the root changes from `/metrics/add/` to `/metrics/search/` the rest of the url selects the correct data location to search and the body of the post will provide the search criteria.

I could have designed the API endpoint to allow clients to pass parameters in the URL `?key=value`, but this could leak sensitive data via the visible URL whereas, a post body is encrypted (when using https) between the client and the API server.
