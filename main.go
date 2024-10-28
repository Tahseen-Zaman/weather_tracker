package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type apiConfigData struct {
	OpenWeatherMapApiKey string `json:"OpenWeatherMapApiKey"`
}

type weatherData struct {
	Name string `json:"name"`
	Main struct {
		Kelvin   float64 `json:"temp"`
		Humidity int     `json:"humidity"`
	} `json:"main"`
}

// Convert Kelvin to Celsius
func (wd weatherData) TemperatureCelsius() float64 {
	return wd.Main.Kelvin - 273.15
}

func loadApiConfig(filename string) (apiConfigData, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return apiConfigData{}, err
	}
	apiConfig := apiConfigData{}
	err = json.Unmarshal(bytes, &apiConfig)
	if err != nil {
		return apiConfigData{}, err
	}
	return apiConfig, nil
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from our weather tracker App!"))
}

func query(city string) (weatherData, error) {
	apiConfig, err := loadApiConfig(".apiConfig")
	if err != nil {
		return weatherData{}, err
	}

	resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?APPID=" + apiConfig.OpenWeatherMapApiKey + "&q=" + city)
	if err != nil {
		return weatherData{}, err
	}
	defer resp.Body.Close()

	var d weatherData
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return weatherData{}, err
	}

	return d, nil
}

func main() {
	http.HandleFunc("/hello", hello)

	http.HandleFunc("/weather/", func(w http.ResponseWriter, r *http.Request) {
		city := strings.SplitN(r.URL.Path, "/", 3)[2]
		data, err := query(city)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Format temperature to 2 decimal places and add % to humidity
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(struct {
			City      string `json:"city"`
			TempC     string `json:"temperature_celsius"`
			Humidity  string `json:"humidity"`
		}{
			City:     data.Name,
			TempC:    fmt.Sprintf("%.2f", data.TemperatureCelsius()),
			Humidity: fmt.Sprintf("%d%%", data.Main.Humidity),
		})
	})

	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
