# Welcome

[Stormkit](https://www.stormkit.io) is a hosting solution for seamless deployment and management of modern web applications.

![Stormkit](./.github/assets/deployment-page.png)

## Cloud Edition

For those who prefer a managed solution, Stormkit offers a Cloud Edition that can be accessed at [app.stormkit.io](app.stormkit.io). The Cloud Edition handles all the hosting, scaling, and maintenance tasks for you, allowing you to focus solely on building and improving your applications.

## Self-Hosted Edition

The Self-Hosted Edition of Stormkit gives you the flexibility to host your own instance of Stormkit on your infrastructure. This version is ideal for organizations that require more control over their hosting environment, need to comply with specific regulatory requirements, or prefer to manage their own infrastructure.

## Getting Started

To get started with the Self-Hosted Edition of Stormkit, you can choose to use either the provided binaries or Docker images.

### Using docker containers

You can use Docker images to run the Self-Hosted Edition. The following images are available:

- ghcr.io/stormkit-io/workerserver:latest
- ghcr.io/stormkit-io/hosting:latest

## Additional services

In addition to the Stormkit's microservices, a PostgreSQL database and a Redis Instance is also required for Stormkit to function properly.

## Local Development

To run Stormkit locally for development purposes:

### Prerequisites

- Go 1.25+
- Node.js 24+
- PostgreSQL 17+
- Redis 7+

### Local DNS

If you are using a MacOS, running the following script will set up a dns mask on your local environment
so that you can access stormkit through `http://stormkit` domain.

```bash
./scripts/dnsqmasq.sh
```

### Update environment variables

Copy [.env.example](./.env.example) and create an `.env` file. Provide the missing variables.

### Running the services

```bash
# Clone the repository
git clone https://github.com/stormkit-io/stormkit-io.git
cd stormkit-io

# Start all services (includes database setup and migrations)
./scripts/start.sh
```

## Project Structure

```
stormkit-io/
├── src/
│   ├── ce/                   # Community Edition (AGPL-3.0)
│   │   ├── api/              # REST API server
│   │   ├── hosting/          # Hosting service
│   │   ├── runner/           # Build and deployment runner
│   │   └── workerserver/     # Background job processing
│   ├── ee/                   # Enterprise Edition (Commercial)
│   │   ├── api/              # Enterprise API features
│   │   ├── hosting/          # Enterprise hosting features
│   │   └── workerserver/     # Enterprise background services
│   ├── lib/                  # Shared libraries and utilities
│   ├── migrations/           # Database migrations
│   ├── mocks/                # Test mocks and fixtures
│   └── ui/                   # Frontend React application
├── scripts/                  # Build and deployment scripts
```

### Component Overview

- **Community Edition (`src/ce/`)**: Open source components under AGPL-3.0
- **Enterprise Edition (`src/ee/`)**: Commercial features requiring a license
- **Shared Libraries (`src/lib/`)**: Common utilities used by both editions
- **Frontend (`src/ui/`)**: React-based web interface

## Testing

Tests require PostgreSQL with a test database named `sktest` and Redis to be running.

### Setup

```bash
# Start services
docker compose up -d db redis

# Create test database
docker compose exec db createdb -U ${POSTGRES_USER} sktest
```

### Running Tests

```bash
# Run all tests (sequential execution required)
go test -p 1 ./...

# With verbose output
go test -p 1 -v ./...

# With coverage
go test -p 1 -coverprofile=coverage.out ./...

# Custom timeout
go test -p 1 -timeout 30m ./...
```
