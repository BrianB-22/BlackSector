# Tick Engine Model Specification
## Version: 0.1
## Status: Draft
## Owner: Core Architecture
## Last Updated: 2026-03-04

---

# 1. Purpose

Defines the server simulation timing model.

All game logic executes inside the tick engine.

---

# 2. Tick Model

Server runs simulation at fixed interval.

Example:

Tick interval: 2000ms

Each tick processes:

- queued commands
- simulation updates
- event generation

---

# 3. Tick Workflow

Command Queue  
↓  
Simulation Step  
↓  
Event Generation  
↓  
Protocol Messages  

---

# 4. Determinism

All simulation updates must be deterministic.

Tick ordering defines event ordering.

---

# End of Document