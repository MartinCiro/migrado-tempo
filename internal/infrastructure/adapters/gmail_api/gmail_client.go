package gmailapi

import (
	"fmt"
	"log"

	"google.golang.org/api/gmail/v1"
)

type GmailClient struct {
	service *gmail.Service
}

func NewGmailClient(service *gmail.Service) *GmailClient {
	return &GmailClient{
		service: service,
	}
}

func (gc *GmailClient) ListUnreadMessages() ([]*gmail.Message, error) {
	// Buscar mensajes no leídos
	call := gc.service.Users.Messages.List("me").Q("is:unread")
	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve messages: %v", err)
	}

	var messages []*gmail.Message
	for _, msg := range response.Messages {
		message, err := gc.service.Users.Messages.Get("me", msg.Id).Format("metadata").Do()
		if err != nil {
			log.Printf("Error getting message %s: %v", msg.Id, err)
			continue
		}
		messages = append(messages, message)
	}

	return messages, nil
}

func (gc *GmailClient) GetMessageDetails(messageId string) (*gmail.Message, error) {
	return gc.service.Users.Messages.Get("me", messageId).Format("full").Do()
}
