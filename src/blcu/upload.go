package blcu

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"time"

	"github.com/HyperloopUPV-H8/Backend-H8/common"
	"github.com/HyperloopUPV-H8/Backend-H8/vehicle/models"
	"github.com/pin/tftp/v3"
)

type uploadRequest struct {
	Board string `json:"board"`
	File  string `json:"file"`
}

func (blcu *BLCU) upload(payload json.RawMessage) error {
	blcu.trace.Debug().Msg("Handling upload")

	var request uploadRequest
	if err := json.Unmarshal(payload, &request); err != nil {
		blcu.trace.Error().Err(err).Stack().Msg("Unmarshal payload")
		return err
	}

	if err := blcu.requestUpload(request.Board); err != nil {
		blcu.trace.Error().Err(err).Stack().Msg("Request upload")
		return err
	}

	decoded, err := base64.StdEncoding.DecodeString(request.File)
	if err != nil {
		blcu.trace.Error().Err(err).Stack().Msg("Decode payload")
		return err
	}

	reader := bytes.NewReader(decoded)
	return blcu.WriteTFTP(reader, int(reader.Size()), blcu.notifyUploadProgress)
}

func (blcu *BLCU) requestUpload(board string) error {
	blcu.trace.Info().Str("board", board).Msg("Requesting upload")

	uploadOrder := blcu.createUploadOrder(board)
	if err := blcu.sendOrder(uploadOrder); err != nil {
		return err
	}

	// TODO: remove hardcoded timeout
	if _, err := common.ReadTimeout(blcu.ackChannel, time.Second*10); err != nil {
		return err
	}

	return nil
}

func (blcu *BLCU) createUploadOrder(board string) models.Order {
	return models.Order{
		ID: blcu.config.Packets.Upload.Id,
		Fields: map[string]models.Field{
			blcu.config.Packets.Upload.Field: {
				Value:     board,
				IsEnabled: true,
			},
		},
	}
}

func (blcu *BLCU) WriteTFTP(reader io.Reader, size int, onProgress func(float64)) error {
	blcu.trace.Info().Msg("Writing TFTP")

	client, err := tftp.NewClient(blcu.addr)
	if err != nil {
		return err
	}

	sender, err := client.Send("a.bin", "octet")
	if err != nil {
		return err
	}

	upload := NewUpload(reader, size, onProgress)
	_, err = sender.ReadFrom(&upload)

	return err
}

type uploadResponse struct {
	Percentage float64 `json:"percentage"`
	IsSuccess  bool    `json:"success,omitempty"`
}

func (blcu *BLCU) notifyUploadFailure() {
	blcu.trace.Warn().Msg("Upload failed")
	blcu.sendMessage(blcu.config.Topics.Download, uploadResponse{Percentage: 0.0, IsSuccess: false})
}

func (blcu *BLCU) notifyUploadSuccess() {
	blcu.trace.Info().Msg("Upload success")
	blcu.sendMessage(blcu.config.Topics.Download, uploadResponse{Percentage: 1.0, IsSuccess: true})
}

func (blcu *BLCU) notifyUploadProgress(percentage float64) {
	blcu.sendMessage(blcu.config.Topics.Download, uploadResponse{Percentage: percentage, IsSuccess: false})
}

type Upload struct {
	reader     io.Reader
	onProgress func(float64)
	total      int
	current    int
}

func NewUpload(reader io.Reader, size int, onProgress func(float64)) Upload {
	return Upload{
		reader:     reader,
		onProgress: onProgress,
		total:      size,
		current:    0,
	}
}

func (upload *Upload) Read(p []byte) (n int, err error) {
	n, err = upload.reader.Read(p)
	if err != nil {
		upload.current += n
		upload.onProgress(float64(upload.current) / float64(upload.total))
	}
	return n, err
}
