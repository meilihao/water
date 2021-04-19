package water

// a set of environment variables.
// 存储water的环境变量
type Environ map[string]interface{}

func (m Environ) GetBool(name string) bool {
	return m[name].(bool)
}

func (m Environ) GetInt(name string) int {
	return m[name].(int)
}

func (m Environ) GetInt64(name string) int64 {
	return m[name].(int64)
}

func (m Environ) GetUint(name string) uint {
	return m[name].(uint)
}

func (m Environ) GetUint64(name string) uint64 {
	return m[name].(uint64)
}

func (m Environ) GetFloat64(name string) float64 {
	return m[name].(float64)
}

func (m Environ) GetString(name string) string {
	return m[name].(string)
}

// panic if name not exist
func (m Environ) Get(name string) interface{} {
	if v, ok := m[name]; ok {
		return v
	} else {
		panic("Environ not exist: " + name)
	}
}

// panic if name already exists
func (m Environ) Set(name string, v interface{}) {
	if m.Has(name) {
		panic("double Environ: " + name)
	} else {
		m[name] = v
	}
}

func (m Environ) Has(name string) bool {
	_, ok := m[name]
	return ok
}

// Set is used to store a new key/value pair exclusively for this context.
func (c *Context) Set(name string, value interface{}) {
	if c.Environ.Has(name) {
		panic("double Environ: " + name)
	} else {
		c.Environ[name] = value
	}
}

// Get returns the value for the given name if it exists, otherwise it panics.
func (c *Context) Get(name string) interface{} {
	if value, exists := c.Environ[name]; exists {
		return value
	}
	panic("Environ not exist: " + name)
}
