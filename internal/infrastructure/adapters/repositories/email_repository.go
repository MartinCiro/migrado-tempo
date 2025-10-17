package repositories

import (
	"email/internal/core/domain"
	"email/internal/infrastructure/adapters/imap"
)

type EmailRepository struct {
	imapClient *imap.IMAPClient
}

func NewEmailRepository(imapClient *imap.IMAPClient) *EmailRepository {
	return &EmailRepository{
		imapClient: imapClient,
	}
}

func (r *EmailRepository) Connect() error {
	if err := r.imapClient.Connect(); err != nil {
		return err
	}

	if err := r.imapClient.Login(); err != nil {
		return err
	}

	return r.imapClient.SelectMailbox("INBOX")
}

func (r *EmailRepository) Disconnect() error {
	return r.imapClient.Logout()
}

func (r *EmailRepository) SearchEmails(criteria domain.EmailCriteria) ([]domain.Email, error) {
	var uids []uint32
	var err error

	if criteria.UnreadOnly {
		uids, err = r.imapClient.SearchUnread()
	} else {
		// Implementar búsqueda general si es necesario
		uids, err = r.imapClient.SearchUnread() // Por ahora solo unread
	}

	if err != nil {
		return nil, err
	}

	if len(uids) == 0 {
		return []domain.Email{}, nil
	}

	imapMessages, err := r.imapClient.FetchMessages(uids)
	if err != nil {
		return nil, err
	}

	var emails []domain.Email
	for _, msg := range imapMessages {
		email := domain.Email{
			ID:      msg.SeqNum,
			Subject: msg.Envelope.Subject,
			Date:    msg.Envelope.Date,
			Read:    !containsFlag(msg.Flags, "\\Seen"),
		}

		if len(msg.Envelope.From) > 0 {
			email.From = msg.Envelope.From[0].Address()
		}

		emails = append(emails, email)
	}

	return emails, nil
}

func containsFlag(flags []string, flag string) bool {
	for _, f := range flags {
		if f == flag {
			return true
		}
	}
	return false
}
