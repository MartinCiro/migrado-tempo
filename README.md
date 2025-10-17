## 🏗️ Estructura del Proyecto

```
play-go/
├── cmd/
│   ├── bot/
│   │   └── main.go                 # Bot standalone
│   └── tiktok-chat/
│       └── main.go                 # Integración TikTok
├── internal/
│   ├── core/                       # Lógica de negocio
│   │   ├── domain/                 # Entidades e interfaces
│   │   ├── application/            # Casos de uso y servicios
│   │   └── usecases/               # Lógica específica
│   └── infrastructure/             # Adaptadores externos
│       ├── adapters/
│       │   ├── providers/          # YouTube, Spotify (futuro)
│       │   ├── player/             # Reproductor ffplay
│       │   └── persistence/        # Almacenamiento en memoria
│       └── delivery/               # CLI y TikTok
├── pkg/
│   ├── ffmpeg/                     # Instalador automático
│   ├── logger/                     # Sistema de logging
│   └── utils/                      # Utilidades compartidas
└── tests/                          # Pruebas unitarias
```
