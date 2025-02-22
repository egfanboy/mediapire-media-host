package transfers

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/egfanboy/mediapire-common/messaging"
	"github.com/egfanboy/mediapire-media-host/internal/app"
	"github.com/egfanboy/mediapire-media-host/internal/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

const (
	// filepath.Dir returns . if the path in question does not have a parent
	noParentDir = "."
)

func handleTransferMessage(ctx context.Context, msg amqp091.Delivery) error {
	var tMsg messaging.TransferReadyMessage

	log.Info().Msg("Received transfer ready message for transfer")

	err := json.Unmarshal(msg.Body, &tMsg)
	if err != nil {
		msg := "failed to unmarshal transfer ready message"
		log.Err(err).Msg(msg)

		sendTransferUpdateMessage(ctx, tMsg.TransferId, &msg)
		return err
	}

	appInstance := app.GetApp()

	if tMsg.TargetId != appInstance.NodeId {
		log.Info().Msg("Transfer ready message is not for us. Skipping.")

		return nil
	}

	if tMsg.Content == nil || len(tMsg.Content) == 0 {
		log.Info().Msgf("Content for transfer %s is empty. Skipping.", tMsg.TransferId)

		return nil
	}

	log.Info().Msgf("Transfer ready message %s is for this node, downloading content to disk", tMsg.TransferId)

	reader := bytes.NewReader(tMsg.Content)

	zipReader, err := zip.NewReader(reader, int64(len(tMsg.Content)))
	if err != nil {
		msg := fmt.Sprintf("failed to read zip file content for transfer %s", tMsg.TransferId)
		log.Err(err).Msg(msg)

		sendTransferUpdateMessage(ctx, tMsg.TransferId, &msg)
		return err
	}

	for _, f := range zipReader.File {
		rc, err := f.Open()
		if err != nil {
			msg := fmt.Sprintf("failed to open file %s  for transfer %s", f.Name, tMsg.TransferId)
			log.Err(err).Msg(msg)

			sendTransferUpdateMessage(ctx, tMsg.TransferId, &msg)
			return err
		}

		defer rc.Close()

		// for now save it in the first directory in the config
		targetDir := appInstance.Directories[0]

		fileDir := filepath.Dir(f.Name)
		if fileDir != noParentDir {
			err := os.MkdirAll(path.Join(targetDir, fileDir), os.ModePerm)
			if err != nil {
				msg := err.Error()
				log.Err(err).Msg(msg)
				sendTransferUpdateMessage(ctx, tMsg.TransferId, &msg)
				return err
			}
		}

		file, err := os.Create(path.Join(targetDir, f.Name))
		if err != nil {
			msg := fmt.Sprintf("failed to create file for %s for transfer %s", f.Name, tMsg.TransferId)
			log.Err(err).Msg(msg)
			sendTransferUpdateMessage(ctx, tMsg.TransferId, &msg)
			return err
		}

		defer file.Close()

		_, err = io.Copy(file, rc)
		if err != nil {
			msg := fmt.Sprintf(
				"failed to write content for file %s to target file on mediahost for transfer %s",
				f.Name,
				tMsg.TransferId,
			)

			log.Err(err).Msg(msg)

			sendTransferUpdateMessage(ctx, tMsg.TransferId, &msg)
			return err
		}
	}

	// success
	sendTransferUpdateMessage(ctx, tMsg.TransferId, nil)

	return nil
}

func sendTransferUpdateMessage(ctx context.Context, transferId string, failureReason *string) {
	msg := messaging.TransferReadyUpdateMessage{
		TransferId: transferId,
	}

	if failureReason != nil {
		msg.Success = false
		msg.FailureReason = *failureReason
	} else {
		msg.Success = true

	}

	err := rabbitmq.PublishMessage(ctx, messaging.TopicTransferReadyUpdate, msg)
	if err != nil {
		log.Err(err).Msg("Failed to send transfer update message")
	}
}

func init() {
	rabbitmq.RegisterConsumer(handleTransferMessage, messaging.TopicTransferReady)
}
