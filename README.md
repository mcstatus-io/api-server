# API Server
![](https://img.shields.io/github/languages/code-size/mcstatus-io/api-server)
![](https://img.shields.io/github/issues/mcstatus-io/api-server)
![](https://img.shields.io/github/actions/workflow/status/mcstatus-io/api-server/go.yml)

This is the REST server that powers the internal API for the website. This program is responsible for handling sign-ins for the website as well as generating access tokens used by the API. If you are looking for the program that retrieves server statuses, please check out [mcstatus-io/ping-server](https://github.com/mcstatus-io/ping-server) instead.

Please note that while this repository may seem to conform to some versioning standard, it most certainly does not. Updates are pushed at random, with no semantic versioning in place. Any update (also known as a *commit*) may suddenly break existing configurations without notice or warranty. If you run a privately hosted API server, please refer to the updated example configuration file before attempting to update to the latest commit. 

## Requirements

- [Go](https://go.dev/)
- [Redis](https://redis.io/)
- [GNU Make](https://www.gnu.org/software/make/)

## Getting Started

```bash
# 1. Clone the repository (or download from this page)
$ git clone https://github.com/mcstatus-io/api-server.git

# 2. Move the working directory into the cloned repository
$ cd api-server

# 3. Run the build script
$ make

# 4. Copy the `config.example.yml` file to `config.yml` and modify details as needed
$ cp config.example.yml config.yml

# 5. Start the development server
$ ./bin/main

# The server will be listening on http://localhost:3002 (default host + port)
```