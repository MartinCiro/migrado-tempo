## 🏗️ Estructura del Proyecto

```
email/
├── cmd/
│   └── email/
│       └── main.go
├── internal/
│   ├── core/
│   │   ├── domain/
│   │   │   ├── email.go
│   │   │   └── auth.go          # Nuevo
│   │   └── ports/
│   │       ├── email_service.go
│   │       ├── repositories.go
│   │       └── auth_service.go  # Nuevo
│   ├── application/
│   │   └── services/
│   │       ├── email_service.go
│   │       └── auth_service.go  # Nuevo
│   └── infrastructure/
│       ├── adapters/
│       │   ├── gmail_api/       # Cambiado de imap
│       │   │   └── gmail_client.go
│       │   ├── oauth2/          # Nuevo
│       │   │   └── oauth2_manager.go
│       │   └── repositories/
│       │       └── email_repository.go
│       └── delivery/
│           └── cli/
│               └── email_reader.go
├── config/
│   └── config.go
├── credentials.json             # Archivo de credenciales OAuth2
├── token.json                   # Token almacenado
└── go.mod                        # Pruebas unitarias
```
