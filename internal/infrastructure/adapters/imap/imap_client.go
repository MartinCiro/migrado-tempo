package imap

import (
	"fmt"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

type IMAPClient struct {
	client   *client.Client
	server   string
	email    string
	password string
}

func NewIMAPClient(server, email, password string) *IMAPClient {
	return &IMAPClient{
		server:   server,
		email:    email,
		password: password,
	}
}

func (ic *IMAPClient) Connect() error {
	c, err := client.DialTLS(ic.server, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to IMAP server: %w", err)
	}

	ic.client = c
	return nil
}

func (ic *IMAPClient) Login() error {
	if ic.client == nil {
		return fmt.Errorf("client not connected")
	}

	if err := ic.client.Login(ic.email, ic.password); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	return nil
}

func (ic *IMAPClient) SelectMailbox(mailbox string) error {
	_, err := ic.client.Select(mailbox, false)
	return err
}

func (ic *IMAPClient) SearchUnread() ([]uint32, error) {
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{"\\Seen"}
	return ic.client.Search(criteria)
}

func (ic *IMAPClient) FetchMessages(uids []uint32) ([]*imap.Message, error) {
	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- ic.client.Fetch(seqset, []imap.FetchItem{
			imap.FetchEnvelope,
			imap.FetchFlags,
			imap.FetchBody,
		}, messages)
	}()

	var result []*imap.Message
	for msg := range messages {
		result = append(result, msg)
	}

	if err := <-done; err != nil {
		return nil, err
	}

	return result, nil
}

func (ic *IMAPClient) Logout() error {
	if ic.client != nil {
		return ic.client.Logout()
	}
	return nil
}
