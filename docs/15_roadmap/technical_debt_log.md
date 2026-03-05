# Technical Debt Log

## Version: 0.1

## Status: Draft

## Owner: Core Architecture

## Last Updated: 2026-03-05

---

# 1. Purpose

Documents known technical debt accepted during development.

Each entry records:

* the simplification made
* why it was accepted
* the risk or limitation it introduces
* the planned resolution (if known)

Technical debt is not necessarily bad. Deliberate, documented debt is a normal part of phased development. Undocumented or unexamined debt is the problem.

---

# 2. Debt Entries

---

## TD-001: Single-Threaded Tick Engine

**Status:** Accepted for v1

**Description:**
The tick loop runs on a single goroutine. All simulation logic — navigation, combat, mining, economy, missions, AI traders — executes sequentially within each tick.

**Why Accepted:**
Keeps the simulation deterministic and simple. Eliminates the need for lock management across subsystems. Sufficient for 50–100 players at 2-second tick intervals.

**Risk:**
Tick duration will increase as player count and active entities grow. A single expensive operation (e.g., 80 concurrent combats) could cause slow ticks.

**Resolution Plan:**
Monitor tick performance via telemetry. If tick duration exceeds targets at scale, evaluate parallelizing independent subsystems (e.g., combat rounds for separate systems can be resolved independently). Any parallelization must preserve determinism.

**Target Milestone:** Address before scaling beyond 100 players (if needed).

---

## TD-002: Line-Delimited JSON Protocol

**Status:** Accepted for v1

**Description:**
The protocol uses line-delimited JSON (NDJSON). Each message is a single JSON object terminated by a newline.

**Why Accepted:**
Human-readable, debuggable, and easy to implement. Sufficient for 50–100 players at 2-second tick intervals.

**Risk:**
JSON parsing overhead per message. No compression. Larger messages than a binary protocol. Potential bottleneck at high player counts or with large payloads (e.g., full market listings).

**Resolution Plan:**
Profile message throughput. If bandwidth or CPU becomes a bottleneck, introduce optional binary framing or per-message gzip compression. Protocol versioning (MAJOR.MINOR) accommodates this without breaking clients.

**Target Milestone:** Evaluate at Milestone 4 load testing.

---

## TD-003: No Delta Compression in Tick Updates

**Status:** Accepted for v1

**Description:**
The server sends discrete event messages per tick, not compressed delta updates. Each client receives only messages relevant to their current context.

**Why Accepted:**
Simpler server-side logic. At 50–100 players with 2-second ticks, the total message volume is manageable.

**Risk:**
A busy tick with many simultaneous events (combat, trades, economic events) could produce a large number of messages to deliver. This could increase bandwidth and client processing time.

**Resolution Plan:**
If message volume becomes an issue, implement message batching per tick per session (wrap all messages for a tick in a single envelope). Deferred to Phase 2+ evaluation.

---

## TD-004: Static World Generation

**Status:** Accepted for v1

**Description:**
The game world (regions, systems, ports, jump connections) is loaded from a static database. There is no procedural generation.

**Why Accepted:**
Avoids procedural generation complexity in v1. The operator populates the world by seeding the database or using a world generation tool (out of scope for v1).

**Risk:**
Limited to the world size and structure set up by the operator. No dynamic world growth. All servers run the same world structure.

**Resolution Plan:**
A world generation tool can be built separately. The schema supports any world size within performance targets. Procedural generation is not on the roadmap but is architecturally possible.

---

## TD-005: No Session Reconnect State Replay

**Status:** Accepted for v1

**Description:**
When a player reconnects (new session_id), the server delivers a state snapshot of their current ship, location, and inventory. However, events that occurred during the disconnected window (market changes, NPC activity nearby, etc.) are not replayed.

**Why Accepted:**
Full event replay would require storing and indexing all events per player. The linger window (5 minutes default) limits the gap. Ship state is accurately restored.

**Risk:**
A player may miss a significant economic event, combat near their location, or mission expiry that occurred while they were disconnected. They see accurate current state but not the history.

**Resolution Plan:**
A "recent events" summary could be sent on reconnect (e.g., "A pirate was spotted near your location while you were away"). Logged as a potential UX improvement, not a critical bug.

---

## TD-006: SQLite Flushing Not Transactional Across Subsystems

**Status:** Accepted for v1

**Description:**
Dirty record flushing to SQLite at the end of each tick is not wrapped in a single atomic transaction across all subsystems. Subsystems flush their own dirty records independently.

**Why Accepted:**
Simplifies subsystem implementation. Each subsystem owns its persistence logic. In practice, a crash mid-flush is unlikely given 2-second tick intervals.

**Risk:**
In a crash mid-tick-flush, some subsystems may have persisted changes and others may not. This could create inconsistent state (e.g., credits deducted but cargo not added).

**Resolution Plan:**
Wrap all subsystem flushes in a single SQLite transaction at the persistence step of the tick loop. This is a low-effort improvement that should be done before Milestone 4.

**Target Milestone:** Milestone 3 or 4.

---

## TD-007: AI Trader Logic is Greedy (No Planning)

**Status:** Accepted for v1

**Description:**
AI traders select routes by evaluating the current best price differential. They do not model future price movements or coordinate with other traders.

**Why Accepted:**
Greedy route selection is simple to implement and sufficient to drive economic baseline behavior.

**Risk:**
Many AI traders may converge on the same high-value route simultaneously, causing rapid price equalization and reducing the value of the route. This may or may not be desirable economically.

**Resolution Plan:**
Introduce a diversity factor (jitter in route selection, region affinity, home port preference) to naturally spread traders. Evaluate behavior during Phase 2 playtesting.

---

## TD-008: No Command Authentication Replay Protection

**Status:** Accepted for v1

**Description:**
Commands submitted by clients include a `correlation_id` (UUID) but the server does not track previously seen IDs to prevent replay attacks.

**Why Accepted:**
Transport-level security (SSH, TLS) protects against replay at the network level. Within an authenticated session, replay is not a meaningful threat.

**Risk:**
If session tokens are somehow compromised, a replay attack within the session window is theoretically possible.

**Resolution Plan:**
Not planned. Transport-level security is the appropriate layer for replay protection. Documenting for completeness.

---

## TD-009: Mission Definitions Not Validated Against World State

**Status:** Accepted for v1

**Description:**
Mission JSON files are validated for structural correctness (objective types, reward fields) but not against live world state (e.g., that a referenced system_id exists in the database).

**Why Accepted:**
World validation would couple mission loading to database state, complicating startup order. Structural validation catches most authoring errors.

**Risk:**
A mission that references a non-existent system_id would fail silently when a player reaches that objective, resulting in a stuck mission.

**Resolution Plan:**
Add a post-load validation pass that checks referenced IDs against loaded world state. Run at server startup after world data is loaded. Emit warnings for invalid references.

**Target Milestone:** Milestone 3.

---

# 3. Resolved Debt

No entries yet. Items will be moved here when resolved.

---

# 4. Debt Review Schedule

Technical debt should be reviewed at each milestone boundary.

During Milestone 4 (Depth and Balance), all `CRITICAL` or `HIGH` severity items must be resolved before the milestone is considered complete.

---

# End of Document
