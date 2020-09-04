# dumb-metrics-api

## The test

Build a metric logging and reporting service that sums metrics by time window
for the most recent hour. Build a lightweight web server that
implements the two main APIs defined below.

### POST metric

POST ​/metric/​{key}`

```json

{​"value"​: ​30}
```

Response: 200 OK

```json
{}
```

## GET metric sum

Returns the sum of all metrics reported for this key over the past hour

Request: `GET ​/metric/​{key}/sum`

Response: 200 OK

```json
{"value": ​400}
```

## Caveats

- All values will be rounded to the nearest integer.
- You can get rid of any reported data after it is more than an hour old since
  we only need to see the most recent hour.

## Usage example

Imagine these are the events logged to your service for a metric "active_visitors"

```json
// 2 hours ago
POST ​/metric/​active_visitors
{​"value"​ = ​4​ }

// 30 minutes ago
POST ​/metric/​active_visitors
{​"value"​ = ​3​ }

// 40 seconds ago
POST ​/metric/​active_visitors
{​"value"​ = ​7​ }

// 5 seconds ago
POST ​/metric/​active_visitors
{​"value"​ = ​2​ }
```

These are the results expected from calling get aggregates:

```json
GET ​/metric/​active_visitors​/sum ​
// returns 12
```

Note that the metric posted 2 hours ago is not included in the sum since we only
care about data in the most recent hour for these API's

## Compiling and Running

### Compile

```sh
> go build server.go
> ./server
```

### Run

```sh
> go run server.go
```

### Consume REST API

#### GET

```sh
curl http://localhost:8080/metric/active_visitors/sum
```

```sh
{}
```

#### POST

```sh
curl http://localhost:8080/metric/active_visitors \
-H "Content-Type: application/json" \
-d '{"value":5}'
```

## ENV vars

- `HOST`defaults to `":"`
- `PORT`defaults to `"8080"`
- `DATA_TIMEOUT_MINUTES` default value `60` minutes

## TODO

- Add docker (Dockerfile, with golang multi-staged compile, docker-compose.yml file also for dev/test)
- Add a file db, or maybe BoltDB, BuntDB, ram-sql, goramdb, etc. and pipe in to a MetricDB interface
