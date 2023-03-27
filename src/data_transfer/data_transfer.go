package data_transfer

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/HyperloopUPV-H8/Backend-H8/vehicle/models"
	ws_models "github.com/HyperloopUPV-H8/Backend-H8/websocket_handle/models"
)

type DataTransfer struct {
	bufMx     sync.Mutex
	packetBuf map[uint16]models.Update
	ticker    *time.Ticker
	channel   chan ws_models.MessageTarget
}

func New(rate time.Duration) (*DataTransfer, chan ws_models.MessageTarget) {
	dataTransfer := &DataTransfer{
		bufMx:     sync.Mutex{},
		packetBuf: make(map[uint16]models.Update),
		ticker:    time.NewTicker(rate),
		channel:   make(chan ws_models.MessageTarget),
	}

	go dataTransfer.run()

	return dataTransfer, dataTransfer.channel
}

func (dataTransfer *DataTransfer) run() {
	for {
		<-dataTransfer.ticker.C
		if len(dataTransfer.packetBuf) == 0 {
			continue
		}
		dataTransfer.channel <- ws_models.NewMessageTargetRaw([]string{}, "podData/update", dataTransfer.getJSON())
	}
}

func (dataTransfer *DataTransfer) getJSON() []byte {
	dataTransfer.bufMx.Lock()
	defer dataTransfer.bufMx.Unlock()
	data, err := json.Marshal(dataTransfer.packetBuf)
	if err != nil {
		log.Fatalf("data transfer: getJSON: %s\n", err)
	}
	dataTransfer.packetBuf = make(map[uint16]models.Update, len(dataTransfer.packetBuf))
	return data
}

func (dataTransfer *DataTransfer) Update(update models.Update) {
	dataTransfer.bufMx.Lock()
	defer dataTransfer.bufMx.Unlock()
	dataTransfer.packetBuf[update.ID] = update
}
