package media

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/egfanboy/mediapire-common/messaging"
	"github.com/egfanboy/mediapire-media-host/internal/app"
	"github.com/egfanboy/mediapire-media-host/internal/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

func sendTransferUpdateMessage(ctx context.Context, transferId string, failureReason *string) {
	msg := messaging.TransferUpdateMessage{
		TransferId: transferId,
	}

	if failureReason != nil {
		msg.Success = false
		msg.FailureReason = *failureReason
	} else {
		msg.Success = true
		msg.NodeId = app.GetApp().NodeId
	}

	err := rabbitmq.PublishMessage(ctx, messaging.TopicTransferUpdate, msg)
	if err != nil {
		log.Err(err).Msg("Failed to send transfer update message")
	}
}

func handleTransferMessage(ctx context.Context, msg amqp091.Delivery) error {
	var tMsg messaging.TransferMessage

	err := json.Unmarshal(msg.Body, &tMsg)
	if err != nil {
		msg := "failed to unmarshal transfer message"
		log.Err(err).Msg(msg)

		sendTransferUpdateMessage(ctx, tMsg.Id, &msg)
		return err
	}

	appInstance := app.GetApp()

	input, ok := tMsg.Inputs[appInstance.NodeId]
	if !ok {
		log.Info().Msg("Transfer has no inputs from this host")

		return nil
	}

	mediaService := NewMediaService()

	content, err := mediaService.DownloadMedia(ctx, input)
	if err != nil {
		msg := err.Error()
		sendTransferUpdateMessage(ctx, tMsg.Id, &msg)
		return err
	}

	file, err := os.Create(path.Join(appInstance.DownloadPath, tMsg.Id+".zip"))
	if err != nil {
		msg := err.Error()
		sendTransferUpdateMessage(ctx, tMsg.Id, &msg)
		return err
	}

	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		msg := "Failed to write content to file"
		log.Err(err).Msg(msg)

		sendTransferUpdateMessage(ctx, tMsg.Id, &msg)
		return err
	}

	err = file.Sync()
	if err != nil {
		log.Err(err)
		msg := err.Error()
		sendTransferUpdateMessage(ctx, tMsg.Id, &msg)
		return err
	}

	// Set a timer that will cleanup the content in 24 hours
	// TODO: use the actual expiry of the transfer
	time.AfterFunc(time.Hour*24, func() {
		mediaService.CleanupDownloadContent(ctx, tMsg.Id)
	})

	sendTransferUpdateMessage(ctx, tMsg.Id, nil)
	return nil
}

func handleDeleteMessage(ctx context.Context, msg amqp091.Delivery) error {
	var deleteMsg messaging.DeleteMediaMessage

	err := json.Unmarshal(msg.Body, &deleteMsg)
	if err != nil {
		msg := "failed to unmarshal delete message"
		log.Err(err).Msg(msg)

		return err
	}

	appInstance := app.GetApp()

	// Get the media for this node
	input, ok := deleteMsg.MediaToDelete[appInstance.NodeId]
	if !ok {
		log.Info().Msg("Delete request has no inputs from this host")

		return nil
	}

	mediaService := NewMediaService()

	err = mediaService.DeleteMedia(ctx, input)
	if err != nil {
		log.Err(err).Msg("Failed to delete all requested media")
	}

	return err
}

func handleUpdateMedia(ctx context.Context, msg amqp091.Delivery) error {
	log.Info().Msgf("Handling message for %s", messaging.TopicUpdateMedia)

	var message messaging.UpdateMediaMessage
	err := json.Unmarshal(msg.Body, &message)
	if err != nil {
		log.Err(err).Msg("failed to unmarshal message")

		return err
	}

	if items, ok := message.Items[app.GetApp().NodeId]; !ok {
		log.Info().Msgf("Ignoring message for %s since no updates are for this node.", messaging.TopicUpdateMedia)

		return nil
	} else {
		for _, item := range items {
			mediaService := NewMediaService()

			_, err := mediaService.UpdateItem(ctx, item.MediaId, item.Content)
			if err != nil {
				msg := fmt.Sprintf("failed to update item %s", item.MediaId)
				log.Err(err).Msg(msg)

				sendMediaUpdateMessage(ctx, message.ChangesetId, &msg)

				return err
			}

		}

		sendMediaUpdateMessage(ctx, message.ChangesetId, nil)
	}

	log.Info().Msgf("Finished handling message for topic %s", messaging.TopicUpdateMedia)
	return nil
}

func sendMediaUpdateMessage(ctx context.Context, changesetId string, failureReason *string) {
	msg := messaging.MediaUpdatedMessage{
		ChangesetId: changesetId,
		NodeId:      app.GetApp().NodeId,
	}

	if failureReason != nil {
		msg.Success = false
		msg.FailureReason = failureReason
	} else {
		msg.Success = true
	}

	err := rabbitmq.PublishMessage(ctx, messaging.TopicMediaUpdated, msg)
	if err != nil {
		log.Err(err).Msg("Failed to send media updated message")
	}
}

func init() {
	rabbitmq.RegisterConsumer(handleTransferMessage, messaging.TopicTransfer)
	rabbitmq.RegisterConsumer(handleDeleteMessage, messaging.TopicDeleteMedia)
	rabbitmq.RegisterConsumer(handleUpdateMedia, messaging.TopicUpdateMedia)
}
