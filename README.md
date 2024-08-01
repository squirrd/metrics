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

### Metrics Agent
This agent can be deployed on a node or in pod.  The go process will poll the system it is running on and push the metrics to the metrics API.  It could have pushed it directly to the influxDB but the choice in this design is to obfuscate the influxDB and its features and only draw out the features that are required.  This also provides a chance to build a RESTful API.  

I chose to use to use [gopsutil](https://github.com/shirou/gopsutil) to pull metrics from the host.

For this exercise, I have deployed the go binary in a pod, this way multiple pods can be deployed at the same time for testing.







---
---
---
1. List
   ~~~
   
   ~~~