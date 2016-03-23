package water

import (
	"github.com/satori/go.uuid"
)

type SerialAdapter interface {
	Id() string
}

type WaterSerial struct{}

func (ws *WaterSerial) Id() string {
	return uuid.NewV4().String()
}
