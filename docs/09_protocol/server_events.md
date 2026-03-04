# Server Events Specification
## Version: 0.1
## Status: Draft
## Owner: Core Architecture
## Last Updated: 2026-03-04

---

# 1. Purpose

Defines events emitted by the simulation engine.

Events inform clients about world changes.

Events are broadcast when relevant.

---

# 2. Event Model

Events originate from simulation subsystems.

Example pipeline:

Simulation Engine  
↓  
Event Generated  
↓  
Protocol Message Created  
↓  
Broadcast to Clients

---

# 3. Event Categories

System events  
Combat events  
Exploration events  
Mining events  
Economy events  
Navigation events  

---

# 4. Example Event

{
 "type": "combat_update",
 "timestamp": 10453,
 "correlation_id": null,
 "payload": {
   "tracking": 0.42,
   "heat": 35
 }
}

---

# End of Document