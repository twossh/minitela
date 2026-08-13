package r15m

import (
	"fmt"

	"github.com/twossh/minitela/internal/metrics"
)

type BatterySyncResult struct {
	Capacity int
	Level    uint32
	Status   string
}

func (c *Connection) SyncBattery() (
	*BatterySyncResult,
	error,
) {
	battery, err := metrics.ReadBattery()
	if err != nil {
		return nil, err
	}

	text := BatteryText(
		battery.Capacity,
	)

	level := BatteryLevel(
		battery.Capacity,
	)

	// Battery percentage is a textual display register.
	if err := c.WriteStringRegister(
		RegisterBatteryText,
		text,
	); err != nil {
		return nil, fmt.Errorf(
			"atualizar percentual da bateria: %w",
			err,
		)
	}

	// Register 1150 controls the graphical battery level.
	//
	// Unlike normal configuration registers such as brightness
	// and current page, this display-state register must not be
	// verified by reading it back: the R15M firmware may return
	// another internal/current value after accepting the write.
	//
	// A valid SET_REGISTER ACK is sufficient here.
	if err := c.WriteRegister(
		RegisterBatteryLevel,
		level,
	); err != nil {
		return nil, fmt.Errorf(
			"atualizar nível gráfico da bateria: %w",
			err,
		)
	}

	return &BatterySyncResult{
		Capacity: battery.Capacity,
		Level:    level,
		Status:   battery.Status,
	}, nil
}
