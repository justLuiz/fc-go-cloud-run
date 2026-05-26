package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"regexp"
)

type FlexibleBool bool

func (flexibleBool *FlexibleBool) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case `true`, `"true"`:
		*flexibleBool = true
	default:
		*flexibleBool = false
	}
	return nil
}

type ViaCEPResponse struct {
	Localidade string      `json:"localidade"`
	Erro       FlexibleBool `json:"erro"`
}

type WeatherAPIResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

type TemperatureResponse struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

type WeatherHandler struct {
	viaCEPBaseURL    string
	weatherAPIBaseURL string
}

func NewWeatherHandler(viaCEPBaseURL string, weatherAPIBaseURL string) *WeatherHandler {
	return &WeatherHandler{
		viaCEPBaseURL:    viaCEPBaseURL,
		weatherAPIBaseURL: weatherAPIBaseURL,
	}
}

func (weatherHandler *WeatherHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	cepValue := request.URL.Query().Get("cep")

	cepPattern := regexp.MustCompile(`^\d{8}$`)
	if !cepPattern.MatchString(cepValue) {
		http.Error(responseWriter, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	viaCEPURL := fmt.Sprintf("%s/ws/%s/json/", weatherHandler.viaCEPBaseURL, cepValue)
	viaCEPHTTPResponse, viaCEPError := http.Get(viaCEPURL)
	if viaCEPError != nil {
		http.Error(responseWriter, "failed to fetch CEP data", http.StatusInternalServerError)
		return
	}
	defer viaCEPHTTPResponse.Body.Close()

	var viaCEPResponse ViaCEPResponse
	viaCEPDecodeError := json.NewDecoder(viaCEPHTTPResponse.Body).Decode(&viaCEPResponse)
	if viaCEPDecodeError != nil {
		http.Error(responseWriter, "failed to decode CEP response", http.StatusInternalServerError)
		return
	}

	if viaCEPResponse.Erro || viaCEPResponse.Localidade == "" {
		http.Error(responseWriter, "can not find zipcode", http.StatusNotFound)
		return
	}

	cityName := viaCEPResponse.Localidade
	encodedCityName := url.QueryEscape(cityName)
	weatherAPIKey := os.Getenv("WEATHER_API_KEY")

	weatherAPIURL := fmt.Sprintf("%s/v1/current.json?key=%s&q=%s&aqi=no", weatherHandler.weatherAPIBaseURL, weatherAPIKey, encodedCityName)
	weatherHTTPResponse, weatherError := http.Get(weatherAPIURL)
	if weatherError != nil {
		http.Error(responseWriter, "failed to fetch weather data", http.StatusInternalServerError)
		return
	}
	defer weatherHTTPResponse.Body.Close()

	var weatherResponse WeatherAPIResponse
	weatherDecodeError := json.NewDecoder(weatherHTTPResponse.Body).Decode(&weatherResponse)
	if weatherDecodeError != nil {
		http.Error(responseWriter, "failed to decode weather response", http.StatusInternalServerError)
		return
	}

	tempCelsius := weatherResponse.Current.TempC
	tempFahrenheit := math.Round((tempCelsius*1.8+32)*10) / 10
	tempKelvin := math.Round((tempCelsius+273)*10) / 10

	temperatureResponse := TemperatureResponse{
		TempC: tempCelsius,
		TempF: tempFahrenheit,
		TempK: tempKelvin,
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusOK)
	json.NewEncoder(responseWriter).Encode(temperatureResponse)
}

func main() {
	viaCEPBaseURL := os.Getenv("VIA_CEP_BASE_URL")
	if viaCEPBaseURL == "" {
		viaCEPBaseURL = "https://viacep.com.br"
	}

	weatherAPIBaseURL := os.Getenv("WEATHER_API_BASE_URL")
	if weatherAPIBaseURL == "" {
		weatherAPIBaseURL = "http://api.weatherapi.com"
	}

	serverAddr := os.Getenv("SERVER_ADDR")
	if serverAddr == "" {
		serverAddr = ":8080"
	}

	weatherHandler := NewWeatherHandler(viaCEPBaseURL, weatherAPIBaseURL)

	mux := http.NewServeMux()
	mux.Handle("/weather", weatherHandler)

	http.ListenAndServe(serverAddr, mux)
}
