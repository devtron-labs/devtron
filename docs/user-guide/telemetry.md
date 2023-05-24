# Telemetry Guide

* [Introduction](#introduction)
* [What data is collected](#what-data-is-collected)
* [Where data is sent](#where-data-is-sent)


Introduction
============

Devtron collects anonymous telemetry data that helps the Devtron team in understanding how the product is being used and in deciding what to focus on next.

The data collected is minimal, **non PII**, statistical in nature and **cannot be used to uniquely identify an user**.

Please see the next section to see what data is collected and sent. Access to collected data is strictly limited to the Devtron team.

As a growing community, it is very valuable in helping us make the Devtron a better product for everyone!



What data is collected
======================

Here is a sample event JSON which is collected and sent:


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


| Key | Description |
| :--- | :--- |
| `event` | Name of the event |
| `distinct_id` | Unique user id or client id|
| `devtronVersion` | devtron version |
| `serverVersion` | kubernetes cluster version |
| `eventType` | event type |
| `ucid` | Unique client id |



### Inception (operator)
Inception sends the installation and upgradation events of the Devtron tool to measure the churn rate.

Events which are sent by Inception :
* `InstallationStart`
* `InstallationInProgress`
* `InstallationSuccess`
* `UpgradeStart`
* `UpgradeInProgress`
* `UpgradeSuccess`

Event is same as sample json with event name mentioned above.

### Devtron (orchestrator)
Orchestrator sends the summary events of the Devtron tool to measure the daily usage.

Events which are sent by Orchestrator :
* `Heartbeat`
* `Summary`

Orchestrator sends the `Summary` event once in 24 hours with the daily operation done by user.

Here is a sample summary JSON which is available under properties:

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

| Key | Description |
| :--- | :--- |
| `cdCountPerDay` | cd pipeline created in last 24 hour |
| `ciCountPerDay` | ci pipeline created in last 24 hour |
| `clusterCount` | total cluster in the system |
| `environmentCount` | total environment in the system |
| `nonProdAppCount` | total non prod apps created |
| `userCount` | total user created in the system |

### Dashboard(Not Collected Anymore)
<del>Dashboard sends the events to measure dashboard visit of the Devtron tool.</del>

<del>Events which are sent by Orchestrator :</del>
<del>* `identify`</del>

<del>Dashboard sends the `identify` event when user visits the Dashboard for the first time.</del>


Where data is sent
======================

The data is sent to Posthog server.