package water

import (
	"strings"

	"github.com/satori/go.uuid"
)

type SerialAdapter interface {
	Id() string
}

type DefaultSerial struct{}

func (s DefaultSerial) Id() string {
	return strings.Replace(uuid.Must(uuid.NewV1()).String(), "-", "", -1)
}
