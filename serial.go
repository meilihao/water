package water

import (
	"github.com/satori/go.uuid"
)

type SerialAdapter interface {
	Id() string
}

type DefaultSerial struct{}

func (s DefaultSerial) Id() string {
	return uuid.NewV1().String()
}
