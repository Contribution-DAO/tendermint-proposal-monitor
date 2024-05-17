# Proposal Monitor

Proposal Monitor is a Go-based application that monitors governance proposals on multiple blockchain networks and sends alerts to channels. It is designed to be lightweight and easy to deploy, with configuration options to customize the monitoring behavior for different chains.

## Features

- Monitor governance proposals on multiple blockchain networks
- Send alerts to channels
- Customizable behavior for alerting near the end of the voting period
- Validator vote status check

## Prerequisites

- Go 1.21.6 or later
- Docker (for building and deploying the application)

## Getting Started

### Configuration

Create a `config.yml` file in the `config/` directory with the following structure:

```yaml
# Global settings
validator_address: "your_validator_address_here"  # The address of the validator to monitor.
proposal_detail_domain: "https://www.mintscan.io"  # The base URL for viewing proposal details. This can be customized if you use a different domain.
voting_alert_behavior_nearing: "only_if_not_voted"  # Specifies when to send alerts near the end of the voting period. Options: "always", "only_if_not_voted".

# Chains to be monitored
chains:
  "ChainA":
    chain_id: "cosmoshub-4"  # The ID of the chain.
    sdk_version: "v1"  # The version of the Cosmos SDK API to use. Options are "v1" or "v1beta1".
    alerts:
      api_endpoint: "https://yourdomain.com"  # The API endpoint to fetch proposals.
      discord:
        webhook: "https://discord.com/api/webhooks/path"  # The Discord webhook URL to send alerts.

  "ChainB":
    chain_id: "cosmoshub-4"  # The ID of the chain.
    sdk_version: "v1beta1"  # The version of the Cosmos SDK API to use. Options are "v1" or "v1beta1".
    alerts:
      api_endpoint: "https://yourdomain.com"  # The API endpoint to fetch proposals.
      discord:
        webhook: "https://discord.com/api/webhooks/path"  # The Discord webhook URL to send alerts.
```

## Running the Application

You can run the application directly or using Docker.

### Running directly

1. Install Go 1.21.6 or later.
2. Clone the repository and navigate to the project directory.
3. Build and run the application:

    ```sh
    go build -o proposal_monitor
    ./proposal_monitor
    ```

### Running with Docker

1. Ensure you have Docker installed.
2. Build and run the Docker container:

    ```sh
    docker build -t proposal_monitor .
    docker run -p 3000:3000 proposal_monitor
    ```

### Using Docker Compose

You can also use Docker Compose to run the application along with any dependencies.

1. Ensure you have Docker Compose installed.
2. Create a `docker-compose.yml` file in the project directory with the following content:

    ```yaml
    version: '3.8'
    services:
      proposal_monitor:
        build: .
        ports:
          - "3000:3000"
        volumes:
          - ./config:/app/config
          - ./data:/app/data
    ```

3. Start the services:

    ```sh
    docker-compose up
    ```

### Testing with Mock Data

To test the application with mock data, use the `--mock` flag when running the application:

```sh
./proposal_monitor --mock
```