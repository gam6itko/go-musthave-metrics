# go-musthave-metrics

## server

### env

Значения по-умолчанию
```dotenv
ADDRESS=localhost:8080
STORE_INTERVAL=300
FILE_STORAGE_PATH=/tmp/metrics-db.json
RESTORE=true
DATABASE_DSN="postgres://postgres:password@postgres:5432/yp_metrics"
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

### iter12

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

### iter12

Сохраните профиль потребления памяти.

Запустим 2 сервиса и дадим им чуток поработать.

#### server

```shell
mkdir -p ./profiles/server
mkdir -p ./profiles/client
curl http://localhost:8080/debug/pprof/allocs > ./profiles/server/allocs.base.pprof
# go tool pprof -http=":9090" -seconds=30 http://localhost:8080/debug/pprof/allocs
go tool pprof -http=":9090" -seconds=30 ./profiles/server/allocs.base.pprof
```

Тут видно что больше всего памяти потребляет compress/flate.NewWriter.

#### client

```shell
mkdir -p ./profiles/client
curl http://localhost:8081/debug/pprof/allocs > ./profiles/client/allocs.base.pprof
# go tool pprof -http=":9090" -seconds=30 http://localhost:8081/debug/pprof/allocs
go tool pprof -http=":9090" -seconds=30 ./profiles/client/allocs.base.pprof
```

Тут видно что больше всего памяти потребляет compress/flate.NewWriter.