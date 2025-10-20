## 🏗️ Estructura del Proyecto

```
email-bot/
├── cmd/
│   └── email/
│       └── main.go
├── internal/
│   ├── core/
│   │   ├── domain/
│   │   │   ├── email.go
│   │   │   ├── auth.go
│   │   │   └── invoice.go
│   │   └── ports/
│   │       ├── email_service.go
│   │       ├── repositories.go
│   │       ├── auth_service.go
│   │       ├── invoice_service.go
│   │       └── execution_service.go
│   ├── application/
│   │   └── services/
│   │       ├── email_service.go
│   │       ├── auth_service.go
│   │       ├── invoice_service.go
│   │       └── execution_service.go
│   └── infrastructure/
│       ├── adapters/
│       │   ├── gmail_api/
│       │   │   └── gmail_client.go
│       │   ├── oauth2/
│       │   │   └── oauth2_manager.go
│       │   ├── repositories/
│       │   │   └── email_repository.go
│       │   └── http/
│       │       └── api_client.go
│       ├── delivery/
│       │   └── cli/
│       │       └── email_reader.go
│       └── persistence/
│           └── file_repository.go
├── config/
│   └── config.go
├── vendor/
│   ├── zip/
│   ├── pdf/
│   ├── xml/
│   └── json/
├── credentials.json
├── token.json
└── go.mod
```
