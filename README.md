# RabbitMQ Exclusive Queue Management with Go

This repository contains Go code demonstrating advanced RabbitMQ queue and channel management techniques, specifically
focusing on exclusive queues and their challenges. The code includes examples for handling channel closures, resource
cleanup, and deferred error handling, providing a robust foundation for building resilient distributed systems.

## Overview

RabbitMQ’s exclusive queues are powerful for isolating resources but come with challenges, particularly when multiple
consumers attempt to interact with them. This repository explores:

- **How RabbitMQ enforces exclusivity for queues.**
- **Monitoring unexpected channel closures using `NotifyClose`.**
- **Proper resource cleanup to prevent memory leaks and resource contention.**
- **Handling deferred errors when using RabbitMQ’s Go `amqp` library.**

These examples are designed to help developers working with RabbitMQ and Go implement reliable messaging systems.

## Features

- **Channel Lifecycle Monitoring**: Detect unexpected closures with RabbitMQ’s `NotifyClose`.
- **Exclusive Queue Enforcement**: Simulate scenarios where multiple consumers interact with exclusive queues.
- **Resource Cleanup**: Demonstrates proper cleanup mechanisms for RabbitMQ channels and connections.
- **Defensive Error Handling**: Examples of robust Go programming techniques for distributed messaging.
- **Structured Logging**: Clear, detailed logs for debugging RabbitMQ behaviors.

## Setup

### 1. Prerequisites

- Go installed.
- Docker & Docker Compose to run RabbitMQ locally.

### 2. Clone the Repository

```bash
git clone https://github.com/rozanecm/rabbit-go-test.git
cd rabbit-go-test
```

### 3. Start RabbitMQ

Use the included docker-compose.yml file to spin up RabbitMQ:

```bash
docker compose up -d
```

### 4. Configure RabbitMQ Connection

Update the config.toml file with your RabbitMQ URI:

```toml
[rabbitmq]
uri = "amqp://user:password@localhost:5672/"
```

## Usage

### Run the application:

```bash
go run main.go
```

What Happens:

1. Declares an exclusive queue.
2. Connects a first consumer to the queue.
3. Simulates a second consumer attempting to connect, triggering RabbitMQ’s exclusivity enforcement.
4. Monitors channel closures and logs related events.

## Code Structure

**Key Files**

- **main.go**: Entry point for running the experiment.
- **rabbit.go**: Core RabbitMQ functionality, including channel management and queue handling.
- **config.toml**: Configuration for RabbitMQ connection details.
- **docker-compose.yml**: Sets up a RabbitMQ instance locally for testing.

**Important Functions**

- **ConsumeFromQueue**: Handles exclusive queue declarations, bindings, and message consumption while implementing error
  handling and cleanup.
- **NotifyClose Usage**: Detects unexpected channel closures, such as exclusivity violations, for real-time handling.

### Logs and Debugging

The application uses logrus for structured logging. Logs include:

- Lifecycle events for queues and channels.
- RabbitMQ errors, such as `ACCESS_REFUSED` for exclusivity violations.
- Contextual information (e.g., file, line number) to help trace issues quickly.

# Contributing

Contributions are welcome! Feel free to open an issue or submit a pull request to improve the codebase or extend its
functionality.

# License

This repository is licensed under the MIT License.
