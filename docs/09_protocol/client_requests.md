# Client Requests Specification
## Version: 0.1
## Status: Draft
## Owner: Core Architecture
## Last Updated: 2026-03-04

---

# 1. Purpose

Defines the structure of commands sent from client to server.

Client requests represent player intent.

Server validates and executes commands within the simulation tick engine.

---

# 2. Command Envelope

Client commands use the standard message envelope.

Example:

{
 "type": "command_submit",
 "timestamp": 0,
 "correlation_id": "uuid",
 "payload": {
   "command": "scan",
   "parameters": {}
 }
}

Client timestamp is ignored.

---

# 3. Command Categories

Commands include:

Navigation commands  
Exploration commands  
Combat commands  
Mining commands  
Economy commands  

---

# 4. Example Commands

scan  
jump  
fire_weapon  
engage_autopilot  
start_mining  
market_buy  
market_sell  

---

# 5. Command Validation

Server must validate:

- command syntax
- player permissions
- resource availability
- cooldown rules

Invalid commands produce `command_reject`.

---

# End of Document