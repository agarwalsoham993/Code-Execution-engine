<div align="center">
  <h1>🚀 Code Execution Engine (v4)</h1>

  <p>
    <strong>A high-performance, self-hosted, scalable code execution platform and autograder written in Go.</strong>
  </p>
  
  <p>
    <a href="#features"><img src="https://img.shields.io/badge/Architecture-Go%20%2B%20Docker-blue?style=flat-square&logo=go" alt="Architecture"></a>
    <a href="#features"><img src="https://img.shields.io/badge/Queue-Redis-red?style=flat-square&logo=redis" alt="Redis"></a>
    <a href="#features"><img src="https://img.shields.io/badge/Database-PostgreSQL-336791?style=flat-square&logo=postgresql" alt="Postgres"></a>
    <a href="#features"><img src="https://img.shields.io/badge/Status-Production%20Ready-success?style=flat-square" alt="Status"></a>
  </p>
</div>

---

## 📖 Overview

The **Code Execution Engine** provides a secure, reliable sandboxing solution for running arbitrary, untrusted code. By spinning up rapidly provisioned Docker containers instead of relying on external API limits (e.g., JDoodle, Piston), this engine gives you total control, zero latency overhead, and unlimited scaling potential.

In **v4**, the platform evolves from a simple runner into a **feature-complete autograding platform**. It now features an interactive **Admin Portal**, dynamic input generation, Postgres-backed persistent question & submission storage, and robust Redis queueing.

## ✨ Key Features (v4 Architecture)

- 🔒 **True Sandboxing:** Uses native Docker container isolation with customizable timeouts and memory limits per job.
- ⚡ **Asynchronous Execution:** High-throughput `Redis` queues paired with dynamically scaling Go Goroutine worker pools.
- 🗄️ **Persistent Data Model:** `PostgreSQL` handles all domain models natively (Questions, Submissions, Admin test cases).
- 🛠️ **Integrated Admin Portal:** 
  - Effortlessly create new algorithmic challenges.
  - **Dynamic Input Generator:** Write Python scripts directly in the browser to auto-generate edge-case test packets.
  - Test and generate "golden" solution outputs straight from the UI.
- 📊 **Real-Time Log Streaming:** Stream asynchronous engine logs directly in the admin portal via a custom RingBuffer API.
- 👥 **Role-Based Isolation:** Admin sandboxing tests are deliberately excluded from standard user history.

---

## 🏗️ Architecture

 <img src="https://github.com/user-attachments/assets/b8216eba-7663-4418-a86c-9f47782ad3ea" width="800" alt="Code Execution Engine Logo">

The engine operates on a highly decoupled Client-Server and Worker-Queue model:

1. **API Server (Fiber):** Ingests raw code / generated scripts and schedules payloads.
2. **Datastore (Postgres):** Saves all questions, test cases, and submissions history.
3. **Queue (Redis):** Manages backpressure for incoming execution requests.
4. **Worker Pool (Manager):** Orchestrates the full Docker filesystem lifecycle—volume mounting, container launch, SIGKILL timeouts, and raw stdout/stderr collection.

### Core File Structure
- `cmd/main.go` — API entrypoint and Service injection.
- `internal/worker/worker.go` — Background job processing, test case generation, and file I/O bridging.
- `internal/sandbox/docker/provider.go` — Low-level Docker API mappings.
- `spec/spec.yaml` — Determines container image, run command, and entrypoint for specific languages. 

---

## 🚀 Quick Start Guide

### 1. Prerequisites
- **Go** (v1.21+)
- **Docker Engine** (Running autonomously in the background)
- **Docker Compose** (For Redis / Postgres provisioning)

```bash
# Example: Install Go (Ubuntu)
sudo apt install golang-go
```

### 2. Provision Infrastructure
Start Redis and PostgreSQL locally using the provided `docker-compose` recipe:

```bash
docker-compose up -d
```
> **Note:** This exposes Postgres on port `5432` and Redis on port `6379`.

### 3. Setup Environment Variables
By default, the engine connects using pre-defined local bindings. To override them, create a `.env` file in the root directory:

```env
RUNNER_API_BINDADDRESS=:8080
RUNNER_SANDBOX_TIMEOUTSECONDS=10 
RUNNER_SANDBOX_MEMORY=250M
RUNNER_DB_DSN=postgres://postgres:mysecretpassword@localhost:5432/runner?sslmode=disable
RUNNER_REDIS_ADDR=localhost:6379
```

### 4. Pull Container Languages
To eliminate delay on the first execution, pre-pull the supported languages:

```bash
docker pull python:alpine
docker pull node:alpine
docker pull golang:alpine
```

### 5. Start the Engine!
Fetch dependencies and start the high-performance Fiber API:

```bash
go mod tidy
go run ./cmd/main.go
```

You should see logs indicating successful bindings:
```text
INFO  Connected to Postgres 
INFO  Connected to Redis 
INFO  Starting Worker Pool (Min: 1, Max: 6) 
INFO  Code Runner Started on port :8080.
```

---

## 🎮 Usage 

Navigate to `http://localhost:8080` in your browser.

- **Solver Mode:** Standard coding interface to select questions, write solutions, and track your execution history.

![alt text](image.png)

- **Admin Portal:** Toggle via the top-right profile button. 

![alt text](image-1.png)

  - Navigate to the **Input Generator** left tab.
  - Write Python logic to yield array dumps (e.g. `print(json.dumps(["1 2", "3 4"]))`).
  - Run your generation, test against a golden solution in the center, and permanently save the algorithm to Postgres!
---
<div align="center">
  <i>Built with ❤️ using Go, Fiber, and Docker</i>
</div>


### 6.(Optional) Monitor REDIS and PostgresDB
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
