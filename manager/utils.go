package manager

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/event-subscriber/events"
	"github.com/rancher/go-rancher/v3"
)

var (
	ErrTimeout = errors.New("Timeout waiting")
)

func wrapHandler(handler func(event *events.Event, apiClient *client.RancherClient) (*client.Publish, error)) func(event *events.Event, apiClient *client.RancherClient) error {
	return func(event *events.Event, apiClient *client.RancherClient) error {
		publish, err := handler(event, apiClient)
		if err != nil {
			return err
		}
		if publish != nil {
			return reply(publish, event, apiClient)
		}
		return nil
	}
}

func emptyReply(event *events.Event) *client.Publish {
	return &client.Publish{
		PreviousId: event.ID,
		Name:       event.ReplyTo,
	}
}

func createPublish(response *client.DeploymentSyncResponse, event *events.Event) *client.Publish {
	reply := emptyReply(event)
	reply.Data = map[string]interface{}{
		"deploymentSyncResponse": response,
	}
	return reply
}

func reply(publish *client.Publish, event *events.Event, apiClient *client.RancherClient) error {
	log.Infof("Reply: %+v", publish)

	_, err := apiClient.Publish.Create(publish)
	if err != nil {
		return fmt.Errorf("Error sending reply %v: %v", event.ID, err)
	}

	return nil
}
