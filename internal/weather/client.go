package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	baseURL = "https://api.weatherapi.com/v1"
)

// Client handles communication with WeatherAPI.com
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new WeatherAPI client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// doRequest performs an HTTP request to WeatherAPI
func (c *Client) doRequest(ctx context.Context, endpoint string, params url.Values) ([]byte, error) {
	params.Set("key", c.apiKey)

	fullURL := fmt.Sprintf("%s/%s?%s", baseURL, endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetCurrentWeather fetches current weather for a location
func (c *Client) GetCurrentWeather(ctx context.Context, location string) (*CurrentWeather, error) {
	params := url.Values{}
	params.Set("q", location)
	params.Set("aqi", "no") // Air quality index not needed

	body, err := c.doRequest(ctx, "current.json", params)
	if err != nil {
		return nil, err
	}

	var result CurrentWeatherResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result.Current, nil
}

// GetHistoricalWeather fetches historical weather for a specific date
// Date format: YYYY-MM-DD
func (c *Client) GetHistoricalWeather(ctx context.Context, location, date string) (*HistoricalWeather, error) {
	params := url.Values{}
	params.Set("q", location)
	params.Set("dt", date)

	body, err := c.doRequest(ctx, "history.json", params)
	if err != nil {
		return nil, err
	}

	var result HistoricalWeatherResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Forecast.ForecastDay) == 0 {
		return nil, fmt.Errorf("no historical data available for date: %s", date)
	}

	return &HistoricalWeather{
		Location: result.Location,
		Day:      result.Forecast.ForecastDay[0].Day,
		Date:     result.Forecast.ForecastDay[0].Date,
	}, nil
}

// GetForecastWeather fetches weather forecast for upcoming days
// days: number of forecast days (1-10)
func (c *Client) GetForecastWeather(ctx context.Context, location string, days int) (*ForecastWeather, error) {
	params := url.Values{}
	params.Set("q", location)
	params.Set("days", fmt.Sprintf("%d", days))
	params.Set("aqi", "no")
	params.Set("alerts", "no")

	body, err := c.doRequest(ctx, "forecast.json", params)
	if err != nil {
		return nil, err
	}

	var result ForecastWeatherResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &ForecastWeather{
		Location: result.Location,
		Current:  result.Current,
		Forecast: result.Forecast,
	}, nil
}

// GetWeatherByCoordinates fetches current weather by lat/lon
func (c *Client) GetWeatherByCoordinates(ctx context.Context, lat, lon float64) (*CurrentWeather, error) {
	location := fmt.Sprintf("%.4f,%.4f", lat, lon)
	return c.GetCurrentWeather(ctx, location)
}

// GetHistoricalWeatherByCoordinates fetches historical weather by lat/lon
func (c *Client) GetHistoricalWeatherByCoordinates(ctx context.Context, lat, lon float64, date string) (*HistoricalWeather, error) {
	location := fmt.Sprintf("%.4f,%.4f", lat, lon)
	return c.GetHistoricalWeather(ctx, location, date)
}