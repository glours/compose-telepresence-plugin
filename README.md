# compose-telepresence CLI
Simple command line which wraps Telepresence to test Compose `provider` services

## Overview

This CLI provides `compose up` and `compose down` commands which Docker Compose will run to:
- Setup the Telepresence configuration to the current Kubernetes context
- Create intercept for `provider` service defined in the Compose file
- Clean intercept and remove traffic manager on Compose down command

> **IMPORTANT: Demo Purpose Only**
> 
> This tool is a demonstration example of how to develop a Docker Compose provider. It is **not production ready** and has the following limitations:
> 
> - The Telepresence configuration is very basic
> - May not work with corporate proxies or custom network configurations
> - May require elevated/admin privileges to install the Telepresence traffic manager
> - Limited error handling and recovery options
> - No support for complex Kubernetes environments
> 
> Use at your own risk in development environments only.

## Build

```shell
make 
```

## Install
Change the directory where you want to install the binary in `Makefile`
```shell
make install
```

## Usage

Basic usage:
```shell
# Start services with telepresence
compose-telepresence compose up --name myservice

# Stop services and clean up telepresence
compose-telepresence compose down --name myservice
```

## Help
To see all available commands:
```shell
compose-telepresence help
```

## Requirements

- Docker and Docker Compose
- Kubernetes cluster with proper access configuration
- Telepresence CLI installed
- Helm CLI installed (used by Telepresence)
