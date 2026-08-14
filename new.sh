cd ~/Aplicativos/MiniTela || exit 1

cat > /tmp/minitela-weatherapi-upgrade.sh <<'MINITELA_SCRIPT'
#!/usr/bin/env bash

set -Eeuo pipefail

PROJECT="$HOME/Aplicativos/MiniTela"
STATE="$HOME/.local/state/minitela"

cd "$PROJECT"

STAMP="$(date +%Y%m%d-%H%M%S)"
BACKUP="$STATE/backups/weatherapi-$STAMP"

mkdir -p "$BACKUP"
mkdir -p "$STATE"

echo "$BACKUP" > "$STATE/last-weather-backup"

trap 'echo; echo "ERRO: atualização interrompida."; echo "Backup disponível em: '"$BACKUP"'"; echo' ERR

echo
echo "===================================================="
echo " MiniTela - Integração WeatherAPI"
echo "===================================================="
echo

echo "[1/11] Criando backup..."

FILES=(
    "internal/config/config.go"
    "cmd/minitela-config/main.go"
    "internal/weather/client.go"
    "internal/weather/display.go"
    "internal/weather/display_test.go"
    "internal/r15m/registers.go"
    "internal/r15m/weather.go"
    "internal/app/runtime.go"
    "cmd/minitela/main.go"
)

for FILE in "${FILES[@]}"; do
    if [[ -f "$FILE" ]]; then
        mkdir -p "$BACKUP/$(dirname "$FILE")"
        cp -a "$FILE" "$BACKUP/$FILE"
    fi
done

echo "Backup criado em:"
echo "$BACKUP"

mkdir -p \
    internal/config \
    internal/weather \
    internal/r15m \
    internal/app \
    cmd/minitela \
    cmd/minitela-config \
    cmd/minitela-weather-test \
    bin

echo
echo "[2/11] Configuração persistente..."

cat > internal/config/config.go <<'EOF'
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const currentVersion = 1

type Config struct {
	Version int `json:"version"`

	LastScreen string `json:"last_screen"`
	City       string `json:"city"`

	WeatherAPIKey string `json:"weather_api_key,omitempty"`

	Brightness *int `json:"brightness,omitempty"`

	RestoreLastScreen bool `json:"restore_last_screen"`
	Autostart         bool `json:"autostart"`
	StartMinimized    bool `json:"start_minimized"`

	MonitorIntervalSeconds int `json:"monitor_interval_seconds"`
}

func Default() Config {
	return Config{
		Version:                currentVersion,
		LastScreen:             "monitor",
		City:                   "",
		WeatherAPIKey:          "",
		Brightness:             nil,
		RestoreLastScreen:      true,
		Autostart:              false,
		StartMinimized:         true,
		MonitorIntervalSeconds: 10,
	}
}

func Path() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf(
			"localizar diretório de configuração: %w",
			err,
		)
	}

	return filepath.Join(
		base,
		"minitela",
		"config.json",
	), nil
}

func Load() (Config, error) {
	path, err := Path()
	if err != nil {
		return Config{}, err
	}

	return LoadFrom(path)
}

func LoadFrom(path string) (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)

	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}

	if err != nil {
		return Config{}, fmt.Errorf(
			"ler configuração: %w",
			err,
		)
	}

	if err := json.Unmarshal(
		data,
		&cfg,
	); err != nil {
		return Config{}, fmt.Errorf(
			"configuração inválida: %w",
			err,
		)
	}

	normalize(&cfg)

	return cfg, nil
}

func Save(cfg Config) error {
	path, err := Path()
	if err != nil {
		return err
	}

	return SaveTo(path, cfg)
}

func SaveTo(
	path string,
	cfg Config,
) error {
	normalize(&cfg)

	dir := filepath.Dir(path)

	if err := os.MkdirAll(
		dir,
		0o700,
	); err != nil {
		return fmt.Errorf(
			"criar diretório de configuração: %w",
			err,
		)
	}

	data, err := json.MarshalIndent(
		cfg,
		"",
		"  ",
	)
	if err != nil {
		return fmt.Errorf(
			"codificar configuração: %w",
			err,
		)
	}

	data = append(data, '\n')

	tmp := path + ".tmp"

	if err := os.WriteFile(
		tmp,
		data,
		0o600,
	); err != nil {
		return fmt.Errorf(
			"gravar configuração temporária: %w",
			err,
		)
	}

	if err := os.Rename(
		tmp,
		path,
	); err != nil {
		_ = os.Remove(tmp)

		return fmt.Errorf(
			"salvar configuração: %w",
			err,
		)
	}

	return nil
}

func normalize(cfg *Config) {
	if cfg.Version <= 0 {
		cfg.Version = currentVersion
	}

	switch cfg.LastScreen {
	case "whatsapp",
		"notes",
		"monitor",
		"weather":
	default:
		cfg.LastScreen = "monitor"
	}

	cfg.City =
		strings.TrimSpace(cfg.City)

	cfg.WeatherAPIKey =
		strings.TrimSpace(cfg.WeatherAPIKey)

	if cfg.MonitorIntervalSeconds < 1 {
		cfg.MonitorIntervalSeconds = 10
	}

	if cfg.Brightness != nil {
		value := *cfg.Brightness

		if value < 0 {
			value = 0
		}

		if value > 100 {
			value = 100
		}

		cfg.Brightness = &value
	}
}
EOF

echo
echo "[3/11] Ferramenta de configuração..."

cat > cmd/minitela-config/main.go <<'EOF'
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/twossh/minitela/internal/config"
	"github.com/twossh/minitela/internal/r15m"
)

func main() {
	city := flag.String(
		"city",
		"",
		"cidade utilizada pelo clima",
	)

	screen := flag.String(
		"screen",
		"",
		"última tela: whatsapp, notes, monitor ou weather",
	)

	brightness := flag.Int(
		"brightness",
		-1,
		"brilho entre 0 e 100",
	)

	weatherKey := flag.String(
		"weather-key",
		"",
		"chave WeatherAPI",
	)

	weatherKeyStdin := flag.Bool(
		"weather-key-stdin",
		false,
		"lê a chave WeatherAPI pela entrada padrão",
	)

	clearWeatherKey := flag.Bool(
		"clear-weather-key",
		false,
		"remove a chave WeatherAPI salva",
	)

	autostart := flag.String(
		"autostart",
		"",
		"on ou off",
	)

	minimized := flag.String(
		"minimized",
		"",
		"on ou off",
	)

	restore := flag.String(
		"restore",
		"",
		"on ou off",
	)

	show := flag.Bool(
		"show",
		false,
		"exibe a configuração",
	)

	flag.Parse()

	if *weatherKey != "" &&
		*weatherKeyStdin {
		fail(fmt.Errorf(
			"use --weather-key ou --weather-key-stdin, não os dois",
		))
	}

	cfg, err := config.Load()
	if err != nil {
		fail(err)
	}

	changed := false

	if *city != "" {
		cfg.City =
			strings.TrimSpace(*city)

		changed = true
	}

	if *screen != "" {
		s, err :=
			r15m.ParseScreen(*screen)

		if err != nil {
			fail(err)
		}

		cfg.LastScreen =
			screenConfigName(s)

		changed = true
	}

	if *brightness >= 0 {
		if *brightness > 100 {
			fail(fmt.Errorf(
				"brilho deve estar entre 0 e 100",
			))
		}

		value := *brightness

		cfg.Brightness = &value

		changed = true
	}

	if *weatherKey != "" {
		cfg.WeatherAPIKey =
			strings.TrimSpace(*weatherKey)

		changed = true
	}

	if *weatherKeyStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fail(fmt.Errorf(
				"ler chave WeatherAPI: %w",
				err,
			))
		}

		value :=
			strings.TrimSpace(
				string(data),
			)

		if value == "" {
			fail(fmt.Errorf(
				"chave WeatherAPI vazia",
			))
		}

		cfg.WeatherAPIKey = value

		changed = true
	}

	if *clearWeatherKey {
		cfg.WeatherAPIKey = ""
		changed = true
	}

	if *autostart != "" {
		value, err :=
			parseBool(*autostart)

		if err != nil {
			fail(err)
		}

		cfg.Autostart = value

		changed = true
	}

	if *minimized != "" {
		value, err :=
			parseBool(*minimized)

		if err != nil {
			fail(err)
		}

		cfg.StartMinimized = value

		changed = true
	}

	if *restore != "" {
		value, err :=
			parseBool(*restore)

		if err != nil {
			fail(err)
		}

		cfg.RestoreLastScreen = value

		changed = true
	}

	if changed {
		if err := config.Save(cfg); err != nil {
			fail(err)
		}
	}

	if changed || *show {
		showConfig(cfg)
	}
}

func showConfig(cfg config.Config) {
	path, _ := config.Path()

	fmt.Printf(
		"Arquivo      : %s\n",
		path,
	)

	fmt.Printf(
		"Última tela  : %s\n",
		cfg.LastScreen,
	)

	fmt.Printf(
		"Cidade       : %s\n",
		cfg.City,
	)

	if cfg.Brightness == nil {
		fmt.Println(
			"Brilho       : automático/não definido",
		)
	} else {
		fmt.Printf(
			"Brilho       : %d%%\n",
			*cfg.Brightness,
		)
	}

	if cfg.WeatherAPIKey == "" {
		fmt.Println(
			"WeatherAPI   : não configurada",
		)
	} else {
		fmt.Println(
			"WeatherAPI   : configurada",
		)
	}

	fmt.Printf(
		"Restaurar    : %t\n",
		cfg.RestoreLastScreen,
	)

	fmt.Printf(
		"Autostart    : %t\n",
		cfg.Autostart,
	)

	fmt.Printf(
		"Minimizado   : %t\n",
		cfg.StartMinimized,
	)

	fmt.Printf(
		"Intervalo    : %ds\n",
		cfg.MonitorIntervalSeconds,
	)
}

func screenConfigName(
	screen r15m.Screen,
) string {
	switch screen {
	case r15m.ScreenWhatsApp:
		return "whatsapp"

	case r15m.ScreenNotes:
		return "notes"

	case r15m.ScreenMonitor:
		return "monitor"

	case r15m.ScreenWeather:
		return "weather"

	default:
		return "monitor"
	}
}

func parseBool(
	value string,
) (bool, error) {
	switch strings.ToLower(
		strings.TrimSpace(value),
	) {
	case "on",
		"true",
		"1",
		"yes",
		"sim":
		return true, nil

	case "off",
		"false",
		"0",
		"no",
		"nao",
		"não":
		return false, nil
	}

	return false, fmt.Errorf(
		"valor inválido %q; use on ou off",
		value,
	)
}

func fail(err error) {
	fmt.Fprintf(
		os.Stderr,
		"Erro: %v\n",
		err,
	)

	os.Exit(1)
}
EOF

echo
echo "[4/11] Cliente WeatherAPI..."

cat > internal/weather/client.go <<'EOF'
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

const defaultForecastURL =
	"https://api.weatherapi.com/v1/forecast.json"

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
		apiKey:
			strings.TrimSpace(apiKey),

		httpClient: &http.Client{
			Timeout: 12 * time.Second,
		},

		forecastURL:
			defaultForecastURL,
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
			Name:
				result.Location.Name,

			Region:
				result.Location.Region,

			Country:
				result.Location.Country,
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

					Condition:
						strings.TrimSpace(
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
EOF

echo
echo "[5/11] Conversão das condições meteorológicas..."

cat > internal/weather/display.go <<'EOF'
package weather

import (
	"fmt"
	"strings"
	"time"
)

func R15MIcon(
	condition string,
) uint32 {
	value :=
		strings.ToLower(
			strings.TrimSpace(
				condition,
			),
		)

	switch {
	case strings.Contains(
		value,
		"sunny",
	),
		strings.Contains(
			value,
			"clear",
		):
		return 0

	case strings.Contains(
		value,
		"partly cloudy",
	):
		return 1

	case strings.Contains(
		value,
		"snow",
	),
		strings.Contains(
			value,
			"ice",
		),
		strings.Contains(
			value,
			"blizzard",
		),
		strings.Contains(
			value,
			"freezing",
		),
		strings.Contains(
			value,
			"sleet",
		):
		return 6

	case strings.Contains(
		value,
		"rain",
	),
		strings.Contains(
			value,
			"drizzle",
		),
		strings.Contains(
			value,
			"thunder",
		),
		strings.Contains(
			value,
			"shower",
		):
		return 3

	case strings.Contains(
		value,
		"cloudy",
	),
		strings.Contains(
			value,
			"overcast",
		),
		strings.Contains(
			value,
			"fog",
		),
		strings.Contains(
			value,
			"mist",
		):
		return 2

	default:
		return 2
	}
}

func TemperatureText(
	minTemp,
	maxTemp int,
) string {
	return fmt.Sprintf(
		"%d°/%d°",
		minTemp,
		maxTemp,
	)
}

func DayText(
	date time.Time,
) string {
	weekday :=
		map[time.Weekday]string{
			time.Sunday:
				"DOM",

			time.Monday:
				"SEG",

			time.Tuesday:
				"TER",

			time.Wednesday:
				"QUA",

			time.Thursday:
				"QUI",

			time.Friday:
				"SEX",

			time.Saturday:
				"SÁB",
		}

	return fmt.Sprintf(
		"%s%02d",
		weekday[
			date.Weekday()
		],
		date.Day(),
	)
}

func CityText(
	value string,
) string {
	value =
		strings.TrimSpace(value)

	runes := []rune(value)

	if len(runes) <= 10 {
		return value
	}

	return string(
		runes[:8],
	) + "..."
}
EOF

cat > internal/weather/display_test.go <<'EOF'
package weather

import (
	"testing"
	"time"
)

func TestR15MIcon(
	t *testing.T,
) {
	tests := []struct {
		condition string
		want      uint32
	}{
		{"Sunny", 0},
		{"Clear", 0},
		{"Partly cloudy", 1},
		{"Cloudy", 2},
		{"Overcast", 2},
		{"Mist", 2},
		{"Patchy rain nearby", 3},
		{"Light drizzle", 3},
		{"Thundery outbreaks", 3},
		{"Light snow", 6},
		{"Blizzard", 6},
	}

	for _, tt :=
		range tests {
		got :=
			R15MIcon(
				tt.condition,
			)

		if got != tt.want {
			t.Fatalf(
				"R15MIcon(%q)=%d esperado=%d",
				tt.condition,
				got,
				tt.want,
			)
		}
	}
}

func TestTemperatureText(
	t *testing.T,
) {
	got :=
		TemperatureText(
			17,
			25,
		)

	if got != "17°/25°" {
		t.Fatalf(
			"TemperatureText=%q",
			got,
		)
	}
}

func TestDayText(
	t *testing.T,
) {
	date :=
		time.Date(
			2026,
			time.August,
			14,
			0,
			0,
			0,
			0,
			time.UTC,
		)

	got :=
		DayText(date)

	if got != "SEX14" {
		t.Fatalf(
			"DayText=%q",
			got,
		)
	}
}

func TestCityText(
	t *testing.T,
) {
	got :=
		CityText(
			"Porto Alegre",
		)

	if got != "Porto Al..." {
		t.Fatalf(
			"CityText=%q",
			got,
		)
	}
}
EOF

echo
echo "[6/11] Registradores do R15M..."

cat > internal/r15m/registers.go <<'EOF'
package r15m

const (
	RegisterCurrentPage uint16 = 2
	RegisterBrightness  uint16 = 7

	// Monitor.
	RegisterBatteryText uint16 = 1082

	RegisterWiFiSSID uint16 = 1083

	RegisterWiFiQuality uint16 = 1084

	RegisterBluetoothName uint16 = 1085

	RegisterBatteryLevel uint16 = 1150

	// Clima.
	RegisterWeatherTodayIcon uint16 = 1110

	RegisterWeatherTomorrowIcon uint16 = 1115

	RegisterWeatherTomorrowDate uint16 = 1119

	RegisterWeatherThirdIcon uint16 = 1120

	RegisterWeatherThirdDate uint16 = 1124

	RegisterWeatherCity uint16 = 2027

	RegisterWeatherTodayTemp uint16 = 2030

	RegisterWeatherTomorrowTemp uint16 = 2031

	RegisterWeatherThirdTemp uint16 = 2032
)
EOF

echo
echo "[7/11] Sincronizador WeatherAPI -> R15M..."

cat > internal/r15m/weather.go <<'EOF'
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

		cache:
			NewRegisterCache(),

		client:
			weather.NewClient(
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
			City:
				weather.CityText(
					forecast.Location.Name,
				),

			TodayTemp:
				weather.TemperatureText(
					today.MinTemp,
					today.MaxTemp,
				),

			TodayIcon:
				weather.R15MIcon(
					today.Condition,
				),

			TomorrowTemp:
				weather.TemperatureText(
					tomorrow.MinTemp,
					tomorrow.MaxTemp,
				),

			TomorrowIcon:
				weather.R15MIcon(
					tomorrow.Condition,
				),

			TomorrowDate:
				weather.DayText(
					tomorrow.Date,
				),

			ThirdTemp:
				weather.TemperatureText(
					third.MinTemp,
					third.MaxTemp,
				),

			ThirdIcon:
				weather.R15MIcon(
					third.Condition,
				),

			ThirdDate:
				weather.DayText(
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
EOF

echo
echo "[8/11] Runtime principal..."

cat > internal/app/runtime.go <<'EOF'
package app

import (
	"context"
	"fmt"
	"time"

	"github.com/twossh/minitela/internal/r15m"
)

const (
	DefaultReconnectDelay =
		2 * time.Second

	WeatherRefreshInterval =
		30 * time.Minute

	WeatherRetryInterval =
		5 * time.Minute
)

type Logger func(
	format string,
	args ...any,
)

type Options struct {
	Screen r15m.Screen

	Brightness *int

	City string

	WeatherAPIKey string

	MonitorInterval time.Duration

	ReconnectDelay time.Duration

	Once bool

	Logf Logger
}

func Run(
	ctx context.Context,
	opts Options,
) error {
	if opts.MonitorInterval <
		time.Second {
		opts.MonitorInterval =
			10 * time.Second
	}

	if opts.ReconnectDelay <
		time.Second {
		opts.ReconnectDelay =
			DefaultReconnectDelay
	}

	if opts.Logf == nil {
		opts.Logf =
			func(
				string,
				...any,
			) {}
	}

	if opts.Once {
		return runOnce(
			ctx,
			opts,
		)
	}

	return runContinuous(
		ctx,
		opts,
	)
}

func runOnce(
	ctx context.Context,
	opts Options,
) error {
	conn, err :=
		connectAndRestore(
			ctx,
			opts,
		)

	if err != nil {
		return err
	}

	defer conn.Close()

	switch opts.Screen {
	case r15m.ScreenMonitor:
		syncer :=
			r15m.NewMonitorSyncer(
				conn,
			)

		result, err :=
			syncer.Sync()

		if err != nil {
			return err
		}

		logMonitorSnapshot(
			opts.Logf,
			result,
		)

	case r15m.ScreenWeather:
		return syncWeatherOnce(
			ctx,
			conn,
			opts,
		)
	}

	return nil
}

func runContinuous(
	ctx context.Context,
	opts Options,
) error {
	for {
		if ctx.Err() != nil {
			return nil
		}

		conn, err :=
			connectAndRestore(
				ctx,
				opts,
			)

		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			opts.Logf(
				"[%s] aguardando MiniTela: %v",
				now(),
				err,
			)

			if !waitContext(
				ctx,
				opts.ReconnectDelay,
			) {
				return nil
			}

			continue
		}

		opts.Logf(
			"[%s] conectado em %s",
			now(),
			conn.Device.Path,
		)

		err =
			runConnected(
				ctx,
				conn,
				opts,
			)

		_ = conn.Close()

		if ctx.Err() != nil {
			return nil
		}

		if err != nil {
			opts.Logf(
				"[%s] conexão perdida: %v",
				now(),
				err,
			)
		}

		opts.Logf(
			"[%s] tentando reconectar...",
			now(),
		)

		if !waitContext(
			ctx,
			opts.ReconnectDelay,
		) {
			return nil
		}
	}
}

func connectAndRestore(
	ctx context.Context,
	opts Options,
) (*r15m.Connection, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	conn, err :=
		r15m.Connect()

	if err != nil {
		return nil, err
	}

	success := false

	defer func() {
		if !success {
			_ = conn.Close()
		}
	}()

	if opts.Brightness != nil {
		value :=
			*opts.Brightness

		if value < 0 {
			value = 0
		}

		if value > 100 {
			value = 100
		}

		actual, err :=
			conn.WriteRegisterVerified(
				r15m.RegisterBrightness,
				uint32(value),
			)

		if err != nil {
			return nil, fmt.Errorf(
				"restaurar brilho: %w",
				err,
			)
		}

		opts.Logf(
			"[%s] brilho restaurado: %d%%",
			now(),
			actual,
		)
	}

	if err := conn.SetScreen(
		opts.Screen,
	); err != nil {
		return nil, fmt.Errorf(
			"restaurar tela %s: %w",
			opts.Screen.String(),
			err,
		)
	}

	opts.Logf(
		"[%s] tela restaurada: %s (%d)",
		now(),
		opts.Screen.String(),
		opts.Screen,
	)

	success = true

	return conn, nil
}

func runConnected(
	ctx context.Context,
	conn *r15m.Connection,
	opts Options,
) error {
	switch opts.Screen {
	case r15m.ScreenMonitor:
		return runMonitor(
			ctx,
			conn,
			opts,
		)

	case r15m.ScreenWeather:
		return runWeather(
			ctx,
			conn,
			opts,
		)

	default:
		return runStaticScreen(
			ctx,
			conn,
			opts,
		)
	}
}

func runMonitor(
	ctx context.Context,
	conn *r15m.Connection,
	opts Options,
) error {
	syncer :=
		r15m.NewMonitorSyncer(
			conn,
		)

	if err := syncMonitor(
		syncer,
		opts.Logf,
	); err != nil {
		return err
	}

	ticker :=
		time.NewTicker(
			opts.MonitorInterval,
		)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			if err := syncMonitor(
				syncer,
				opts.Logf,
			); err != nil {
				return err
			}
		}
	}
}

func syncMonitor(
	syncer *r15m.MonitorSyncer,
	logf Logger,
) error {
	result, err :=
		syncer.Sync()

	if err != nil {
		return err
	}

	logMonitorSnapshot(
		logf,
		result,
	)

	return nil
}

func logMonitorSnapshot(
	logf Logger,
	result *r15m.MonitorSnapshot,
) {
	wifi := "desconectado"

	if result.WiFiConnected {
		wifi =
			fmt.Sprintf(
				"%s (%d%%)",
				result.WiFiDisplay,
				result.WiFiQuality,
			)
	}

	bluetooth :=
		"desconectado"

	if result.BluetoothConnected {
		bluetooth =
			result.BluetoothDisplay
	}

	logf(
		"[%s] bateria=%d%% nível=%d | wifi=%s | bluetooth=%s | escritas=%d",
		now(),
		result.BatteryPercent,
		result.BatteryLevel,
		wifi,
		bluetooth,
		result.Writes,
	)
}

func syncWeatherOnce(
	ctx context.Context,
	conn *r15m.Connection,
	opts Options,
) error {
	if opts.City == "" {
		return fmt.Errorf(
			"cidade não configurada",
		)
	}

	if opts.WeatherAPIKey == "" {
		return fmt.Errorf(
			"chave WeatherAPI não configurada",
		)
	}

	opts.Logf(
		"[%s] consultando clima para %s...",
		now(),
		opts.City,
	)

	syncer :=
		r15m.NewWeatherSyncer(
			conn,
			opts.City,
			opts.WeatherAPIKey,
		)

	result, err :=
		syncer.Sync(ctx)

	if err != nil {
		return fmt.Errorf(
			"consultar/sincronizar clima: %w",
			err,
		)
	}

	logWeatherSnapshot(
		opts.Logf,
		result,
	)

	return nil
}

func runWeather(
	ctx context.Context,
	conn *r15m.Connection,
	opts Options,
) error {
	if opts.City == "" {
		return fmt.Errorf(
			"cidade não configurada",
		)
	}

	if opts.WeatherAPIKey == "" {
		return fmt.Errorf(
			"chave WeatherAPI não configurada",
		)
	}

	syncer :=
		r15m.NewWeatherSyncer(
			conn,
			opts.City,
			opts.WeatherAPIKey,
		)

	opts.Logf(
		"[%s] consultando clima para %s...",
		now(),
		opts.City,
	)

	refreshDelay :=
		WeatherRefreshInterval

	result, err :=
		syncer.Sync(ctx)

	if err != nil {
		opts.Logf(
			"[%s] clima: consulta inicial falhou: %v",
			now(),
			err,
		)

		refreshDelay =
			WeatherRetryInterval
	} else {
		logWeatherSnapshot(
			opts.Logf,
			result,
		)
	}

	heartbeat :=
		time.NewTicker(
			opts.MonitorInterval,
		)

	defer heartbeat.Stop()

	refresh :=
		time.NewTimer(
			refreshDelay,
		)

	defer refresh.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-heartbeat.C:
			page, err :=
				conn.ReadRegister(
					r15m.RegisterCurrentPage,
				)

			if err != nil {
				return err
			}

			if page != uint32(
				r15m.ScreenWeather,
			) {
				if err :=
					conn.SetScreen(
						r15m.ScreenWeather,
					); err != nil {
					return err
				}
			}

		case <-refresh.C:
			opts.Logf(
				"[%s] atualizando clima...",
				now(),
			)

			result, err :=
				syncer.Sync(ctx)

			if err != nil {
				opts.Logf(
					"[%s] clima: atualização falhou: %v",
					now(),
					err,
				)

				refresh.Reset(
					WeatherRetryInterval,
				)

				continue
			}

			logWeatherSnapshot(
				opts.Logf,
				result,
			)

			refresh.Reset(
				WeatherRefreshInterval,
			)
		}
	}
}

func logWeatherSnapshot(
	logf Logger,
	result *r15m.WeatherSnapshot,
) {
	logf(
		"[%s] clima=%s | hoje=%s icon=%d | amanhã=%s %s icon=%d | terceiro=%s %s icon=%d | escritas=%d",
		now(),
		result.City,
		result.TodayTemp,
		result.TodayIcon,
		result.TomorrowDate,
		result.TomorrowTemp,
		result.TomorrowIcon,
		result.ThirdDate,
		result.ThirdTemp,
		result.ThirdIcon,
		result.Writes,
	)
}

func runStaticScreen(
	ctx context.Context,
	conn *r15m.Connection,
	opts Options,
) error {
	ticker :=
		time.NewTicker(
			opts.MonitorInterval,
		)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			page, err :=
				conn.ReadRegister(
					r15m.RegisterCurrentPage,
				)

			if err != nil {
				return err
			}

			if page != uint32(
				opts.Screen,
			) {
				opts.Logf(
					"[%s] página alterada (%d); restaurando %s...",
					now(),
					page,
					opts.Screen.String(),
				)

				if err :=
					conn.SetScreen(
						opts.Screen,
					); err != nil {
					return err
				}
			}
		}
	}
}

func waitContext(
	ctx context.Context,
	duration time.Duration,
) bool {
	timer :=
		time.NewTimer(
			duration,
		)

	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false

	case <-timer.C:
		return true
	}
}

func now() string {
	return time.Now().Format(
		"15:04:05",
	)
}
EOF

echo
echo "[9/11] Executável principal MiniTela..."

cat > cmd/minitela/main.go <<'EOF'
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/twossh/minitela/internal/app"
	"github.com/twossh/minitela/internal/config"
	"github.com/twossh/minitela/internal/r15m"
)

const (
	appName =
		"MiniTela"

	appVersion =
		"0.1.0-dev"
)

func main() {
	screenFlag :=
		flag.String(
			"screen",
			"",
			"seleciona e salva: whatsapp, notes, monitor ou weather",
		)

	brightnessFlag :=
		flag.Int(
			"set-brightness",
			-1,
			"define e salva brilho entre 0 e 100",
		)

	noRestore :=
		flag.Bool(
			"no-restore",
			false,
			"não restaura a última tela",
		)

	once :=
		flag.Bool(
			"once",
			false,
			"sincroniza uma vez e encerra",
		)

	flag.Parse()

	if *brightnessFlag < -1 ||
		*brightnessFlag > 100 {
		fail(
			"brilho deve estar entre 0 e 100",
			2,
		)
	}

	cfg, err :=
		config.Load()

	if err != nil {
		fail(
			fmt.Sprintf(
				"carregar configuração: %v",
				err,
			),
			1,
		)
	}

	explicitScreen := false
	changed := false

	if *screenFlag != "" {
		screen, err :=
			r15m.ParseScreen(
				*screenFlag,
			)

		if err != nil {
			fail(
				err.Error(),
				2,
			)
		}

		cfg.LastScreen =
			screenConfigName(
				screen,
			)

		explicitScreen = true
		changed = true
	}

	if *brightnessFlag >= 0 {
		value :=
			*brightnessFlag

		cfg.Brightness =
			&value

		changed = true
	}

	if changed {
		if err := config.Save(
			cfg,
		); err != nil {
			fail(
				fmt.Sprintf(
					"salvar configuração: %v",
					err,
				),
				1,
			)
		}
	}

	printHeader(cfg)

	if *noRestore &&
		!explicitScreen {
		runWithoutRestore(cfg)
		return
	}

	if !cfg.RestoreLastScreen &&
		!explicitScreen {
		runWithoutRestore(cfg)
		return
	}

	screen, err :=
		r15m.ParseScreen(
			cfg.LastScreen,
		)

	if err != nil {
		fail(
			fmt.Sprintf(
				"interpretar última tela: %v",
				err,
			),
			1,
		)
	}

	ctx, stop :=
		signal.NotifyContext(
			context.Background(),
			os.Interrupt,
			syscall.SIGTERM,
		)

	defer stop()

	interval :=
		time.Duration(
			cfg.MonitorIntervalSeconds,
		) * time.Second

	fmt.Printf(
		"Tela desejada : %s (%d)\n",
		screen.String(),
		screen,
	)

	fmt.Printf(
		"Atualização   : %s\n",
		interval,
	)

	fmt.Println(
		"Reconexão     : automática",
	)

	if !*once {
		fmt.Println(
			"Ctrl+C        : encerrar",
		)
	}

	fmt.Println()

	err =
		app.Run(
			ctx,
			app.Options{
				Screen:
					screen,

				Brightness:
					cfg.Brightness,

				City:
					cfg.City,

				WeatherAPIKey:
					cfg.WeatherAPIKey,

				MonitorInterval:
					interval,

				ReconnectDelay:
					app.DefaultReconnectDelay,

				Once:
					*once,

				Logf:
					func(
						format string,
						args ...any,
					) {
						fmt.Printf(
							format+"\n",
							args...,
						)
					},
			},
		)

	if err != nil {
		fail(
			fmt.Sprintf(
				"MiniTela: %v",
				err,
			),
			1,
		)
	}

	if !*once {
		fmt.Println()
		fmt.Println(
			"MiniTela encerrada.",
		)
	}
}

func printHeader(
	cfg config.Config,
) {
	fmt.Printf(
		"%s %s\n",
		appName,
		appVersion,
	)

	fmt.Println(
		"MiniTela para Positivo R15M",
	)

	fmt.Printf(
		"Sistema       : %s/%s\n",
		runtime.GOOS,
		runtime.GOARCH,
	)

	fmt.Println()

	path, _ :=
		config.Path()

	fmt.Printf(
		"Configuração  : %s\n",
		path,
	)

	fmt.Printf(
		"Última tela   : %s\n",
		cfg.LastScreen,
	)

	fmt.Printf(
		"Cidade        : %s\n",
		cfg.City,
	)

	if cfg.Brightness == nil {
		fmt.Println(
			"Brilho        : manter atual",
		)
	} else {
		fmt.Printf(
			"Brilho        : %d%%\n",
			*cfg.Brightness,
		)
	}

	if cfg.WeatherAPIKey == "" {
		fmt.Println(
			"WeatherAPI    : não configurada",
		)
	} else {
		fmt.Println(
			"WeatherAPI    : configurada",
		)
	}

	fmt.Printf(
		"Restaurar     : %t\n",
		cfg.RestoreLastScreen,
	)

	fmt.Println()
}

func runWithoutRestore(
	cfg config.Config,
) {
	fmt.Println(
		"Modo          : sem restauração",
	)

	fmt.Println(
		"Conectando somente para diagnóstico...",
	)

	conn, err :=
		r15m.Connect()

	if err != nil {
		fail(
			err.Error(),
			1,
		)
	}

	defer conn.Close()

	if cfg.Brightness != nil {
		if _, err :=
			conn.WriteRegisterVerified(
				r15m.RegisterBrightness,
				uint32(
					*cfg.Brightness,
				),
			); err != nil {
			fail(
				fmt.Sprintf(
					"aplicar brilho: %v",
					err,
				),
				1,
			)
		}
	}

	page, err :=
		conn.ReadRegister(
			r15m.RegisterCurrentPage,
		)

	if err != nil {
		fail(
			fmt.Sprintf(
				"ler página: %v",
				err,
			),
			1,
		)
	}

	fmt.Printf(
		"Tela atual    : %s (%d)\n",
		r15m.Screen(page).String(),
		page,
	)

	fmt.Println(
		"Status        : pronta",
	)
}

func screenConfigName(
	screen r15m.Screen,
) string {
	switch screen {
	case r15m.ScreenWhatsApp:
		return "whatsapp"

	case r15m.ScreenNotes:
		return "notes"

	case r15m.ScreenMonitor:
		return "monitor"

	case r15m.ScreenWeather:
		return "weather"

	default:
		return "monitor"
	}
}

func fail(
	message string,
	code int,
) {
	fmt.Fprintf(
		os.Stderr,
		"Erro: %s\n",
		message,
	)

	os.Exit(code)
}
EOF

echo
echo "[10/11] Criando teste independente da WeatherAPI..."

cat > cmd/minitela-weather-test/main.go <<'EOF'
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/twossh/minitela/internal/config"
	"github.com/twossh/minitela/internal/weather"
)

func main() {
	cfg, err :=
		config.Load()

	if err != nil {
		fail(err)
	}

	if cfg.City == "" {
		fail(fmt.Errorf(
			"cidade não configurada",
		))
	}

	if cfg.WeatherAPIKey == "" {
		fail(fmt.Errorf(
			"WeatherAPI não configurada",
		))
	}

	fmt.Println(
		"=== MiniTela WeatherAPI Test ===",
	)

	fmt.Printf(
		"Cidade configurada: %s\n",
		cfg.City,
	)

	fmt.Println(
		"WeatherAPI: configurada",
	)

	fmt.Println()
	fmt.Println(
		"Consultando...",
	)

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			15*time.Second,
		)

	defer cancel()

	client :=
		weather.NewClient(
			cfg.WeatherAPIKey,
		)

	forecast, err :=
		client.GetForecast(
			ctx,
			cfg.City,
		)

	if err != nil {
		fail(err)
	}

	fmt.Println()
	fmt.Printf(
		"Local: %s - %s - %s\n",
		forecast.Location.Name,
		forecast.Location.Region,
		forecast.Location.Country,
	)

	for i, day :=
		range forecast.Days {
		fmt.Printf(
			"Dia %d: %s | %s | %d°/%d° | icon=%d\n",
			i+1,
			day.Date.Format(
				"02/01/2006",
			),
			day.Condition,
			day.MinTemp,
			day.MaxTemp,
			weather.R15MIcon(
				day.Condition,
			),
		)
	}

	fmt.Println()
	fmt.Println(
		"WeatherAPI: OK",
	)
}

func fail(err error) {
	fmt.Fprintf(
		os.Stderr,
		"Erro: %v\n",
		err,
	)

	os.Exit(1)
}
EOF

echo
echo "[11/11] Formatando, testando e compilando..."

gofmt -w \
    internal/config/config.go \
    cmd/minitela-config/main.go \
    internal/weather/client.go \
    internal/weather/display.go \
    internal/weather/display_test.go \
    internal/r15m/registers.go \
    internal/r15m/weather.go \
    internal/app/runtime.go \
    cmd/minitela/main.go \
    cmd/minitela-weather-test/main.go

go mod tidy

echo
echo "=== GO TEST ==="
go test -count=1 ./...

echo
echo "=== GIT DIFF CHECK ==="
git diff --check

echo
echo "=== BUILD ==="

go build \
    -trimpath \
    -o bin/MiniTela \
    ./cmd/minitela

go build \
    -trimpath \
    -o bin/MiniTelaConfig \
    ./cmd/minitela-config

go build \
    -trimpath \
    -o bin/MiniTelaWeatherTest \
    ./cmd/minitela-weather-test

echo
echo "===================================================="
echo " ATUALIZAÇÃO CONCLUÍDA"
echo "===================================================="

echo
echo "Executáveis:"
ls -lh \
    bin/MiniTela \
    bin/MiniTelaConfig \
    bin/MiniTelaWeatherTest

echo
echo "Backup:"
echo "$BACKUP"

echo
echo "Próximo passo:"
echo "./bin/MiniTelaConfig --show"
echo
MINITELA_SCRIPT

chmod +x /tmp/minitela-weatherapi-upgrade.sh

/tmp/minitela-weatherapi-upgrade.sh