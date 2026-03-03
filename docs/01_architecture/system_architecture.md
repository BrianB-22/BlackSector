# System Architecture Specification

## Version: 0.1
## Status: Draft
## Owner: Core Architecture

## Last Updated: 2026-03-02

---

# 1. Purpose

Defines the high-level structural architecture of the SpaceGame server.

This document describes:

- Major architectural layers
- Component boundaries
- Data flow
- Dependency direction
- Runtime topology

This is the structural blueprint of the platform.

---

# 2. Scope

## IN SCOPE

- Layered architecture model
- Component responsibilities
- Dependency rules
- Runtime component layout
- Data ownership boundaries

## OUT OF SCOPE

- Detailed subsystem math
- Deployment infrastructure
- Horizontal clustering
- DevOps tooling

---

# 3. Architectural Principles

- Simulation-first architecture
- Server authoritative core
- Strict domain separation
- Unidirectional dependency flow
- Adapter-based interface model
- Deterministic execution core

---

# 4. Layered Architecture Model

The system is structured in five primary layers:

1\. Transport Layer

2\. Protocol Layer

3\. Application Layer

4\. Simulation Core

5\. Persistence Layer

Dependency direction flows downward only.

---

# 5. Runtime Component Overview

High-level runtime structure:

Transport Layer  

           ↓  

Protocol Handler  

           ↓  

Session Manager  

           ↓  

Command Queue  

           ↓  

Tick Engine (Simulation Core)  

           ↓  

Event Dispatcher  

           ↓  

Session Outbound Buffers  

Persistence runs asynchronously beside the simulation core.

---

# 6. Layer Descriptions

---

## 6.1 Transport Layer

Responsibilities:

- Accept SSH connections
- Optional Telnet support
- Future direct TCP support
- Raw message read/write

Constraints:

- No simulation logic
- No world state mutation
- No command validation beyond envelope structure

---

## 6.2 Protocol Layer

Responsibilities:

- Parse message envelope
- Validate message structure
- Route by message type
- Enforce version compatibility

Constraints:

- Must not mutate world state
- Must not contain simulation logic
- Must remain UI-agnostic

---

## 6.3 Application Layer

Includes:

- Session Manager
- Command Validator
- Rate Limiter
- Authentication handler
- Admin CLI

Responsibilities:

- Manage session lifecycle
- Submit commands to queue
- Handle command correlation
- Enforce session policies

Constraints:

- No direct world mutation
- No cross-session shared state

---

## 6.4 Simulation Core

Includes:

- Tick Engine
- Combat System
- Mining System
- Exploration System
- Economy Engine
- Propulsion System
- Heat \& Energy Model
- Hazard System

Responsibilities:

- Own all mutable world state
- Execute deterministic updates
- Emit structured events
- Trigger snapshot operations

Constraints:

- Single-threaded execution
- No blocking I/O
- No external service calls

---

## 6.5 Persistence Layer

Includes:

- Snapshot writer
- Event log writer
- Recovery loader

Responsibilities:

- Periodic state snapshot
- Append-only event log
- Recovery from crash

Constraints:

- Must operate asynchronously
- Must not block tick engine

---

# 7. Domain Ownership Boundaries

World state ownership:

- Universe map → Simulation Core
- Player ships → Simulation Core
- Market state → Simulation Core
- Session authentication → Application Layer
- Socket connections → Transport Layer

No upward dependencies allowed.

Transport must never depend on Simulation Core internals.

---

# 8. Message Flow

Inbound Flow:

1\. Client sends message

2\. Transport receives raw data

3\. Protocol parses envelope

4\. Application validates command

5\. Command enqueued

6\. Simulation consumes command during tick

Outbound Flow:

1\. Simulation emits event

2\. Event placed in outbound buffer

3\. Session goroutine flushes to client

4\. Adapter renders TEXT or forwards JSON

---

# 9. World State Model

World state resides entirely inside Simulation Core memory.

State categories:

- Systems
- Entities
- Ships
- Markets
- Active engagements
- Mining fields
- Exploration anomalies

Session layer only sees event projections of state.

---

# 10. Determinism Boundary

Deterministic boundary exists at:

Tick Engine input (command queue)

Given identical:

- Initial snapshot
- Command order
- Tick interval

The Simulation Core must produce identical results.

Everything outside that boundary (transport, network timing) is nondeterministic but must not influence state.

---

# 11. Error Containment Strategy

Layered containment:

- Transport errors → session isolated
- Protocol errors → message rejected
- Application errors → command rejected
- Simulation panic → halt server
- Persistence failure → log and retry

No lower layer may corrupt upper layer state.

---

# 12. Scalability Model (v1)

Designed for:

- 50–100 concurrent players
- Single process
- Single simulation thread
- Vertical scaling

Future horizontal scaling would require:

- System partitioning
- Sharded economic engine
- Distributed event bus

Not part of v1.

---

# 13. Security Model (Architectural)

- No client authority
- No direct DB access
- Strict message validation
- Session isolation
- Rate limiting at application layer

World state mutation only via Tick Engine.

---

# 14. Observability Model

System must expose:

- Tick duration metrics
- Command queue depth
- Active sessions count
- Event emission count
- Overrun frequency
- Snapshot duration

Metrics collection must not block tick.

---

# 15. Non-Goals (v1)

- Microservices architecture
- Distributed clusters
- Multi-region failover
- Real-time physics engine
- Client-side prediction

---

# 16. Future Evolution

Architecture allows for:

- Regional simulation partitioning
- Protocol binary variant
- GUI client ecosystem
- Modifiable subsystem modules
- Replay debugging tools

Core invariant must remain:

Single authoritative deterministic simulation loop.

---

# End of Document
