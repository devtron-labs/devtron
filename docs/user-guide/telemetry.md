# Telemetry


- Note: Devtron does not sent personal information and URLs from client cluster 

## OVERVIEW

Devtron uses telemetry data for monitoring and managing clients engaged with our open source application. to provide
better service and feature we have to measure the current usage.

Once a client starts installation, applications start sending events on different stages. commonly `installation start`
, `installation in progress`,
`installation failure`, `upgrade start`,`upgrade success`,`upgrade failure` etc.

## Events sent from these applications

`Inception (operator)`

`Devtron (orchestrator)`

`Dashboard`

* Sample event json below for all types of events comes from different applications
* Only properties are different for each event type. for example below is the sample event for heartbeat event.

| Key | Description |
| :--- | :--- |
| `event` | Name of the new app you want to Create |
| `distinct_id` | Unique user id or client id|
| `devtronVersion` | devtron version |
| `serverVersion` | kubernetes cluster version |
| `eventType` | event type |
| `ucid` | Unique client id |

```json
{
  "id": "017ah6af-8h60-0000-abfc-a0a25hd823d6",
  "timestamp": "2021-06-29T07:33:02.001000+00:00",
  "event": "Heartbeat",
  "distinct_id": "qadgrtuxogziz8ak",
  "properties": {
    "$geoip_city_name": "Columbus",
    "$geoip_continent_code": "NA",
    "$geoip_continent_name": "North America",
    "$geoip_country_code": "US",
    "$geoip_country_name": "United States",
    "$geoip_latitude": 39.9625,
    "$geoip_longitude": -83.0061,
    "$geoip_postal_code": "43215",
    "$geoip_subdivision_1_code": "OH",
    "$geoip_subdivision_1_name": "Ohio",
    "$geoip_time_zone": "America/New_York",
    "$ip": "18.117.165.2",
    "$lib": "posthog-go",
    "$lib_version": "1.0.2",
    "devtronVersion": "v1",
    "eventType": 0,
    "serverVersion": "v1.17.17",
    "timestamp": "2021-06-29T07:33:02.001372393Z",
    "ucid": "qadgrtuxogziz8ak"
  }
}
```

### 1. Event Sends from Inception (operator)

* `installation start`
* `installation in progress`
* `installation failure`
* `upgrade start`
* `upgrade success`
* `upgrade failure`


* All the above event's come from the operator when devtron is being installed or upgraded. event is same as sample json
  except `event:"InstallationStart"` changes.

### 2. Event Sends from devtron (orchestrator)

* `Heartbeat`
* `Summary`


* All the above event's come from the operator when devtron is being installed or upgraded. event is same as sample json
  except `event:"Heartbeat"` changes.
* `Summary` events send the daily operation done by user, single event in 24 hour sent the below data.

| Key | Description |
| :--- | :--- |
| `cdCountPerDay` | cd pipeline created in last 24 hour |
| `ciCountPerDay` | ci pipeline created in last 24 hour |
| `clusterCount` | total cluster in the system |
| `environmentCount` | total environment in the system |
| `nonProdAppCount` | total non prod apps created |
| `userCount` | total user created in the system |

```json
{
  "summary": {
    "cdCountPerDay": 1,
    "ciCountPerDay": 1,
    "clusterCount": 1,
    "environmentCount": 1,
    "nonProdAppCount": 1,
    "userCount": 2
  }
}
```

### 3. Event Sends from dashboard

* `"identify`

* All the above event's come from Dashboards when unique users visited the dashboard for the first time. event is same as 
  sample json except `"event": "$identify"` changes.