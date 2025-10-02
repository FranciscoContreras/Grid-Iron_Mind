package handlers

import (
	"net/http"
	"strconv"

	"github.com/francisco/gridironmind/internal/weather"
	"github.com/francisco/gridironmind/pkg/response"
)

type WeatherHandler struct {
	weatherClient *weather.Client
}

func NewWeatherHandler(weatherClient *weather.Client) *WeatherHandler {
	return &WeatherHandler{
		weatherClient: weatherClient,
	}
}

// HandleCurrentWeather gets current weather for a location
// GET /api/v1/weather/current?location=Kansas+City,MO
// GET /api/v1/weather/current?lat=39.0489&lon=-94.4839
func (h *WeatherHandler) HandleCurrentWeather(w http.ResponseWriter, r *http.Request) {
	location := r.URL.Query().Get("location")
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")

	var weatherData *weather.CurrentWeather
	var err error

	if latStr != "" && lonStr != "" {
		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "INVALID_PARAMETER", "Invalid latitude value")
			return
		}

		lon, err := strconv.ParseFloat(lonStr, 64)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "INVALID_PARAMETER", "Invalid longitude value")
			return
		}

		weatherData, err = h.weatherClient.GetWeatherByCoordinates(r.Context(), lat, lon)
	} else if location != "" {
		weatherData, err = h.weatherClient.GetCurrentWeather(r.Context(), location)
	} else {
		response.Error(w, http.StatusBadRequest, "MISSING_PARAMETER", "Either 'location' or 'lat'+'lon' parameters are required")
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "WEATHER_API_ERROR", err.Error())
		return
	}

	response.Success(w, weatherData)
}

// HandleHistoricalWeather gets historical weather for a specific date
// GET /api/v1/weather/historical?location=Kansas+City,MO&date=2024-12-21
// GET /api/v1/weather/historical?lat=39.0489&lon=-94.4839&date=2024-12-21
func (h *WeatherHandler) HandleHistoricalWeather(w http.ResponseWriter, r *http.Request) {
	location := r.URL.Query().Get("location")
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")
	date := r.URL.Query().Get("date")

	if date == "" {
		response.Error(w, http.StatusBadRequest, "MISSING_PARAMETER", "'date' parameter is required (format: YYYY-MM-DD)")
		return
	}

	var weatherData *weather.HistoricalWeather
	var err error

	if latStr != "" && lonStr != "" {
		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "INVALID_PARAMETER", "Invalid latitude value")
			return
		}

		lon, err := strconv.ParseFloat(lonStr, 64)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "INVALID_PARAMETER", "Invalid longitude value")
			return
		}

		weatherData, err = h.weatherClient.GetHistoricalWeatherByCoordinates(r.Context(), lat, lon, date)
	} else if location != "" {
		weatherData, err = h.weatherClient.GetHistoricalWeather(r.Context(), location, date)
	} else {
		response.Error(w, http.StatusBadRequest, "MISSING_PARAMETER", "Either 'location' or 'lat'+'lon' parameters are required")
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "WEATHER_API_ERROR", err.Error())
		return
	}

	response.Success(w, weatherData)
}

// HandleForecastWeather gets weather forecast for upcoming days
// GET /api/v1/weather/forecast?location=Kansas+City,MO&days=7
func (h *WeatherHandler) HandleForecastWeather(w http.ResponseWriter, r *http.Request) {
	location := r.URL.Query().Get("location")
	if location == "" {
		response.Error(w, http.StatusBadRequest, "MISSING_PARAMETER", "'location' parameter is required")
		return
	}

	daysStr := r.URL.Query().Get("days")
	days := 7 // default to 7 days
	if daysStr != "" {
		parsedDays, err := strconv.Atoi(daysStr)
		if err != nil || parsedDays < 1 || parsedDays > 10 {
			response.Error(w, http.StatusBadRequest, "INVALID_PARAMETER", "days must be between 1 and 10")
			return
		}
		days = parsedDays
	}

	weatherData, err := h.weatherClient.GetForecastWeather(r.Context(), location, days)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "WEATHER_API_ERROR", err.Error())
		return
	}

	response.Success(w, weatherData)
}