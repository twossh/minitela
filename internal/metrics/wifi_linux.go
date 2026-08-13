//go:build linux

package metrics

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type WiFi struct {
	Interface string
	SSID      string
	SignalDBM int
	Quality   int
}

func ReadWiFi() (*WiFi, error) {
	iface, err := findConnectedWiFiInterface()
	if err != nil {
		return nil, err
	}

	output, err := exec.Command(
		"iw",
		"dev",
		iface,
		"link",
	).Output()
	if err != nil {
		return nil, fmt.Errorf(
			"consultar Wi-Fi %s: %w",
			iface,
			err,
		)
	}

	var (
		ssid       string
		signalDBM  int
		haveSignal bool
	)

	scanner := bufio.NewScanner(
		strings.NewReader(string(output)),
	)

	for scanner.Scan() {
		line := strings.TrimSpace(
			scanner.Text(),
		)

		if strings.HasPrefix(line, "SSID:") {
			ssid = strings.TrimSpace(
				strings.TrimPrefix(
					line,
					"SSID:",
				),
			)
			continue
		}

		if strings.HasPrefix(line, "signal:") {
			fields := strings.Fields(line)

			if len(fields) >= 2 {
				value, err := strconv.ParseFloat(
					fields[1],
					64,
				)
				if err == nil {
					signalDBM = int(value)
					haveSignal = true
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if ssid == "" {
		return nil, fmt.Errorf(
			"SSID não encontrado em %s",
			iface,
		)
	}

	if !haveSignal {
		return nil, fmt.Errorf(
			"sinal Wi-Fi não encontrado em %s",
			iface,
		)
	}

	return &WiFi{
		Interface: iface,
		SSID:      ssid,
		SignalDBM: signalDBM,
		Quality:   SignalQuality(signalDBM),
	}, nil
}

func findConnectedWiFiInterface() (
	string,
	error,
) {
	output, err := exec.Command(
		"iw",
		"dev",
	).Output()
	if err != nil {
		return "", fmt.Errorf(
			"executar iw dev: %w",
			err,
		)
	}

	var interfaces []string

	scanner := bufio.NewScanner(
		strings.NewReader(string(output)),
	)

	for scanner.Scan() {
		line := strings.TrimSpace(
			scanner.Text(),
		)

		if strings.HasPrefix(
			line,
			"Interface ",
		) {
			iface := strings.TrimSpace(
				strings.TrimPrefix(
					line,
					"Interface ",
				),
			)

			if iface != "" {
				interfaces = append(
					interfaces,
					iface,
				)
			}
		}
	}

	for _, iface := range interfaces {
		linkOutput, err := exec.Command(
			"iw",
			"dev",
			iface,
			"link",
		).Output()
		if err != nil {
			continue
		}

		if !strings.Contains(
			string(linkOutput),
			"Not connected",
		) {
			return iface, nil
		}
	}

	return "", fmt.Errorf(
		"nenhuma interface Wi-Fi conectada",
	)
}

// SignalQuality converts Wi-Fi signal strength to 0..100.
//
// -100 dBm or worse -> 0%
// -50 dBm or better -> 100%
func SignalQuality(dbm int) int {
	if dbm <= -100 {
		return 0
	}

	if dbm >= -50 {
		return 100
	}

	return 2 * (dbm + 100)
}
