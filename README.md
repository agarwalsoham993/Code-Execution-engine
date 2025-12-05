# Code Execution Engine 

A scalable, self-contained Code Execution Engine written in Go. Its core function is to execute user-submitted code in secure, isolated Docker containers and return the results via a simple REST API.

The engine provides a reliable sandboxing solution for running arbitrary code, using docker containers. Without use of standard libraries like JUDGE 0 / API based code-execution engine PISTON.

# 💡 Architectural Overview

The engine operates on a client-server model. The API server (using Fiber since GOlang but RESTful) receives an execution request and hands it off to the Sandbox Manager, which orchestrates the entire Docker container lifecycle—from creation to code execution, timeout monitoring, and final cleanup.

## Key Files:

`cmd/main.go` : Initializes all components (Config, Specs, Docker Provider) and starts the API server (:8080).

`internal/sandbox/manager.go` : Manages the full job lifecycle: file creation, container launch, execution, timeout, and cleanup.

`internal/sandbox/docker/provider.go ` : Handles low-level Docker API interactions (pulling images, mounting volumes, container execution).

`spec/spec.yaml` : YAML file defining execution commands and images for supported languages.

```
python3:
  image: "python:alpine"
  cmd: '/bin/sh -c "sleep 0.05; python3 main.py"'
  filename: "main.py"
  language: "python"
```

Update to a specific version `(e.g., golang:1.20-alpine)` or a custom image.

`internal/api/v1/routes.go` : Defines the routes: GET /v1/spec and POST /v1/exec.

## Environment Variables (e.g., in a .env file)

The configuration file `internal/config/envprovider.go` defines defaults, but the easiest way to change sandbox limits and server addresses is by creating a .env file in the project root.
```
Example .env file:
RUNNER_API_BINDADDRESS=:8080
RUNNER_SANDBOX_TIMEOUTSECONDS=20 # Reduce execution timeout to 10 seconds
RUNNER_SANDBOX_MEMORY=100M       # Increase memory limit for larger jobs
```


## Execution directions

Download the repository and save it as code-engine

### Prerequisites Check
Go: A working Go installation (version 1.21+).
Docker: Docker Engine must be installed and running in the background.

### Get Dependencies:
Open your terminal in the project's root directory and fetch the required Go modules.

`go mod tidy`

### Start the Server:
Run the main application file. The server will automatically load any settings from a local .env file or use defaults, and start listening on (eg : http://localhost:8080.)

open the directory in terminal and run the following command from there:

`go run ./cmd/main.go`


### Output
Code Runner Started on port 8080...

### Open the gui by double clicking on index.html
Test the execution engine by inserting your desired code.

### Close the server 
press Ctrl+C inside the same terminal to soft stop the process instead of abrupt closing of terminal.
