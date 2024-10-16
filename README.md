[![CI](https://github.com/Gasoid/regular-go-bot/actions/workflows/ci.yml/badge.svg)](https://github.com/Gasoid/regular-go-bot/actions/workflows/ci.yml)

## Command list

- help
- estimation
- changelog
- currency
- joke
- holiday
- weather
- chat_info
- random
- b64encode
- b64decode


## Settings / Env variables
You have to provide 3 env variables:

- BOT_TOKEN - telegram token
- GIST_LOGS_URL - gist url
- OWM_API_KEY - api key for weather command

## How it works

<img width="511" alt="Screenshot 2022-07-17 at 21 12 16" src="https://user-images.githubusercontent.com/833157/179421331-6f380348-994a-433f-8475-415134d8d169.png">



## Metrics
Prometheus metrics are exposed on 8080 port: http://localhost:8080/metrics

```bash
curl http://localhost:8080/metrics
```

## Health endpoint
url: http://localhost:8080/health

```bash
curl http://localhost:8080/health
```


## Compilation routine

```
go mod download
go build ./
```

or

```
docker build -t bot:test ./
```

## License
This program is published under the terms of the MIT License. Please check the LICENSE file for more details.
