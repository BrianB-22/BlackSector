# Persistence Model Specification

## Version: 0.1
## Status: Draft
## Owner: Core Architecture

## Last Updated: 2026-03-02

---

# 1. Purpose

Defines how world state is persisted, recovered, and protected against data loss.

The persistence model must:

- Preserve deterministic simulation integrity
- Avoid blocking the tick engine
- Support crash recovery
- Enable replay debugging
- Maintain data consistency

Persistence is a durability mechanism, not an execution mechanism.

---

# 2. Scope

## IN SCOPE

- Snapshot model
- Event log model
- Crash recovery flow
- Async write model
- Data consistency guarantees
- Snapshot frequency rules

## OUT OF SCOPE

- Database vendor selection
- Backup infrastructure
- Cloud deployment strategy
- Horizontal replication

---

# 3. Design Principles

- Simulation must never block on disk I/O.
- Persistence must be append-only where possible.
- Snapshots must be consistent at tick boundary.
- Recovery must restore exact deterministic state.
- World state writes must be atomic.

---

# 4. Persistence Strategy Overview

The server uses a hybrid persistence model:

1\. Periodic full-state snapshots

2\. Append-only event log between snapshots

Structure:

Snapshot (full state at tick N)  

\+  

Event Log (ticks N+1 → current)  

=  

Recoverable deterministic state

---

# 5. Snapshot Model

## 5.1 Snapshot Trigger

Snapshot occurs when:

tick\_number % SnapshotIntervalTicks == 0

SnapshotIntervalTicks configurable (default: 300 ticks).

---

## 5.2 Snapshot Characteristics

- Full serialized world state
- Includes:

     - Universe map

     - All entities

     - Ships

     - Markets

     - Active engagements

     - Mining fields

     - Exploration data

     - PRNG state

     - Current tick number

- Written asynchronously
- Atomic write required

---

## 5.3 Snapshot Atomicity

Snapshot write must:

1\. Write to temporary file

2\. Flush to disk

3\. Rename atomically

No partial snapshots allowed.

---

# 6. Event Log Model

## 6.1 Purpose

Event log captures:

- All validated commands
- Deterministic simulation events (optional)
- Snapshot markers

Event log enables replay between snapshots.

---

## 6.2 Event Log Characteristics

- Append-only
- Ordered by tick
- Immutable
- Buffered writes
- Async flush

Log entry structure:

{

     "tick": <int>,

     "command": { ... }

}

Only validated commands are logged.

---

# 7. Recovery Model

On server startup:

1\. Locate latest snapshot

2\. Load snapshot into memory

3\. Restore PRNG state

4\. Replay event log entries after snapshot tick

5\. Resume tick loop

If no snapshot exists:

- Initialize new universe

Recovery must produce identical state to pre-crash state.

---

# 8. Determinism Requirements

To guarantee deterministic replay:

- PRNG seed stored in snapshot
- Command ordering preserved
- No time-based randomness
- No external system time usage in simulation

Replay must produce identical results.

---

# 9. Async Persistence Architecture

Tick Engine triggers persistence job.

Persistence Worker:

- Runs in separate goroutine
- Serializes snapshot
- Writes to disk
- Reports success/failure via channel

Tick Engine must never wait for persistence completion.

---

# 10. Failure Handling

## 10.1 Snapshot Failure

If snapshot fails:

- Log error
- Continue simulation
- Retry next interval

Simulation must not halt.

---

## 10.2 Event Log Failure

If event log write fails:

- Enter safe mode
- Stop accepting new commands
- Alert admin
- Preserve memory state

Event log corruption risks determinism.

---

# 11. Snapshot Interval Strategy

Tradeoff considerations:

Short interval:

- Faster recovery
- Higher disk usage

Long interval:

- Smaller disk footprint
- Longer replay time

Default target:

- 5–10 minutes equivalent tick duration

---

# 12. Performance Constraints

Persistence must:

- Never block tick
- Use bounded memory buffers
- Avoid large heap allocations inside tick
- Maintain constant-time snapshot trigger check

Event log flush batching recommended.

---

# 13. Data Integrity Guarantees

Persistence guarantees:

- No partial snapshot exposure
- No reordering of commands
- No mixed-tick snapshots
- No cross-tick state leakage

State captured at tick boundary only.

---

# 14. Storage Model (v1)

Recommended v1 storage approach:

- Flat file snapshot (binary or JSON)
- Append-only log file
- Local filesystem

No distributed database in v1.

---

# 15. Observability \& Monitoring

System must log:

- Snapshot duration
- Snapshot size
- Replay duration
- Event log write latency
- Disk error events

Metrics must not block simulation.

---

# 16. Non-Goals (v1)

- Real-time replication
- Multi-region failover
- Live hot-standby nodes
- Incremental snapshot diffing
- Database-backed entity persistence

---

# 17. Future Extensions

- Incremental snapshotting
- Snapshot compression
- Remote backup replication
- Multi-node state partitioning
- Live state streaming

All future evolution must preserve deterministic replay model.

---

# End of Document
