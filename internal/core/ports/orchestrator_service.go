package ports

import "context"

type OrchestratorService interface {
	RunFullProcess(ctx context.Context) error
	ProcessEmailsOnly(ctx context.Context) error
}
