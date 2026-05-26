package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWeatherHandler(t *testing.T) {
	testCases := []struct {
		testName              string
		cepQueryParam         string
		viaCEPResponseBody    string
		weatherAPIResponseBody string
		expectedStatusCode    int
		expectedBody          string
		checkExactJSON        bool
	}{
		{
			testName:              "valid CEP returns 200 with correct temperatures",
			cepQueryParam:         "01310100",
			viaCEPResponseBody:    `{"localidade":"São Paulo","erro":false}`,
			weatherAPIResponseBody: `{"current":{"temp_c":28.5}}`,
			expectedStatusCode:    http.StatusOK,
			expectedBody:          `{"temp_C":28.5,"temp_F":83.3,"temp_K":301.5}`,
			checkExactJSON:        true,
		},
		{
			testName:           "CEP with 7 digits returns 422",
			cepQueryParam:      "0131010",
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedBody:       "invalid zipcode",
			checkExactJSON:     false,
		},
		{
			testName:           "CEP with letters returns 422",
			cepQueryParam:      "0131010A",
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedBody:       "invalid zipcode",
			checkExactJSON:     false,
		},
		{
			testName:           "ViaCEP returns erro true (bool) results in 404",
			cepQueryParam:      "00000000",
			viaCEPResponseBody: `{"erro":true}`,
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "can not find zipcode",
			checkExactJSON:     false,
		},
		{
			testName:           "ViaCEP returns erro true (string) results in 404",
			cepQueryParam:      "00000000",
			viaCEPResponseBody: `{"erro":"true"}`,
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "can not find zipcode",
			checkExactJSON:     false,
		},
		{
			testName:              "temperature conversion C=0 returns F=32 K=273",
			cepQueryParam:         "01310100",
			viaCEPResponseBody:    `{"localidade":"São Paulo","erro":false}`,
			weatherAPIResponseBody: `{"current":{"temp_c":0}}`,
			expectedStatusCode:    http.StatusOK,
			expectedBody:          `{"temp_C":0,"temp_F":32,"temp_K":273}`,
			checkExactJSON:        true,
		},
		{
			testName:              "temperature conversion C=100 returns F=212 K=373",
			cepQueryParam:         "01310100",
			viaCEPResponseBody:    `{"localidade":"São Paulo","erro":false}`,
			weatherAPIResponseBody: `{"current":{"temp_c":100}}`,
			expectedStatusCode:    http.StatusOK,
			expectedBody:          `{"temp_C":100,"temp_F":212,"temp_K":373}`,
			checkExactJSON:        true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			mockViaCEPServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				responseWriter.Header().Set("Content-Type", "application/json")
				responseWriter.Write([]byte(testCase.viaCEPResponseBody))
			}))
			defer mockViaCEPServer.Close()

			mockWeatherAPIServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				responseWriter.Header().Set("Content-Type", "application/json")
				responseWriter.Write([]byte(testCase.weatherAPIResponseBody))
			}))
			defer mockWeatherAPIServer.Close()

			weatherHandler := NewWeatherHandler(mockViaCEPServer.URL, mockWeatherAPIServer.URL)

			requestRecorder := httptest.NewRecorder()
			httpRequest := httptest.NewRequest(http.MethodGet, "/weather?cep="+testCase.cepQueryParam, nil)

			weatherHandler.ServeHTTP(requestRecorder, httpRequest)

			if requestRecorder.Code != testCase.expectedStatusCode {
				t.Errorf("expected status code %d, got %d", testCase.expectedStatusCode, requestRecorder.Code)
			}

			responseBodyString := strings.TrimSpace(requestRecorder.Body.String())

			if testCase.checkExactJSON {
				var actualTemperatureResponse TemperatureResponse
				var expectedTemperatureResponse TemperatureResponse

				json.Unmarshal([]byte(responseBodyString), &actualTemperatureResponse)
				json.Unmarshal([]byte(testCase.expectedBody), &expectedTemperatureResponse)

				if actualTemperatureResponse.TempC != expectedTemperatureResponse.TempC {
					t.Errorf("expected temp_C %f, got %f", expectedTemperatureResponse.TempC, actualTemperatureResponse.TempC)
				}
				if actualTemperatureResponse.TempF != expectedTemperatureResponse.TempF {
					t.Errorf("expected temp_F %f, got %f", expectedTemperatureResponse.TempF, actualTemperatureResponse.TempF)
				}
				if actualTemperatureResponse.TempK != expectedTemperatureResponse.TempK {
					t.Errorf("expected temp_K %f, got %f", expectedTemperatureResponse.TempK, actualTemperatureResponse.TempK)
				}
			} else {
				if !strings.Contains(responseBodyString, testCase.expectedBody) {
					t.Errorf("expected body to contain %q, got %q", testCase.expectedBody, responseBodyString)
				}
			}
		})
	}
}
