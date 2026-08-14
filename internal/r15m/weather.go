package r15m

import (
	"context"
	"fmt"

	"github.com/twossh/minitela/internal/weather"
)

type WeatherSyncer struct {
	conn *Connection

	cache *RegisterCache

	client *weather.Client

	city string
}

type WeatherSnapshot struct {
	City string

	TodayTemp string

	TodayIcon uint32

	TomorrowTemp string

	TomorrowIcon uint32

	TomorrowDate string

	ThirdTemp string

	ThirdIcon uint32

	ThirdDate string

	Writes int
}

func NewWeatherSyncer(
	conn *Connection,
	city string,
	apiKey string,
) *WeatherSyncer {
	return &WeatherSyncer{
		conn: conn,

		cache: NewRegisterCache(),

		client: weather.NewClient(
			apiKey,
		),

		city: city,
	}
}

func (s *WeatherSyncer) Sync(
	ctx context.Context,
) (*WeatherSnapshot, error) {
	if s == nil ||
		s.conn == nil ||
		s.cache == nil ||
		s.client == nil {
		return nil, fmt.Errorf(
			"sincronizador de clima inválido",
		)
	}

	forecast, err :=
		s.client.GetForecast(
			ctx,
			s.city,
		)

	if err != nil {
		return nil, err
	}

	if len(forecast.Days) < 3 {
		return nil, fmt.Errorf(
			"previsão possui menos de 3 dias",
		)
	}

	today :=
		forecast.Days[0]

	tomorrow :=
		forecast.Days[1]

	third :=
		forecast.Days[2]

	result :=
		&WeatherSnapshot{
			City: weather.CityText(
				forecast.Location.Name,
			),

			TodayTemp: weather.TemperatureText(
				today.MinTemp,
				today.MaxTemp,
			),

			TodayIcon: weather.R15MIcon(
				today.Condition,
			),

			TomorrowTemp: weather.TemperatureText(
				tomorrow.MinTemp,
				tomorrow.MaxTemp,
			),

			TomorrowIcon: weather.R15MIcon(
				tomorrow.Condition,
			),

			TomorrowDate: weather.DayText(
				tomorrow.Date,
			),

			ThirdTemp: weather.TemperatureText(
				third.MinTemp,
				third.MaxTemp,
			),

			ThirdIcon: weather.R15MIcon(
				third.Condition,
			),

			ThirdDate: weather.DayText(
				third.Date,
			),
		}

	if err := s.writeString(
		result,
		RegisterWeatherCity,
		result.City,
	); err != nil {
		return nil, err
	}

	if err := s.writeNum(
		result,
		RegisterWeatherTodayIcon,
		result.TodayIcon,
	); err != nil {
		return nil, err
	}

	if err := s.writeString(
		result,
		RegisterWeatherTodayTemp,
		result.TodayTemp,
	); err != nil {
		return nil, err
	}

	if err := s.writeNum(
		result,
		RegisterWeatherTomorrowIcon,
		result.TomorrowIcon,
	); err != nil {
		return nil, err
	}

	if err := s.writeString(
		result,
		RegisterWeatherTomorrowTemp,
		result.TomorrowTemp,
	); err != nil {
		return nil, err
	}

	if err := s.writeString(
		result,
		RegisterWeatherTomorrowDate,
		result.TomorrowDate,
	); err != nil {
		return nil, err
	}

	if err := s.writeNum(
		result,
		RegisterWeatherThirdIcon,
		result.ThirdIcon,
	); err != nil {
		return nil, err
	}

	if err := s.writeString(
		result,
		RegisterWeatherThirdTemp,
		result.ThirdTemp,
	); err != nil {
		return nil, err
	}

	if err := s.writeString(
		result,
		RegisterWeatherThirdDate,
		result.ThirdDate,
	); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *WeatherSyncer) writeString(
	result *WeatherSnapshot,
	regID uint16,
	value string,
) error {
	changed, err :=
		s.cache.WriteStringIfChanged(
			s.conn,
			regID,
			value,
		)

	if err != nil {
		return fmt.Errorf(
			"registrador %d: %w",
			regID,
			err,
		)
	}

	if changed {
		result.Writes++
	}

	return nil
}

func (s *WeatherSyncer) writeNum(
	result *WeatherSnapshot,
	regID uint16,
	value uint32,
) error {
	changed, err :=
		s.cache.WriteNumIfChanged(
			s.conn,
			regID,
			value,
		)

	if err != nil {
		return fmt.Errorf(
			"registrador %d: %w",
			regID,
			err,
		)
	}

	if changed {
		result.Writes++
	}

	return nil
}
