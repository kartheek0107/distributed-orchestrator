# Distributed Task Orchestrator

A distributed system for running tasks across multiple worker nodes. Built this to understand how systems like Kubernetes, Celery, and Apache Airflow work under the hood.

**Status**: Work in progress - currently building out the core scheduler and worker communication

---

## What is this?

Imagine you have 100 data processing jobs that need to run. You could run them one by one on your laptop, but that would take forever. This system lets you distribute those jobs across multiple machines (workers) automatically.

Think of it like this:
- **Scheduler** = Boss who assigns work
- **Workers** = Employees who do the work  
- **Redis** = Whiteboard where everyone checks for new tasks
- **gRPC** = Fast internal communication (like Slack for machines)

## Why I built this

I wanted to understand:
1. How do distributed systems handle failures? (what if a worker crashes mid-task?)
2. How do you prevent race conditions when multiple workers try to grab the same job?
3. How does gRPC compare to REST for internal service communication?
4. How do production systems like Kubernetes orchestrate workloads?

## Current Features

- [x] Basic scheduler that accepts jobs via REST API
- [x] Workers that connect to scheduler via gRPC
- [x] Redis-based job queue (persistent across restarts)
- [x] Worker health monitoring with heartbeats
- [ ] Distributed locking (Redlock) - in progress
- [ ] Automatic job reassignment when workers fail
- [ ] Rate limiting and worker pooling

---

### The interesting parts

**1. Heartbeat mechanism**  
Workers send "I'm alive" signals every 3 seconds. If a worker doesn't respond for 15 seconds, the scheduler assumes it crashed and reassigns its jobs to other workers.

**2. Redlock for distributed locking**  
When multiple workers try to grab the same job, we need to make sure only one gets it. I'm implementing the Redlock algorithm which uses Redis to create locks that work across multiple machines.

**3. gRPC vs REST**  
- REST for job submission (simple, works everywhere)
- gRPC for internal communication (way faster, uses Protocol Buffers)

---

## Getting Started

### Prerequisites
- Go 1.21+
- Redis running locally (or use Docker)
- protoc (Protocol Buffer compiler) for generating gRPC code

### Quick setup
```bash
# Clone and setup
git clone https://github.com/kartheek0107/go-task-orchestrator.git
cd go-task-orchestrator
go mod download

# Generate gRPC code (only needed if you modify .proto files)
make proto

# Build everything
make build
```

### Running locally

The easiest way is using Docker Compose:
```bash
docker-compose up
```

This starts:
- 1 Redis instance
- 1 Scheduler
- 3 Workers

Or run manually:
```bash
# Terminal 1: Redis
redis-server

# Terminal 2: Scheduler
./bin/scheduler --config config/scheduler.yaml

# Terminal 3-5: Workers (run this 3 times with different IDs)
./bin/worker --config config/worker.yaml --id worker-1
./bin/worker --config config/worker.yaml --id worker-2
./bin/worker --config config/worker.yaml --id worker-3
```

### Submit a test job
```bash
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test_job",
    "command": "echo Hello from worker",
    "priority": 1,
    "timeout": 30
  }'
```

Check job status:
```bash
curl http://localhost:8080/api/v1/jobs/{job_id}
```

---

## Project Structure
go-task-orchestrator/
├── cmd/
│   ├── scheduler/      # Scheduler entry point
│   └── worker/         # Worker entry point
├── internal/
│   ├── scheduler/      # Scheduler logic (job distribution, monitoring)
│   ├── worker/         # Worker logic (task execution)
│   ├── queue/          # Redis queue implementation
│   ├── lock/           # Redlock for distributed locking
│   └── heartbeat/      # Health monitoring
├── api/proto/          # gRPC definitions
└── config/             # YAML configs

---

## What I learned

**Concurrency is hard**  
Initially tried using channels for everything, but realized RWMutex is better for shared state like the worker registry. Channels are great for passing data between goroutines, not so much for protecting shared memory.

**gRPC is awesome for internal services**  
Setting up protobuf definitions felt like extra work at first, but once it's done, the type safety and performance are worth it. Plus, auto-generated client/server code saves a ton of time.

**Distributed locking is tricky**  
The Redlock algorithm sounds simple (acquire lock on majority of Redis nodes), but implementing it correctly with timeouts, retries, and clock drift is surprisingly complex.

**Heartbeats > Ping/Pong**  
Instead of scheduler pinging workers (can overwhelm with many workers), workers send periodic heartbeats. Scheduler just tracks timestamps and marks workers dead if no heartbeat for N seconds. Much more scalable.

---

## Testing fault tolerance

Want to see the system handle failures? Try this:
```bash
# Start system with 3 workers
docker-compose up --scale worker=3

# Submit a long-running job
curl -X POST http://localhost:8080/api/v1/jobs \
  -d '{"name":"long_task","command":"sleep 30","timeout":60}'

# Kill a worker while it's processing
docker-compose kill worker-1

# Watch the scheduler reassign the job to another worker
docker-compose logs -f scheduler
```

The job should complete successfully even though we killed the worker mid-execution.

---

## Development
```bash
# Run tests
make test

# Run with race detector (catches concurrency bugs)
make test-race

# Build just the scheduler
make build-scheduler

# Format code
make fmt
```

---

## Roadmap

**Phase 1: Basic Communication** ✅
- [x] gRPC setup between scheduler and workers
- [x] Job submission REST API
- [x] Redis job queue

**Phase 2: Fault Tolerance** (Current)
- [x] Heartbeat monitoring
- [ ] Worker failure detection
- [ ] Job reassignment
- [ ] Redlock implementation

**Phase 3: Production Features** (Next)
- [ ] Rate limiting
- [ ] Worker pooling
- [ ] Job priorities
- [ ] Retry logic with exponential backoff

**Phase 4: Nice-to-haves**
- [ ] Task dependencies (DAG execution)
- [ ] Web dashboard for monitoring
- [ ] Metrics (Prometheus)
- [ ] Distributed tracing

---

## Resources I found helpful

- [Redis Distributed Locks](https://redis.io/docs/manual/patterns/distributed-locks/) - Redlock algorithm explanation
- [gRPC Go Tutorial](https://grpc.io/docs/languages/go/basics/) - Official gRPC guide
- [Designing Data-Intensive Applications](https://dataintensive.net/) - Chapter on distributed systems
- [Building Microservices](https://samnewman.io/books/building_microservices/) - Service communication patterns

---

## Contributing

This is a learning project, but if you spot bugs or have suggestions, feel free to open an issue or PR!

---


**Built by Kartheek Budime**  
[GitHub](https://github.com/kartheek0107) • [LinkedIn](https://linkedin.com/in/kartheek-budime)

Learning distributed systems one project at a time 
EOF