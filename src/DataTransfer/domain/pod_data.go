package domain

import (
	excelParser "github.com/HyperloopUPV-H8/Backend-H8/Shared/ExcelParser/domain/board"
	packetParser "github.com/HyperloopUPV-H8/Backend-H8/Shared/PacketAdapter/domain"
)

type PodData struct {
	Boards map[string]Board
}

func (podData *PodData) UpdatePacket(pu packetParser.PacketUpdate) {
	packet := podData.GetPacket(pu.ID)
	packet.UpdatePacket(pu)
}

func (podData *PodData) GetPacket(id uint16) *Packet {
	for _, board := range podData.Boards {
		packet, exists := board.Packets[id]
		if exists {
			return &packet
		}
	}
	return nil
}

func NewPodData(rawBoards map[string]excelParser.Board) PodData {
	return PodData{
		Boards: getBoards(rawBoards),
	}
}

func getBoards(rawBoards map[string]excelParser.Board) map[string]Board {
	boards := make(map[string]Board)
	for name, board := range rawBoards {
		board := Board{
			Name:    name,
			Packets: NewPackets(board.GetPackets()),
		}
		boards[board.Name] = board
	}
	return boards
}