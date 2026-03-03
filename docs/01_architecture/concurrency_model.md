# Concurrency Model Specification

## Version: 0.1
## Status: Draft
## Owner: Core Architecture

## Last Updated: 2026-03-02

---

# 1. Purpose

Defines the concurrency architecture for the SpaceGame server.

This specification governs:

- Simulation threading model
- Session concurrency
- Command ingestion
- Tick execution isolation
- Blocking I/O policy
- Synchronization guarantees

The goal is deterministic, scalable, race-free execution.

---

# 2. Scope

## IN SCOPE

- Simulation execution model
- Goroutine responsibilities
- Command queue design
- Event dispatch model
- Tick loop isolation
- Concurrency guarantees

## OUT OF SCOPE

- Transport encryption
- Database schema
- Deployment topology
- Horizontal sharding (future)

---

# 3. Design Principles

- Single authoritative simulation loop
- No shared mutable state outside tick
- Deterministic execution order
- Lock-free simulation core
- No blocking I/O inside tick
- Clear separation of concerns

---

# 4. High-Level Concurrency Architecture

The server uses a hybrid model:

- One authoritative simulation goroutine
- Multiple concurrent session goroutines
- Asynchronous I/O handling
- Lock-free command queue ingestion

Structure:

Sessions (goroutines)  

           ↓  

Command Queue (thread-safe)  

           ↓  

Simulation Tick Loop (single goroutine)  

           ↓  

Event Buffer  

           ↓  

Session Outbound Queues  

---

# 5. Simulation Thread Model

## 5.1 Authoritative Tick Goroutine

Exactly one goroutine runs the simulation loop.

Responsibilities:

- Process queued commands
- Advance world state
- Resolve combat
- Resolve mining
- Resolve exploration
- Update economy
- Emit events
- Schedule anomaly updates
- Manage persistence triggers

This goroutine is the only component allowed to mutate world state.

---

# 6. Tick Model

Tick duration is configurable.

Configuration parameter:

- TickIntervalMs

Default: 2000ms

Tick loop structure:

1\. Capture tick start time  

2\. Drain command queue  

3\. Validate commands  

4\. Apply state mutations  

5\. Resolve subsystems  

6\. Emit events to buffer  

7\. Schedule async persistence  

8\. Sleep until next tick  

Tick must complete within configured interval.

---

# 7. Session Concurrency Model (Go)

Each client connection runs in its own goroutine.

## Responsibilities

- Read incoming messages  
- Validate envelope format  
- Submit commands to thread-safe queue  
- Receive outbound events  
- Render TEXT mode or forward GUI messages  

## Constraints

- Must not mutate world state  
- Must not block simulation  
- Must not access shared simulation memory  

Session goroutines operate independently and communicate with the simulation only through the command queue and outbound event buffers.

---

# 8. Command Queue

The command queue must be:

- Thread-safe  
- Non-blocking for producers  
- Drained only by simulation goroutine  

Recommended implementation:

- Buffered channel  

OR  

- Lock-free ring buffer  

Rules:

- Commands timestamped upon enqueue  
- Commands validated inside tick  
- Command rate limiting per session enforced before enqueue  

---

# 9. World State Isolation

All mutable world state lives inside simulation goroutine.

Session goroutines may only access:

- Immutable snapshots  
- Event messages  
- Their own session state  

No shared state pointers across goroutines.

---

# 10. Event Dispatch Model

Simulation does not write directly to sockets.

Instead:

1\. Simulation writes events to per-session outbound buffers  

2\. After tick completes  

3\. Session goroutines flush outbound buffers asynchronously  

This prevents blocking I/O inside tick.

---

# 11. Blocking I/O Policy

Strict rule:

No blocking I/O inside simulation goroutine.

## Prohibited Inside Tick

- Network writes  
- Database writes  
- File writes  
- Log flushes  
- External service calls  

## Allowed Inside Tick

- Pure memory operations  
- Deterministic calculations  

Persistence must be delegated to async workers.

---

# 12. Persistence Model (Concurrency Aspect)

Snapshotting must:

- Be triggered inside tick  
- Be executed asynchronously  
- Never block simulation  

Event logs may be written via buffered async writer.

If persistence lags, simulation continues.

---

# 13. Failure Handling

If simulation goroutine panics:

- Server halts simulation  
- Latest snapshot retained  

On restart:

1\. Load last snapshot  

2\. Replay event log  

3\. Resume tick loop  

No partial state mutation allowed outside tick boundary.

---

# 14. Scalability Targets (v1)

Designed to support:

- 50–100 concurrent users  
- 500–1000 star systems  
- <50ms tick execution time  
- O(1) per-entity update  

Single-threaded simulation sufficient at this scale.

---

# 15. Determinism Guarantees

Given identical:

- Universe seed  
- Command order  
- Tick interval  

Simulation must produce identical results.

Concurrency must not introduce nondeterminism.

---

# 16. Race Condition Prevention

The following are forbidden:

- Shared mutable maps accessed by sessions  
- Direct session writes to world state  
- Cross-goroutine entity mutation  
- Simulation logic outside tick goroutine  

Use Go race detector in development.

---

# 17. Non-Goals (v1)

- Multi-threaded simulation partitioning  
- Per-system sharding  
- Horizontal scaling cluster  
- Distributed tick coordination  

These may be considered in v2+.

---

# 18. Future Expansion Path

If scaling beyond 200 users required:

Possible evolution paths:

- Regional partitioning  
- Per-system goroutines with strict boundaries  
- Sharded economic engine  
- Event bus architecture  

Must preserve deterministic simulation property.

---

# End of Document
