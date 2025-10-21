# Email Bot - Procesador Automático de Facturas

Bot desarrollado en Go con arquitectura hexagonal para procesar automáticamente emails con facturas electrónicas, extraer archivos ZIP adjuntos y enviar los datos a una API externa.

## 🏗️ Estructura del Proyecto

```
email-bot/
├── cmd/
│   └── email/
│       └── main.go                 # Punto de entrada de la aplicación
├── internal/
│   ├── core/
│   │   ├── domain/
│   │   │   ├── email.go           # Entidades: Email, Attachment, Criteria, Filter
│   │   │   ├── auth.go            # Entidades: Auth, Token, NIT
│   │   │   └── invoice.go         # Entidades: InvoiceData, InvoiceResponse
│   │   │   └── ai.go                    # Nuevo: Entidades para AI
│   │   └── ports/
│   │       ├── email_service.go   # Interfaz: EmailService
│   │       ├── repositories.go    # Interfaz: EmailRepository, FileRepository
│   │       ├── auth_service.go    # Interfaz: AuthService
│   │       ├── invoice_service.go # Interfaz: InvoiceService
│   │       └── execution_service.go # Interfaz: ExecutionService
│   │       └── ai_service.go        # Nuevo: Interfaz para AI
│   ├── application/
│   │   └── services/
│   │       ├── email_service.go   # Servicio: Lógica de aplicación para emails
│   │       ├── auth_service.go    # Servicio: Autenticación con API externa
│   │       ├── invoice_service.go # Servicio: Procesamiento de facturas XML
│   │       └── execution_service.go # Servicio: Orquestador principal
│   │       └── ai_service.go            # Nuevo: Servicio de AI
│   └── infrastructure/
│       ├── adapters/
│       │   ├── gmail_api/
│       │   │   └── gmail_client.go # Cliente Gmail API (EmailRepository)
│       │   ├── oauth2/
│       │   │   └── oauth2_manager.go # Manejo de autenticación OAuth2
│       │   ├── repositories/
│       │   │   └── email_repository.go # Adaptador de repositorio de emails
│       │   └── http/
│       │   │   └── api_client.go   # Cliente HTTP para APIs externas
│       │   └── ai/                      # Nuevo: Adaptador de AI
│       │       └── gemini_client.go     # Nuevo: Cliente Gemini
│       ├── delivery/
│       │   └── cli/
│       │       └── email_reader.go # CLI - Interfaz de línea de comandos
│       └── persistence/
│           └── file_repository.go  # Manejo de archivos locales
├── config/
│   └── config.go                  # Configuración de la aplicación
├── vendor/                        # Archivos temporales de procesamiento
│   ├── zip/                       # Archivos ZIP descargados de emails
│   ├── pdf/                       # PDFs extraídos de ZIPs (deXXXXXXXXXX.pdf)
│   ├── xml/                       # XMLs extraídos de ZIPs (adXXXXXXXXXX.xml)
│   └── json/                      # Archivos JSON temporales (NITs, cache)
├── credentials.json               # Credenciales OAuth2 de Gmail API
├── token.json                     # Token de autenticación OAuth2
├── prompts/                       # Nuevo: Prompts para Gemini
│   ├── invoice_processing.txt
│   └── xml_analysis.txt
└── go.mod                        # Dependencias del proyecto
```

## 📋 Funcionalidades Principales

### 🔍 Filtrado de Emails
- Busca emails no leídos en etiqueta específica "No eliminar/pb scrapping"
- Filtra por NITs configurados en la API externa
- Procesa solo emails que coinciden con los criterios de negocio

### 📎 Procesamiento de Adjuntos
- Descarga automáticamente archivos ZIP adjuntos
- Extrae archivos XML y PDF de los ZIPs
- Convierte archivos a base64 para envío a API
- Naming convention:
  - XML: `ad{id}.xml`
  - PDF: `de{id}.pdf`

### 🔄 Flujo de Procesamiento
1. **Autenticación** con API externa
2. **Obtención de NITs** desde API o cache local
3. **Procesamiento de archivos existentes** en vendor/zip/
4. **Filtrado y procesamiento de emails** de Gmail
5. **Envío de datos** a API externa
6. **Limpieza** de archivos temporales

## ⚙️ Configuración

### Variables de Entorno
```bash
export USER_API="tu_usuario"
export PASS_API="tu_password"
```

### Archivos de Configuración
- `credentials.json`: Credenciales OAuth2 de Gmail API
- `config.go`: Configuración de paths e intervalos

## 🚀 Uso

```bash
# Ejecutar la aplicación
go run cmd/email/main.go

# Modo debug
export DEBUG=true
go run cmd/email/main.go
```

## 🏗️ Arquitectura

### Hexagonal (Ports & Adapters)
- **Core**: Lógica de negocio y contratos
- **Application**: Casos de uso y servicios
- **Infrastructure**: Implementaciones concretas y adaptadores

### Patrones Implementados
- **Repository Pattern**: Abstracción del acceso a datos
- **Dependency Injection**: Inyección de dependencias en constructores
- **Service Layer**: Separación de responsabilidades

## 📊 Logs y Monitoreo

La aplicación incluye logs detallados para:
- Procesamiento de emails
- Descarga de adjuntos
- Conversión a base64
- Comunicación con APIs
- Errores y advertencias

## 🔧 Dependencias Principales

- **Gmail API**: Comunicación con Gmail
- **OAuth2**: Autenticación con Google
- **HTTP Client**: Comunicación con API externa
- **File Handling**: Procesamiento de archivos ZIP, XML, PDF

## 🗂️ Estructura de Archivos Temporales

```
vendor/
├── zip/     # ZIPs descargados (timestamp_nombre.zip)
├── pdf/     # PDFs extraídos (deXXXXXXXXXX.pdf)
├── xml/     # XMLs extraídos (adXXXXXXXXXX.xml)
└── json/    # Cache de NITs y respuestas API
```

Los archivos en `vendor/` son temporales y se limpian automáticamente después del procesamiento.