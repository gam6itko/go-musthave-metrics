# go-musthave-metrics

## server

### env

Значения по-умолчанию
```dotenv
ADDRESS=localhost:8080
STORE_INTERVAL=300
FILE_STORAGE_PATH=/tmp/metrics-db.json
RESTORE=true
DATABASE_DSN="postgres://postgres:password@172.22.0.2:5432/yp_metrics"
KEY="key"
```

### get docker ip

```shell
docker-compose up -d
docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' yp-metrics-postgres
```

## agent

```dotenv
ADDRESS=localhost:8080
REPORT_INTERVAL=10
POLL_INTERVAL=2
KEY="abc"
RATE_LIMIT=4
```

# sprint

## iter7

### update

```http request
POST localhost:8080/update/
Content-Type: application/json

{
  "id": "PollCount",
  "type": "counter",
  "delta": 1
}
```

### get value 

```http request
POST localhost:8080/value/
Content-Type: application/json

{
  "id": "PollCount", 
  "type": "counter", 
}
```

## iter12

```http request
POST localhost:8080/updates/
Content-Type: application/json

[
  {
    "id": "PollCount",
    "type": "counter",
    "delta": 1
  },
  {
    "id": "GaugeABC",
    "type": "gauge",
    "value": 19.17
  }
]
```