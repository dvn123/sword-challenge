# Sword Challenge Server (SCS)

API server written in Go to manage tasks according to provided specification. Uses a MySQL database for persistence and RabbitMQ broker for task notifications.

## Dependencies

* Mysql
* RabbitMQ
* Docker (for testing and running compose)

## API

All APIs protected by authentication return 401 if no login is present and 403 if the role does not allow the operation. Every endpoint can return a 500 if there's a problem with the server or 400 in
the case of malformed input.

### Tasks API

Task ID should always be an integer for these APIs. Example valid task JSON:

````json
{
  "id": 1,
  "summary": "olaolaola",
  "completedDate": "2021-10-23T22:50:23Z",
  "user": {
    "id": 1,
    "username": "joel"
  }
}
````

| Method     | Path       |   Auth | Possible HTTP Responses                                    |
|----------|------------|---------------------------|-------------------------------|
| GET | `/api/v1/tasks` | Authenticated only.<br /> Own task or manager  | 200 + list of tasks of the authenticated user
| DELETE | `/api/v1/tasks/:task-id` |Authenticated only.<br /> Manager only. | 200 if task was deleted. <br/>404 if the task doesn't exist
| PUT | `/api/v1/tasks/:task-id` |Authenticated only.<br /> Own task or manager | 200 + updated task if it exists and user has permissions. <br/>404 if the task doesn't exist
| POST | `/api/v1/tasks` |Authenticated only.<br /> Own task or manager | 200 + task if task was created.

### Users API

| Method     | Path       | Auth | Description                           |
|----------|------------|--------|------------------------------|
| POST | `/api/v1/login` |Open | Mock login - returns a UUID if the user ID exists

## Usage

To run tests and check coverage run
`go test ./... -coverprofile='c.out' -coverpkg='./...' && go tool cover -func='c.out'`

Starting server and all dependencies:

```shell
docker-compose up --build
```

The application will create 2 users on startup, user with ID 1 is admin and the one with ID 2 is a technician.

### Demo

Login

```shell
curl -L -X POST 'http://localhost:8081/api/v1/login' -H 'Content-Type: application/json' --data-raw '{ "id": 2 }' -v
```

Copy cookie from login response header and replace $SCS_TOKEN `$SCS_TOKEN=$TOKEN_FROM_COOKIE`

Create a task:

```shell
curl -L -X POST 'http://localhost:8081/api/v1/tasks' -H "x-auth-token: $SCS_TOKEN" -H 'Content-Type: application/json' --data-raw '{ "summary": "olaolaola", "user": { "id": 1 } }' -v
```

View the task:

```shell
curl -L -X GET 'http://localhost:8081/api/v1/tasks' -H "x-auth-token: $SCS_TOKEN" -v
```

Complete the task:

```shell
curl -L -X PUT 'http://localhost:8081/api/v1/tasks/1' -H "x-auth-token: $SCS_TOKEN" -H 'Content-Type: application/json' --data-raw '{ "completedDate": "2021-10-23T22:50:23Z", "user": { "id": 1 } }' -v
 ```

Delete the task:

````shell
curl -L -X DELETE 'http://localhost:8081/api/v1/tasks/1' -H 'x-auth-token: $SCS_TOKEN' -v
````

# Comments

* Users API is only for testing, a user logging in with only an ID is obviously not ideal :D
* Server starts before RabbitMQ when using docker-compose, will fail a few times but eventually it will start

### Encryption

Summary is encrypted using AES-GCM-256 and a random IV since it contains PII. There are also no logs of the decrypted summary on the app. Every time a task is decrypted there is a log printed with the
authenticated user which functions as a sort of audit log.

### Notifications

For the notifications feature a RabbitMQ server was used. The message is published in the default exchange with routing key "tasks" and a queue consumes from the default exchange. This was the
simplest RabbitMQ setup which fulfilled the specification.

The queue is declared by the server on startup.

### Future Work

* Server should be more configurable in general and structure can be improved, structs should be used to pass configs, viper can be used to load the configs
* Go-Migrate pulls in way too many dependencies, would swap for another library as it makes the docker image large
* APIs should return an error object with details when an error occurs
* Do a general observability check, we have some logs already but would add traces and metrics
* Add Swagger/OpenAPI spec
* Support encryption key rotation

### Tests

* Mocking RabbitMQ in Go is not very easy, had to use test containers which require Docker on the test runner, would prefer to find another approach
* The AMQP client lib also panics instead of returning an error when the connection is not available, which makes it harder to test failures
* Because the task notification is only a log it has no side effects, makes it harder to test as well
* Security issues are checked thoroughly by auth + handler tests
* In spite of this we still have good coverage

### Kubernetes deployment

I used a simple helm chart with a service, deployment and hpa config that should work for this use case.

The deployment defines the deployment of the service as a pod, notables features are the HTTP health check probes. The service creates an IP for the set of pods so they are reachable internally. HPA
sets the deployment to auto scale when the CPU reaches 80%. There are 2 different configurations, one for QA and another for PRD which serve as an example of how to change values between environments.

It's also worth mentioning the server supports graceful shutdown, which is important in an environment where servers are ephemeral like k8s.

I focused on the deployment of the server itself, managing RabbitMQ or MySQL clusters from the repo of a server is not a good practice.

To view the resulting files run `helm template scs-chart --values scs-chart/values-qa.yaml` from the kubernetes directory.