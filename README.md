# Proposal Monitor

Proposal Monitor is a Go-based application that monitors governance proposals on multiple blockchain networks and sends alerts to Discord channels. It is designed to be lightweight and easy to deploy, with configuration options to customize the monitoring behavior for different chains.

## Features

- Monitor governance proposals on multiple blockchain networks
- Send alerts to Discord channels
- Healthcheck support for monitoring application uptime
- Configurable check intervals and API endpoints

## Prerequisites

- Go 1.21.6 or later
- Docker (for building and deploying the application)

## Getting Started

### Configuration

Create a `config.yml` file in the `config/` directory with the following structure:

```yaml
# Global settings
check_interval: 600 # The interval (in seconds) at which the monitor checks for new proposals

# Healthcheck settings
healthcheck:
  enabled: no # Enable or disable healthcheck
  ping_url: https://hc-ping.com/path # URL to send pings for healthcheck
  ping_rate: 60 # Rate (in seconds) at which pings are sent

# Chains to be monitored
chains:
  "ChainA":
    chain_id: cosmoshub-4 # The ID of the chain
    alerts:
      api_endpoint: https://yourdomain/cosmos/gov/v1beta1/proposals # API endpoint to fetch proposals
      discord:
        enabled: yes # Enable or disable Discord alerts for this chain
        webhook: https://discord.com/api/webhooks/path # Discord webhook URL for this chain

  "ChainB":
    chain_id: cosmoshub-4 # The ID of the chain
    alerts:
      api_endpoint: https://yourdomain/cosmos/gov/v1beta1/proposals # API endpoint to fetch proposals
      discord:
        enabled: yes # Enable or disable Discord alerts for this chain
        webhook: https://discord.com/api/webhooks/path # Discord webhook URL for this chain
