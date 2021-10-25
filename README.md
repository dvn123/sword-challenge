# Sword Challenge Server (SCS)

API for managing Tasks and Users according to provided specification

## API

### Tasks

| Method     | Path       | Description                           |
|----------|------------|---------------------------------------|
| GET | `/api/v1/tasks/:task-id` |TODO |
| DELETE | `/api/v1/tasks/:task-id` |TODO |
| PUT | `/api/v1/tasks/:task-id` |TODO |
| POST | `/api/v1/tasks` |TODO |

### Users

| Method     | Path       | Description                           |
|----------|------------|---------------------------------------|
| POST | `/api/v1/login` |TODO |

## Dependencies

* Mysql
* RabbitMQ

## Usage

## Comments

# TODO

- [x] Docker
- [x] Setup Migrations
- [x] Gin Boilerplate
- [ ] API Errors
- [x] Logging
- [x] Authentication
- [ ] Metrics
- [ ] Traces
- [x] Healthcheck
- [x] Database Seeding
- [ ] Helm charts
- [x] Graceful shutdown
- [ ] Create task summary

# Users

- [x] API Routes
- [x] Setup basic tables
- [ ] Add password (hash + salt)
- [ ] Add roles routes
- [ ] Implement role APIs
- [ ] Implement user APIs
- [ ] Add auth to APIs
- [x] Login endpoint

# Tasks

- [x] API Routes
- [x] Setup tables
- [x] Implement APIs
- [x] Create Notification system
