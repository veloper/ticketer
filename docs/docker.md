# Docker

## Pull and Run

```bash
docker pull veloper/ticketer
docker run -p 8300:8300 \
  -e TICKETER_ADMIN_USERNAME=admin \
  -e TICKETER_ADMIN_PAT=pat_admin \
  veloper/ticketer
```

## Build from Source

```bash
docker build -t ticketer .
```

## Docker Compose

```yaml
services:
  ticketer:
    image: veloper/ticketer
    ports:
      - "8300:8300"
    environment:
      TICKETER_ADMIN_USERNAME: admin
      TICKETER_ADMIN_PAT: pat_admin
    volumes:
      - ticketer-data:/data

volumes:
  ticketer-data:
```

## Using tktrctl in Compose

```bash
# One-off commands
docker compose exec ticketer tktrctl projects create "Game" GAME

# Automated setup service
docker compose --profile setup run setup
```

## Setup Service

```yaml
services:
  ticketer:
    image: veloper/ticketer
    ports: ["8300:8300"]
    environment:
      TICKETER_ADMIN_USERNAME: admin
      TICKETER_ADMIN_PAT: pat_admin
    volumes:
      - ticketer-data:/data

  setup:
    image: veloper/ticketer
    profiles: ["setup"]
    environment:
      TICKETER_HOST: http://ticketer:8300
      TICKETER_PAT: pat_admin
    depends_on:
      ticketer:
        condition: service_started
    command: >
      tktrctl projects create "Asteroid Game" ASTEROID-GAME &&
      tktrctl issues create ASTEROID-GAME "Fix login" --type bug --priority 1
```
