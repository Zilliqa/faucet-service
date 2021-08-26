# faucet-service

## Prerequisites

- [Golang](https://golang.org)
- [Docker](https://www.docker.com/community-edition)

## Environment Variables

Create a `.env` file with the following variables

| Name             | Description                 | Type         |
| ---------------- | --------------------------- | ------------ |
| ENV_TYPE         | Environment Type            | String       |
| RECAPTCHA_SECRET | reCAPTCHA secret            | String       |
| PRIVATE_KEY      | Service Account Private Key | SecureString |

## Installation and Usage

### `make deps`

Installs dependencies.

### `make build`

Builds the project.

### `make test`

Runs tests.

### `make cover`

Shows an HTML presentation of the source code decorated with coverage information.

### `make start`

Runs Docker container.

## API Documentation

View documentation on <a>https://editor.swagger.io/</a> by pasting openapi.yml from root directory (Recommended)
Or, install vscode plugins for openapi to be able to preview it using swaggerUI.
