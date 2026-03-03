# Tick Engine Specification

## Version: 0.1
## Status: Draft
## Owner: Core Architecture

## Last Updated: 2026-03-02

---

# 1. Purpose

Defines the authoritative simulation loop responsible for advancing the universe state.

The Tick Engine:

- Executes all world mutations
- Processes queued commands
- Resolves subsystem logic
- Emits structured events
- Maintains deterministic execution

It is the core of the server.

---

# 2. Scope

## IN SCOPE

- Tick lifecycle
- Execution order
- Command drain rules
- Subsystem resolution order
- Event emission timing
- Determinism guarantees
- Performance constraints

## OUT OF SCOPE

- Transport handling
- Session authentication
- Database schema
- Horizontal scaling

---

# 3. Design Principles

- Single-threaded authoritative execution
- Deterministic order of operations
- No blocking I/O
- Bounded tick duration
- Stable execution phases
- Explicit subsystem ordering

---

# 4. Tick Configuration

Configurable parameters:

- TickIntervalMs (default: 2000)
- MaxCommandsPerTick
- SnapshotIntervalTicks
- MaxTickOverrunMs

Tick duration must be configurable via server configuration.

---

# 5. Tick Lifecycle Overview

Each tick executes the following phases:

1\. Tick Start

2\. Command Drain

3\. Command Validation

4\. State Mutation

5\. Subsystem Resolution

6\. Event Emission

7\. Snapshot Trigger Check

8\. Tick End / Sleep

No phase may be reordered without explicit architectural change.

---

# 6. Detailed Tick Phases

## 6.1 Tick Start

- Record tick start timestamp
- Increment global tick counter
- Initialize event buffer

---

## 6.2 Command Drain Phase

- Drain command queue up to MaxCommandsPerTick
- Commands sorted by enqueue order
- No execution yet

If queue exceeds limit:

- Remaining commands processed next tick
- Optional warning emitted

---

## 6.3 Command Validation Phase

For each drained command:

- Validate session authorization
- Validate command schema
- Validate resource requirements
- Validate cooldown rules

Invalid commands:

- Emit command\_reject event
- Do not mutate state

---

## 6.4 State Mutation Phase

Valid commands are applied in deterministic order.

Examples:

- Initiate combat
- Fire weapon
- Start mining
- Activate scan
- Initiate jump
- Change engine burn

All mutations occur here.

No external calls permitted.

---

## 6.5 Subsystem Resolution Phase

Subsystems resolve in fixed order:

1\. Propulsion updates

2\. Detection \& scanning

3\. Combat resolution

4\. Mining extraction

5\. Hazard checks

6\. Exploration updates

7\. Economy adjustments

8\. Heat \& energy regeneration

9\. Status effect decay

Ordering is fixed to preserve determinism.

---

## 6.6 Event Emission Phase

After state fully resolved:

- Generate structured events
- Attach authoritative timestamp
- Assign to per-session outbound buffers
- Clear local event buffer

No network writes here.

---

## 6.7 Snapshot Trigger Phase

If tick % SnapshotIntervalTicks == 0:

- Dispatch async snapshot job
- Snapshot captures full world state
- Must not block tick

---

## 6.8 Tick End Phase

- Measure elapsed time
- If elapsed < TickIntervalMs:

       Sleep remaining duration

- If elapsed > TickIntervalMs:

       Record overrun metric

Tick must never overlap.

---

# 7. Deterministic Execution Rules

The following must be deterministic:

- Command processing order
- Subsystem resolution order
- Random number generation
- Event emission ordering

Random number generation must:

- Use seeded PRNG
- Seed derived from universe seed + tick counter
- Never use system time

Given identical inputs, simulation must reproduce identical outputs.

---

# 8. Time Model

Time in universe is defined by tick count.

All durations expressed in:

- Ticks
- Not milliseconds

Example:

- Weapon cooldown: 3 ticks
- Scan channel: 2 ticks
- Disable duration: 4 ticks

TickIntervalMs affects real-time pacing only, not logic.

---

# 9. Performance Constraints

Tick must:

- Complete under TickIntervalMs
- Target <50ms execution time at 100 players
- Maintain O(1) per-entity update complexity
- Avoid global recalculation loops

No allocation-heavy operations inside tick.

---

# 10. Overrun Handling

If tick execution exceeds TickIntervalMs:

- Log overrun event
- Continue immediately to next tick
- Do not skip ticks
- Do not execute multiple ticks simultaneously

Repeated overruns trigger warning state.

---

# 11. Error Handling

If panic occurs inside tick:

- Halt simulation
- Preserve last stable snapshot
- Write crash log
- Stop accepting commands

On restart:

- Load last snapshot
- Replay event log
- Resume tick

---

# 12. Integration Points

Depends On:

- Concurrency Model
- Command Queue
- Subsystem Specifications
- Persistence Model

Exposes:

- TickUpdateEvent
- SubsystemEvents
- SnapshotTriggerEvent

---

# 13. Non-Goals (v1)

- Variable-rate per-subsystem ticks
- Per-system tick partitioning
- Parallel subsystem execution
- Real-time microsecond simulation

---

# 14. Future Extensions

- Adaptive tick interval
- Partitioned regional ticks
- Event batching
- Load-based dynamic scheduling

Must preserve deterministic core.

---

# End of Document
