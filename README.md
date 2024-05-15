# go-musthave-metrics

## server

### env

Значения по-умолчанию
```dotenv
ADDRESS=localhost:8080
STORE_INTERVAL=300
FILE_STORAGE_PATH=/tmp/metrics-db.json
RESTORE=true
DATABASE_DSN=postgres://postgres:password@172.22.0.2:5432/yp_metrics
KEY=key
CRYPTO_KEY=/tmp/go-musthave-metrics/private.pem
```

### get docker ip

```shell
docker-compose up -d
docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' yp-musthave-metrics-postgres
```

## agent

```dotenv
ADDRESS=localhost:8080
REPORT_INTERVAL=10
POLL_INTERVAL=2
KEY="abc"
RATE_LIMIT=4
CRYPTO_KEY=/tmp/go-musthave-metrics/public.key
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

### iter16

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


### iter18

```shell
swag init --dir=./cmd/agent --output ./swagger/agent
swag init --dir=./cmd/server --output ./swagger/server
```


### iter20

Проверить код своим родным линтером
```shell
go build -o staticlint ./cmd/staticlint/
./staticlint ./internal/server/...
```

### iter21

Сгенерировать ключи командой
```shell
go build -o keygen ./cmd/keygen 
./keygen -path .
```
