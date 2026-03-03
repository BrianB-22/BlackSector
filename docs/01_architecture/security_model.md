# Security Model Specification

## Version: 0.1
## Status: Draft
## Owner: Core Architecture

## Last Updated: 2026-03-02

---

# 1. Purpose

Defines the security model governing:

- Server authority
- Client trust boundaries
- Session isolation
- Command validation
- Rate limiting
- Data protection
- Exploit prevention

The security model protects simulation integrity and prevents state corruption.

---

# 2. Scope

## IN SCOPE

- Trust boundaries
- Authentication model
- Authorization rules
- Command validation requirements
- Rate limiting
- Replay protection
- Log security
- Exploit containment

## OUT OF SCOPE

- Network firewall configuration
- OS-level hardening
- Cloud provider IAM configuration
- Physical server security

---

# 3. Core Security Principles

- Server is authoritative.
- Client is untrusted.
- No client-calculated results accepted.
- All state mutation occurs inside tick.
- All commands validated before execution.
- No shared mutable state outside simulation.
- Fail closed, not open.

---

# 4. Trust Boundaries

Trust boundaries exist at:

1\. Network ingress

2\. Protocol parsing

3\. Command validation

4\. Tick execution

Everything before tick execution is considered untrusted input.

---

# 5. Authentication Model (v1)

Authentication may be:

- Username/password
- Token-based
- Local account store

Requirements:

- Passwords hashed (never plaintext)
- Authentication must occur before command submission
- Session must be bound to authenticated identity

No anonymous mutation of world state.

---

# 6. Authorization Model

Authorization rules:

- Session may only mutate its own ship
- Session may only issue commands permitted by state
- Admin commands restricted by role
- No cross-session command authority

Privilege levels:

- Player
- Admin
- System

Role enforcement must occur before command enqueue.

---

# 7. Command Validation

All commands must be validated in two stages:

Stage 1 – Application Layer:

- Envelope validation
- Schema validation
- Rate limit check
- Authentication check

Stage 2 – Tick Validation:

- Resource check (energy, heat)
- Cooldown check
- State validity check
- Range and engagement rules

Invalid commands:

- Rejected
- Logged
- Must not mutate state

---

# 8. Rate Limiting

Rate limiting required to prevent:

- Command flooding
- Denial-of-service via enqueue saturation
- Log flooding
- Resource exhaustion

Recommended constraints:

- Max commands per session per tick
- Max malformed messages per minute
- Max connection attempts per IP

Exceeding limits triggers:

- Warning
- Temporary throttle
- Session termination (if repeated)

---

# 9. Replay Protection

Command replay attacks must be mitigated by:

- Session-bound correlation IDs
- Rejecting duplicate correlation IDs
- Validating session state for every command

Tick timestamp must not be client-controlled.

---

# 10. Data Integrity Protection

The following are forbidden:

- Direct client writes to database
- Client-provided state overrides
- Client-provided damage values
- Client-provided economic values
- Client-provided random values

All critical values computed server-side.

---

# 11. Simulation Integrity Rules

Simulation integrity requires:

- Single-threaded tick
- Deterministic PRNG
- Immutable snapshot boundary
- No external system time usage
- No external service calls inside tick

Any violation risks state divergence.

---

# 12. Logging Security

Logs must not:

- Store plaintext passwords
- Store authentication tokens
- Expose internal memory dumps
- Expose secret seeds

Sensitive fields must be redacted.

---

# 13. Denial of Service Mitigation

Mitigation layers:

Transport Layer:

- Connection limits
- Read timeouts
- Message size limits

Protocol Layer:

- Envelope validation
- Max payload size

Application Layer:

- Rate limiting
- Session throttle
- Command queue cap

Simulation Layer:

- MaxCommandsPerTick
- Bounded queue

---

# 14. Crash Containment

If subsystem panic occurs:

- Halt simulation
- Preserve last snapshot
- Do not continue partially corrupted state

Never continue after undefined state mutation.

---

# 15. Persistence Security

Snapshot and event log must:

- Be write-protected
- Not allow partial overwrite
- Be validated on load
- Fail safely if corrupted

Recovery must verify:

- Snapshot integrity
- Log sequence integrity

---

# 16. Admin Security

Admin commands must:

- Require elevated authentication
- Be logged at INFO or higher
- Include session ID
- Include timestamp
- Include parameters used

Admin actions must be auditable.

---

# 17. Exploit Classes to Prevent

The system must explicitly defend against:

- Command flooding
- Energy bypass exploits
- Cooldown bypass exploits
- Duplicate command injection
- Tick desynchronization
- Replay injection
- PRNG manipulation
- Snapshot tampering

All mitigation must be server-side.

---

# 18. Security Testing Requirements

Security testing must include:

- Invalid command fuzzing
- Protocol envelope fuzzing
- Rate limit boundary tests
- Snapshot corruption simulation
- Replay integrity tests
- Concurrency race detection

Security tests must run in CI.

---

# 19. Non-Goals (v1)

- End-to-end encryption beyond SSH
- Zero-trust multi-node cluster
- Enterprise IAM integration
- Blockchain-based persistence
- Anti-cheat kernel drivers

---

# 20. Future Extensions

- Multi-factor authentication
- Role-based fine-grained permissions
- Intrusion detection hooks
- Secure admin audit viewer
- Distributed cluster security model

Security evolution must not compromise deterministic core.

---

# End of Document
