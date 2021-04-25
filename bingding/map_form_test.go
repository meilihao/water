package binding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapFormOldMapSlice(t *testing.T) {
	b := Uri
	m := map[string][]string{
		"name": {"mike", "job"},
	}

	var objNotInit map[string][]string = map[string][]string{}
	err := b.Bind(m, &objNotInit)
	assert.Equal(t, err, nil)
	assert.Equal(t, len(objNotInit), 1)
}

func TestMapFormMapSlice(t *testing.T) {
	m := map[string][]string{
		"name": {"mike", "job"},
	}

	var objNotInit map[string][]string
	err := MapForm(&objNotInit, m, nil, "")
	assert.Equal(t, err, nil)
	assert.Equal(t, len(objNotInit), 1)

	var obj map[string][]string = map[string][]string{}
	err = MapForm(&obj, m, nil, "")
	assert.Equal(t, err, nil)
	assert.Equal(t, len(obj), 1)
}

func TestMapFormMapString(t *testing.T) {
	m := map[string][]string{
		"name": {"mike", "job"},
	}

	var objNotInit map[string]string
	err := MapForm(&objNotInit, m, nil, "")
	assert.Equal(t, err, nil)
	assert.Equal(t, len(objNotInit), 1)
	assert.Equal(t, len(objNotInit), 1)
	assert.Equal(t, objNotInit["name"], "job")

	var obj map[string]string = map[string]string{}
	err = MapForm(&obj, m, nil, "")
	assert.Equal(t, err, nil)
	assert.Equal(t, len(obj), 1)
	assert.Equal(t, obj["name"], "job")
}
