package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultForecastURL = "https://api.weatherapi.com/v1/forecast.json"

type Client struct {
	apiKey      string
	httpClient  *http.Client
	forecastURL string
}

type Location struct {
	Name    string
	Region  string
	Country string
}

type Day struct {
	Date      time.Time
	Condition string
	MinTemp   int
	MaxTemp   int
}

type Forecast struct {
	Location Location
	Days     []Day
}

func NewClient(
	apiKey string,
) *Client {
	return &Client{
		apiKey: strings.TrimSpace(apiKey),

		httpClient: &http.Client{
			Timeout: 12 * time.Second,
		},

		forecastURL: defaultForecastURL,
	}
}

func (c *Client) GetForecast(
	ctx context.Context,
	city string,
) (*Forecast, error) {
	city =
		strings.TrimSpace(city)

	if city == "" {
		return nil, fmt.Errorf(
			"cidade não configurada",
		)
	}

	if c.apiKey == "" {
		return nil, fmt.Errorf(
			"chave WeatherAPI não configurada",
		)
	}

	values := url.Values{}

	values.Set(
		"key",
		c.apiKey,
	)

	values.Set(
		"q",
		city,
	)

	values.Set(
		"days",
		"3",
	)

	values.Set(
		"aqi",
		"no",
	)

	values.Set(
		"alerts",
		"no",
	)

	requestURL :=
		c.forecastURL +
			"?" +
			values.Encode()

	req, err :=
		http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			requestURL,
			nil,
		)

	if err != nil {
		return nil, fmt.Errorf(
			"criar consulta WeatherAPI: %w",
			err,
		)
	}

	req.Header.Set(
		"User-Agent",
		"MiniTela/0.1",
	)

	resp, err :=
		c.httpClient.Do(req)

	if err != nil {
		// Não propagamos o *url.Error porque ele pode
		// incluir a URL completa e consequentemente
		// expor a chave da API nos logs.
		return nil, fmt.Errorf(
			"falha de conexão com WeatherAPI",
		)
	}

	defer resp.Body.Close()

	if resp.StatusCode !=
		http.StatusOK {
		var apiError struct {
			Error struct {
				Code int `json:"code"`

				Message string `json:"message"`
			} `json:"error"`
		}

		_ = json.NewDecoder(
			resp.Body,
		).Decode(&apiError)

		if apiError.Error.Message != "" {
			return nil, fmt.Errorf(
				"WeatherAPI HTTP %d: %s",
				resp.StatusCode,
				apiError.Error.Message,
			)
		}

		return nil, fmt.Errorf(
			"WeatherAPI retornou HTTP %d",
			resp.StatusCode,
		)
	}

	var result struct {
		Location struct {
			Name string `json:"name"`

			Region string `json:"region"`

			Country string `json:"country"`
		} `json:"location"`

		Forecast struct {
			ForecastDay []struct {
				Date string `json:"date"`

				Day struct {
					MaxTempC float64 `json:"maxtemp_c"`

					MinTempC float64 `json:"mintemp_c"`

					Condition struct {
						Text string `json:"text"`
					} `json:"condition"`
				} `json:"day"`
			} `json:"forecastday"`
		} `json:"forecast"`
	}

	if err := json.NewDecoder(
		resp.Body,
	).Decode(&result); err != nil {
		return nil, fmt.Errorf(
			"decodificar resposta WeatherAPI: %w",
			err,
		)
	}

	if len(
		result.Forecast.ForecastDay,
	) < 3 {
		return nil, fmt.Errorf(
			"WeatherAPI retornou somente %d dias",
			len(
				result.Forecast.ForecastDay,
			),
		)
	}

	forecast := &Forecast{
		Location: Location{
			Name: result.Location.Name,

			Region: result.Location.Region,

			Country: result.Location.Country,
		},

		Days: make(
			[]Day,
			0,
			3,
		),
	}

	for i := 0; i < 3; i++ {
		item :=
			result.Forecast.
				ForecastDay[i]

		date, err := time.Parse(
			"2006-01-02",
			item.Date,
		)

		if err != nil {
			return nil, fmt.Errorf(
				"data inválida %q: %w",
				item.Date,
				err,
			)
		}

		forecast.Days =
			append(
				forecast.Days,
				Day{
					Date: date,

					Condition: strings.TrimSpace(
						item.Day.
							Condition.
							Text,
					),

					MinTemp: int(
						math.Round(
							item.Day.
								MinTempC,
						),
					),

					MaxTemp: int(
						math.Round(
							item.Day.
								MaxTempC,
						),
					),
				},
			)
	}

	return forecast, nil
}
