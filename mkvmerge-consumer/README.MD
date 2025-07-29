                    MKV MEDIA AUTOMATION SYSTEM
                    ============================

┌─────────────────┐    ┌───────────────────────────────────────────────────┐
│   External      │    │                RabbitMQ Server                    │
│   Message       │───▶│                x.x.x.x:5672                       │
│   Producer      │    │         vhost: vhost-media-automation             │
└─────────────────┘    └───────────────────────────────────────────────────┘
                                              │
                                              ▼
         ┌────────────────────────────────────────────────────────────────────┐
         │                        QUEUE SYSTEM                                │
         │                                                                    │
         │ ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐  │
         │ │ mkvmerge.tasks  │    │  mkvmerge.done  │    │ Dead Letter     │  │
         │ │ (Input Queue)   │    │ (Output Queue)  │    │ Queues (DLQs)   │  │
         │ │                 │    │                 │    │                 │  │
         │ │ Messages:       │    │ Messages:       │    │ Failed messages │  │
         │ │ {               │    │ {               │    │ with error info │  │
         │ │   torrentName,  │    │   filename,     │    │                 │  │
         │ │   category      │    │   status,       │    │                 │  │
         │ │ }               │    │   time          │    │                 │  │
         │ └─────────────────┘    │ }               │    │                 │  │
         │                        └─────────────────┘    └─────────────────┘  │
         └────────────────────────────────────────────────────────────────────┘
                    │                              │
                    ▼                              ▼
    
    ┌─────────────────────────────────┐    ┌─────────────────────────────────┐
    │        MKV CONSUMER             │    │        MKV NOTIFIER             │
    │     (mkvmerge-consumer)         │    │     (mkvmerge-notifier)         │
    │                                 │    │                                 │
    │ Consumes: mkvmerge.tasks        │    │ Consumes: mkvmerge.done         │
    │                                 │    │                                 │
    │ PROCESSING FLOW:                │    │ NOTIFICATION FLOW:              │
    │ ┌─────────────────────────────┐ │    │ ┌─────────────────────────────┐ │
    │ │ 1. Receive message          │ │    │ │ 1. Receive completion msg   │ │
    │ │ 2. Parse JSON               │ │    │ │ 2. Parse JSON               │ │
    │ │ 3. Map category to path     │ │    │ │ 3. Format notification      │ │
    │ │ 4. Find folder              │ │    │ │ 4. Send to Telegram         │ │
    │ │ 5. Locate .mkv files        │ │    │ │ 5. Acknowledge message      │ │
    │ │ 6. For each .mkv:           │ │    │ └─────────────────────────────┘ │
    │ │    ├─ Get track info        │ │    │                                 │
    │ │    ├─ Check languages       │ │    │ ERROR HANDLING:                 │
    │ │    ├─ Filter tracks         │ │    │ • JSON parsing errors → DLQ     │
    │ │    ├─ Run mkvmerge          │ │    │ • Missing fields → DLQ          │
    │ │    └─ Replace original      │ │    │ • Telegram failures → DLQ       │
    │ │ 7. Publish to done queue    │ │    │                                 │
    │ │ 8. Acknowledge message      │ │    │ TELEGRAM OUTPUT:                │
    │ └─────────────────────────────┘ │    │ 🎬 MKV Processing Complete      │
    │                                 │    │ 📁 File: filename.mkv           │
    │ TRACK FILTERING:                │    │ ✅ Status: processed            │
    │ • Keep all video tracks         │    │ 🕒 Completed: timestamp         │
    │ • Keep English audio only       │    │ 📂 Full Path: /path/to/file     │
    │ • Keep English subtitles        │    │                                 │
    │ • If no ENG audio, keep all     │    │                                 │
    │                                 │    │                                 │
    │ CATEGORY MAPPINGS:              │    │                                 │
    │ • local-movies  → /mnt/vault/   │    │                                 │
    │ • local-tvshows → media/jello/  │    │                                 │    │                                 │    │                                 │
    │ ERROR HANDLING:                 │    │                                 │
    │ • JSON parsing errors → DLQ     │    │                                 │
    │ • Unknown categories → DLQ      │    │                                 │
    │ • Missing folders → DLQ         │    │                                 │
    │ • No successful processing →    │    │                                 │
    │   message stays in queue        │    │                                 │
    └─────────────────────────────────┘    └─────────────────────────────────┘
                    │                                              │
                    ▼                                              ▼
    ┌─────────────────────────────────┐                          ┌─────────────────┐
    │        FILE SYSTEM              │                          │   TELEGRAM      │
    │                                 │                          │     BOT         │
    │ Media Storage Paths:            │                          │                 │
    │ • /mnt/vault/media/jello/       │                          │ Sends formatted │
    │ • /mnt/vault-media/             │                          │ notifications   │
    │                                 │                          │ to configured   │
    │ Operations:                     │                          │ chat/user       │
    │ • Scan for .mkv files           │                          │                 │
    │ • Analyze track information     │                          │                 │
    │ • Create temporary files        │                          │                 │
    │ • Replace originals             │                          │                 │
    │ • Preserve directory structure  │                          │                 │
    └─────────────────────────────────┘                          └─────────────────┘

                              SYSTEM FEATURES
                              ===============
    • Configuration via godotenv and viper (environment variables and config files)
    • Dead Letter Queue (DLQ) support for failed messages
    • Graceful shutdown handling (SIGINT/SIGTERM)
    • Message persistence and durability
    • QoS settings for controlled processing
    • Comprehensive error logging

## Configuration

This service uses a combination of config files (YAML) and environment variables for configuration:

### Configuration Methods (in order of precedence)

1. **Environment Variables**: Highest precedence
2. **`.env` File**: Loaded at startup if present
3. **Configuration Files**: `config.yaml` in the following locations:
   - Current directory (`./config.yaml`)
   - Config directory (`./config/config.yaml`)
   - System config directory (`/etc/mkvmerge-consumer/config.yaml`)
4. **Default Values**: Used when no other configuration is provided

### RabbitMQ Configuration

```yaml
# config.yaml
rabbitmq:
  host: "10.10.40.19"
  port: "5672"
  username: "mkvmerge-consumer"
  password: "mkvmerge-consumer"
  vhost: "media-automation"
  queue:
    tasks: "mkvmerge.tasks"
    done: "mkvmerge.done"
    dlq: "mkvmerge.tasks_DLQ"
```

**Corresponding Environment Variables:**
```
RABBITMQ_HOST=10.10.40.19
RABBITMQ_PORT=5672
RABBITMQ_USERNAME=mkvmerge-consumer
RABBITMQ_PASSWORD=mkvmerge-consumer
RABBITMQ_VHOST=media-automation
RABBITMQ_QUEUE_TASKS=mkvmerge.tasks
RABBITMQ_QUEUE_DONE=mkvmerge.done
RABBITMQ_QUEUE_DLQ=mkvmerge.tasks_DLQ
```

### File Path Configuration

```yaml
# config.yaml
paths:
  categories:
    local-movies: "/mnt/vault/media/jello/movies"
    local-tvshows: "/mnt/vault/media/jello/tvshows"
```

**Corresponding Environment Variable:**
```
# JSON string format
PATHS_CATEGORIES='{"local-movies":"/mnt/vault/media/jello/movies","local-tvshows":"/mnt/vault/media/jello/tvshows"}'
```

## Docker Usage

### Docker Compose

The easiest way to run this service is using Docker Compose:

```bash
# Start the service
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the service
docker-compose down
```

### Manual Docker Build and Run

```bash
# Build the Docker image
docker build -t mkvmerge-consumer .

# Run the container
docker run -d \
  --name mkvmerge-consumer \
  -v /mnt/vault/media:/mnt/vault/media:rw \
  -v $(pwd)/config:/app/config:ro \
  -e RABBITMQ_HOST=10.10.40.19 \
  mkvmerge-consumer
```