# go-musthave-metrics

## env

Значения по-умолчанию
```dotenv
ADDRESS=localhost:8080
STORE_INTERVAL=300
FILE_STORAGE_PATH=/tmp/metrics-db.json
RESTORE=true
```

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