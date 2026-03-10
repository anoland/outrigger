# Project Overview

This project, "outrigger", is a command-line tool written in Go for managing and rebalancing a Proxmox Virtual Environment (PVE) cluster. It utilizes the `go-proxmox` library to interact with the Proxmox API.

The primary functionality of this tool is to analyze the memory usage of the nodes in a Proxmox cluster and identify opportunities to improve the balance by migrating virtual machines (VMs) between nodes. The goal is to reduce the standard deviation of memory usage across the cluster, leading to a more stable and efficient environment.

The core logic for determining the rebalancing plan is contained in the `gemini_code.go` file. This logic can be run in "dry-run" mode, which will only print the proposed changes, or in "live" mode, which will execute the VM migrations.

# Building and Running

## Prerequisites

- Go 1.18 or later
- A Proxmox VE cluster

## Building

To build the executable, run the following command:

```sh
go build
```

## Running

To run the application, you will need to provide the following environment variables:

- `PVE_ENDPOINT`: The URL of your Proxmox API endpoint (e.g., `https://pve.example.com/api2/json`).
- `PVE_TOKEN_ID`: Your Proxmox API token ID.
- `PVE_TOKEN_SECRET`: Your Proxmox API token secret.

You can then run the application with the following command:

```sh
go run main.go
```

### Command-line Flags

- `-dry-run`: If set to `true`, the application will only print the proposed rebalancing plan without executing any migrations. This is the default behavior.
- `-threshold`: The minimum improvement in the cluster's imbalance score (standard deviation) required to trigger a VM migration. The default value is `0.15` (15%).

# Development Conventions

- The code is formatted using the standard Go formatting tools (`gofmt`, `goimports`).
- The project is organized into two packages:
    - `main`: Contains the main application entry point and Proxmox API client initialization.
    - `gemini`: Contains the core logic for cluster analysis and rebalancing.
- All dependencies are managed using Go modules.
