# Project Overview

`losh` (Local Share) is a reverse proxy that allows you to expose an application running on your local machine to the Internet. It is designed to be lightweight and simple, similar in concept to `frp`.

## Core Features

- **Local Exposing**: Easily share local development services with the public internet.
- **Protocol Support**:
  - **TCP & UDP**: Expose any low-level network service.
  - **HTTP & HTTPS**: Forward web traffic to internal services via customized domain names.

## Architecture Concept

`losh` consists of two main components:
1. **Server**: Deployed on a public-facing server with a public IP/domain.
2. **Client**: Runs on your local machine, connecting to the remote server and forwarding local traffic.
