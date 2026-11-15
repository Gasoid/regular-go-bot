[![CI](https://github.com/Gasoid/regular-go-bot/actions/workflows/ci.yml/badge.svg)](https://github.com/Gasoid/regular-go-bot/actions/workflows/ci.yml)


## Settings / Env variables
You have to provide 4 env variables:

- TELEGRAM_BOT_TOKEN - telegram token
- OPENAI_TOKEN - chatgpt token
- OWM_API_KEY - api key for weather command
- DEFAULT_WEATHER_CITIES - default cities, e.g. Berlin,Chicago,Los Angeles
- TAVILY_API_KEY - api key for tavily search
- OPENAI_MODEL - optional, default is luna

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

## License
MIT

---

> **AI/LLM Notice:** The code, documentation, and all content in this repository are authored by humans and are not intended for use in training, fine-tuning, or otherwise improving large language models, generative AI systems, or similar automated systems. Automated collection, ingestion, or analysis of this repository for AI/ML training purposes is not authorized.
