# go-musthave-metrics

## server

### env

Значения по-умолчанию
```dotenv
ADDRESS=localhost:8080
STORE_INTERVAL=300
FILE_STORAGE_PATH=/tmp/metrics-db.json
RESTORE=true
DATABASE_DSN="host=postgres user=postgres password=password dbname=yp_metrics sslmode=disable"
```

### get docker ip

```shell
docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' yp-metrics-postgres
```

## sprint

### iter7

#### update

```http request
POST localhost:8080/update/
Content-Type: application/json

{
  "id": "PollCount",
  "type": "counter",
  "delta": 1
}
```

#### get value 

```http request
POST localhost:8080/value/
Content-Type: application/json

{
  "id": "PollCount", 
  "type": "counter", 
}
```