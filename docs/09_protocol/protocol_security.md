# Protocol Security Specification
## Version: 0.1
## Status: Draft
## Owner: Core Architecture
## Last Updated: 2026-03-04

---

# 1. Purpose

Defines protocol-level protections against malicious clients.

---

# 2. Security Principles

Server authoritative simulation.

Clients cannot:

- modify world state
- spoof messages
- bypass validation

---

# 3. Validation Rules

Server validates:

- message structure
- required fields
- payload schema
- command authorization

---

# 4. Replay Protection

Correlation IDs prevent replay of command responses.

Timestamp ordering prevents stale message execution.

---

# End of Document