package r15m

type RegisterCache struct {
	nums    map[uint16]uint32
	strings map[uint16]string
}

func NewRegisterCache() *RegisterCache {
	return &RegisterCache{
		nums:    make(map[uint16]uint32),
		strings: make(map[uint16]string),
	}
}

func (c *RegisterCache) Reset() {
	if c == nil {
		return
	}

	clear(c.nums)
	clear(c.strings)
}

func (c *RegisterCache) WriteNumIfChanged(
	conn *Connection,
	regID uint16,
	value uint32,
) (bool, error) {
	if current, ok := c.nums[regID]; ok &&
		current == value {
		return false, nil
	}

	if err := conn.WriteRegister(
		regID,
		value,
	); err != nil {
		return false, err
	}

	c.nums[regID] = value

	return true, nil
}

func (c *RegisterCache) WriteStringIfChanged(
	conn *Connection,
	regID uint16,
	value string,
) (bool, error) {
	if current, ok := c.strings[regID]; ok &&
		current == value {
		return false, nil
	}

	if err := conn.WriteStringRegister(
		regID,
		[]byte(value),
	); err != nil {
		return false, err
	}

	c.strings[regID] = value

	return true, nil
}
