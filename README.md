# Cloud Run - Brazilian CEP Weather Service

An HTTP REST API that takes a Brazilian CEP (postal code), looks up the city via ViaCEP, and returns the current temperature in Celsius, Fahrenheit, and Kelvin using WeatherAPI.

## Requirements

- A WeatherAPI key (free at https://www.weatherapi.com)
- Docker (for containerized runs)
- Go 1.22+ (for local runs)

## API

### GET /weather?cep={8-digit-CEP}

**Success (200):**
```json
{"temp_C": 28.5, "temp_F": 83.3, "temp_K": 301.5}
```

**Invalid CEP format (422):**
```
invalid zipcode
```

**CEP not found (404):**
```
can not find zipcode
```

## Running with Docker

Build the image:
```bash
docker build -t weather-service .
```

Run with your WeatherAPI key:
```bash
docker run -p 8080:8080 -e WEATHER_API_KEY=your_key_here weather-service
```

## Running Tests

```bash
go test ./...
```

## Live URL

The service is deployed and accessible at:

```
https://fc-go-cloud-run-453977528449.us-central1.run.app/weather?cep=01310100
```

## Cloud Run Deployment

After building and pushing the image to Google Container Registry, deploy with:
```bash
gcloud run deploy weather-service \
  --image gcr.io/YOUR_PROJECT/weather-service \
  --platform managed \
  --region us-central1 \
  --set-env-vars WEATHER_API_KEY=your_key_here \
  --allow-unauthenticated
```
