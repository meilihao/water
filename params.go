package water

import (
	"strconv"
)

// Param is a single URL parameter, consisting of a key and a value.
type Param struct {
	Name  string
	Value string
}

// Params is a Param-slice, as returned by the router.
// The slice is ordered, the first URL parameter is also the first slice value.
// It is therefore safe to read values by the index.
type Params []Param

// String returns value by given param name.
// panic if param not exits
func (p Params) String(name string) string {
	for i := range p {
		if p[i].Name == name {
			return p[i].Value
		}
	}

	panic("Params not exist: " + name)
}

// Bool return bool  with error.
func (p Params) Bool(name string) (bool, error) {
	return strconv.ParseBool(p.String(name))
}

// Int return int with error.
func (p Params) Int(name string) (int, error) {
	return strconv.Atoi(p.String(name))
}

// Int64 returns int64 with error.
func (p Params) Int64(name string) (int64, error) {
	return strconv.ParseInt(p.String(name), 10, 64)
}

// Uint returns uint with error.
func (p Params) Uint(name string) (uint, error) {
	v, err := strconv.ParseUint(p.String(name), 10, 64)
	return uint(v), err
}

// Uint64 returns uint64 with error.
func (p Params) Uint64(name string) (uint64, error) {
	return strconv.ParseUint(p.String(name), 10, 64)
}

// Float64 returns float64 with error.
func (p Params) Float64(name string) (float64, error) {
	return strconv.ParseFloat(p.String(name), 64)
}

// MustBool returns bool.
func (p Params) MustBool(name string) bool {
	v, _ := p.Bool(name)
	return v
}

// MustInt returns int.
func (p Params) MustInt(name string) int {
	v, _ := p.Int(name)
	return v
}

// MustInt64 returns int64.
func (p Params) MustInt64(name string) int64 {
	v, _ := p.Int64(name)
	return v
}

// MustUint returns uint.
func (p Params) MustUint(name string) uint {
	v, _ := p.Uint(name)
	return uint(v)
}

// MustUint64 returns uint64.
func (p Params) MustUint64(name string) uint64 {
	v, _ := p.Uint64(name)
	return v
}

// MustFloat64 returns float64l
func (p Params) MustFloat64(name string) float64 {
	v, _ := p.Float64(name)
	return v
}
