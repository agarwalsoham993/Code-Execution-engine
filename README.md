# Code Execution Engine 

A scalable, self-contained Code Execution Engine written in Go. Its core function is to execute user-submitted code in secure, isolated Docker containers and return the results via a simple REST API.

The engine provides a reliable sandboxing solution for running arbitrary code, using docker containers. Without use of standard libraries like JUDGE 0 / API based code-execution engine PISTON.

# 💡 Architectural Overview

The engine operates on a client-server model. The API server (using Fiber since GOlang but RESTful) receives an execution request and hands it off to the Sandbox Manager, which orchestrates the entire Docker container lifecycle—from creation to code execution, timeout monitoring, and final cleanup.

<img width="1000" height="737" alt="image" src="https://github.com/user-attachments/assets/b8216eba-7663-4418-a86c-9f47782ad3ea" />

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

### Prerequisites Check
Go: A working Go installation (version 1.21+).
Docker: Docker Engine must be installed and running in the background.

```{bash}
sudo apt install golang-go {example for linux ubuntu}
```
for installing docker on your operating system follow instructions on -> https://docs.docker.com/engine/install/

### 1.Start infrastructure:
We need Redis (for the queue) and PostgreSQL (for the database) running.
Run this command in the project root:

```{bash}
docker-compose up -d
```

This will start the database on port 5432 and Redis on port 6379.

### 2.Configure Environment

Ensure you have the .env file created in the root directory.
In the current system there is no private password created for the in-docker postgres(mysecretpassword)
If you changed the password in docker-compose.yaml, make sure to update RUNNER_DB_DSN in the .env file.

### 3.Pull Language Images

Pre-Pull docker images  manually to save time on the first run:
```{bash}
docker pull python:alpine
docker pull node:alpine
docker pull golang:alpine
```

### 4.(Optional) Monitor REDIS and PostgresDB
In a NEW TERMINAL WINDOW (inside same directory)
check redis and postgres are running inside docker with `docker ps` in bash
<img width="1145" height="58" alt="image" src="https://github.com/user-attachments/assets/1b6f2bec-e743-4e72-a643-9c9e18110716" />

#### Turn on REDIS CLI 
`docker exec -it runner_queue redis-cli`

Type `MONITOR` inside bash , Live updates in REDIS server are logged.

#### Turn on the Postgres CLI in a NEW terminal window
`docker exec -it runner_db psql -U postgres -d runner`
Turns on the docker contained posgres CLI , type in `\dt` for viewing all the available databases inside.

<img width="309" height="110" alt="image" src="https://github.com/user-attachments/assets/5b583f39-c02a-41f3-bc2e-c88ca575b27b" />

Now type in `SELECT * FROM {name of the database}submissions;` for viewing the contents of the database

<img width="1467" height="137" alt="image" src="https://github.com/user-attachments/assets/8bc66cb8-b478-41c8-bcd7-cd7bc3f25ba7" />

### 5.Get Dependencies installed:
Open your terminal in the project's root directory and fetch the required Go modules.

`go mod tidy`

### 6.Start the Server:
Run the main application file. The server will automatically load any settings from a local .env file or use defaults, and start listening on (eg : http://localhost:8080.)

open the directory in terminal and run the following command from there:

`go run ./cmd/main.go`


### Output
```
INFO Connected to Postgres
INFO Connected to Redis
INFO Code Runner Started on port :8080 with 3 workers...
INFO Worker started, waiting for jobs... worker_id=1
INFO Worker started, waiting for jobs... worker_id=2
INFO Worker started, waiting for jobs... worker_id=3
```

### Interact as a user by going to http://localhost:8080 on supported browser 
Test the execution engine by inserting your desired code.
<img width="1844" height="823" alt="Screenshot from 2025-12-28 11-13-00" src="https://github.com/user-attachments/assets/b0eec81d-bf28-487d-bace-de0f2909a723" />

### 7.(Optional) Sharing your locally hosted html page online like for local hackathon
```{bash}
using Ngrok 
1. Login on Ngrok website and obtain authentication ID'S for setting up config file
2. Setup authentication detail in the laptop's own config file where NGROK was downloaded
3.Start your server on your local host
host your server online with -> ngrok http 8080
It will provide with a sharable local link hosting your primary page and using your server for testing codes
```
Ensure CORS service for the API is enabled : In our REST API its already enabled.

### 8.Close the server 
press Ctrl+C inside the same terminal to soft stop the process instead of abrupt closing of terminal.

close the dockerized postgres and redis ->`docker-compose down`
