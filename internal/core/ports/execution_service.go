package ports

import (
	"context"
)

// ExecutionService define el orquestador principal que coordina todo el proceso
type ExecutionService interface {
	Run(ctx context.Context) error
	ProcessExistingFiles(ctx context.Context, token string) error
	ProcessEmails(ctx context.Context, nits []string, token string) error
	SendAlarm(ctx context.Context, token string, alarmData map[string]string) error
}
