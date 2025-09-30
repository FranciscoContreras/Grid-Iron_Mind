package weather

// Location represents location information from WeatherAPI
type Location struct {
	Name      string  `json:"name"`
	Region    string  `json:"region"`
	Country   string  `json:"country"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	TzID      string  `json:"tz_id"`
	LocalTime string  `json:"localtime"`
}

// CurrentWeather represents current weather conditions
type CurrentWeather struct {
	TempF          float64   `json:"temp_f"`
	TempC          float64   `json:"temp_c"`
	IsDay          int       `json:"is_day"`
	Condition      Condition `json:"condition"`
	WindMPH        float64   `json:"wind_mph"`
	WindKPH        float64   `json:"wind_kph"`
	WindDegree     int       `json:"wind_degree"`
	WindDir        string    `json:"wind_dir"`
	PressureMB     float64   `json:"pressure_mb"`
	PressureIn     float64   `json:"pressure_in"`
	PrecipMM       float64   `json:"precip_mm"`
	PrecipIn       float64   `json:"precip_in"`
	Humidity       int       `json:"humidity"`
	Cloud          int       `json:"cloud"`
	FeelsLikeF     float64   `json:"feelslike_f"`
	FeelsLikeC     float64   `json:"feelslike_c"`
	VisKM          float64   `json:"vis_km"`
	VisMiles       float64   `json:"vis_miles"`
	UV             float64   `json:"uv"`
	GustMPH        float64   `json:"gust_mph"`
	GustKPH        float64   `json:"gust_kph"`
}

// Condition represents weather condition
type Condition struct {
	Text string `json:"text"`
	Icon string `json:"icon"`
	Code int    `json:"code"`
}

// Day represents daily weather statistics
type Day struct {
	MaxTempF          float64   `json:"maxtemp_f"`
	MaxTempC          float64   `json:"maxtemp_c"`
	MinTempF          float64   `json:"mintemp_f"`
	MinTempC          float64   `json:"mintemp_c"`
	AvgTempF          float64   `json:"avgtemp_f"`
	AvgTempC          float64   `json:"avgtemp_c"`
	MaxWindMPH        float64   `json:"maxwind_mph"`
	MaxWindKPH        float64   `json:"maxwind_kph"`
	TotalPrecipMM     float64   `json:"totalprecip_mm"`
	TotalPrecipIn     float64   `json:"totalprecip_in"`
	AvgVisKM          float64   `json:"avgvis_km"`
	AvgVisMiles       float64   `json:"avgvis_miles"`
	AvgHumidity       float64   `json:"avghumidity"`
	DailyChanceOfRain int       `json:"daily_chance_of_rain"`
	DailyChanceOfSnow int       `json:"daily_chance_of_snow"`
	Condition         Condition `json:"condition"`
	UV                float64   `json:"uv"`
}

// ForecastDay represents a single forecast day
type ForecastDay struct {
	Date      string `json:"date"`
	DateEpoch int64  `json:"date_epoch"`
	Day       Day    `json:"day"`
}

// Forecast represents forecast data
type Forecast struct {
	ForecastDay []ForecastDay `json:"forecastday"`
}

// CurrentWeatherResponse is the API response for current weather
type CurrentWeatherResponse struct {
	Location Location       `json:"location"`
	Current  CurrentWeather `json:"current"`
}

// HistoricalWeather represents historical weather data
type HistoricalWeather struct {
	Location Location `json:"location"`
	Day      Day      `json:"day"`
	Date     string   `json:"date"`
}

// HistoricalWeatherResponse is the API response for historical weather
type HistoricalWeatherResponse struct {
	Location Location `json:"location"`
	Forecast Forecast `json:"forecast"`
}

// ForecastWeather represents forecast weather data
type ForecastWeather struct {
	Location Location       `json:"location"`
	Current  CurrentWeather `json:"current"`
	Forecast Forecast       `json:"forecast"`
}

// ForecastWeatherResponse is the API response for forecast weather
type ForecastWeatherResponse struct {
	Location Location       `json:"location"`
	Current  CurrentWeather `json:"current"`
	Forecast Forecast       `json:"forecast"`
}