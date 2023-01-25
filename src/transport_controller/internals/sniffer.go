package internals

import (
	"log"

	"github.com/HyperloopUPV-H8/Backend-H8/transport_controller/models"
	"github.com/google/gopacket/pcap"
)

type Sniffer struct {
	source *pcap.Handle
	config models.Config
}

func OpenSniffer(device string, live bool, config models.Config) *Sniffer {
	sniffer := &Sniffer{
		source: obtainSource(device, live, config),
		config: config,
	}

	go sniffer.StartReading()

	return sniffer
}

func obtainSource(device string, live bool, config models.Config) *pcap.Handle {
	var (
		source *pcap.Handle
		err    error
	)

	log.Println(pcap.FindAllDevs())

	if live {
		source, err = pcap.OpenLive(device, config.Snaplen, config.Promisc, config.Timeout)
	} else {
		source, err = pcap.OpenOffline(device)
	}

	if err != nil {
		log.Fatalf("sniffer: obtainSource: %s\n", err)
	}

	err = source.SetBPFFilter(config.BPF)
	if err != nil {
		log.Fatalf("sniffer: obtainSource: %s\n", err)
	}

	return source
}

func (sniffer *Sniffer) StartReading() {
	for {
		payload, _, err := sniffer.source.ReadPacketData()
		if err != nil {
			continue
		}

		if payload[12] != 0x08 || payload[13] != 0x00 {
			continue
		}

		switch payload[23] {
		case 0x04:
			switch payload[43] {
			case 0x11:
				sniffer.config.Dump <- payload[62:]
			case 0x06:
				sniffer.config.Dump <- payload[74:]
			}
		case 0x11:
			sniffer.config.Dump <- payload[42:]
		case 0x06:
			sniffer.config.Dump <- payload[54:]
		}

	}
}

func (sniffer *Sniffer) Close() {
	sniffer.source.Close()
}

func (sniffer *Sniffer) Stats() (recieved int, dropped int) {
	stats, err := sniffer.source.Stats()
	if err != nil {
		return
	}
	return stats.PacketsReceived, stats.PacketsDropped
}
