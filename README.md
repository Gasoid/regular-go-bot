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

- TOKEN - telegram token
- GIST_LOGS_URL - gist url
- OWM_API_KEY - api key for weather command



## Metrics
Prometheus metrics are exposed on 8080 port: http://localhost:8080/metrics

```bash
curl http://localhost:8080/metrics
```

## Health endpoint
http://localhost:8080/health


## License
This program is published under the terms of the MIT License. Please check the LICENSE file for more details.
