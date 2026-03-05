# Message Format Specification

## Version: 0.1
## Status: Draft
## Owner: Core Architecture

## Last Updated: 2026-03-02

---

## 1. Purpose

Defines the universal message envelope format used for all client-server communication.

This specification governs:

- Message structure  
- Required and optional fields  
- Correlation rules  
- Timestamp authority  
- Payload discipline  
- Forward compatibility rules  

All protocol messages MUST conform to this format.

---

## 2. Scope

\*\*IN SCOPE:\*\*

- Envelope structure  
- Field requirements  
- Message typing  
- Correlation behavior  
- Timestamp model  
- Error formatting  
- Extensibility rules  

\*\*OUT OF SCOPE:\*\*

- Transport layer (SSH/TCP/Telnet)  
- Handshake semantics (see `handshake\_protocol.md`)  
- UI rendering behavior  
- Internal simulation logic  

---

## 3. Design Principles

- Single universal envelope  
- Server authoritative timestamps  
- Strict typing via `type` field  
- Correlation required for command-response  
- Clients must ignore unknown fields  
- Envelope must remain stable across subsystems  

---

## 4. Envelope Structure

All messages MUST follow this structure:

{

     "type": "<string>",

     "timestamp": <integer>,

     "correlation\_id": "<uuid\_or\_null>",

     "payload": { ... }

}

---

## 5. Field Definitions

### 5.1 type (Required)

- String  
- Lowercase snake\_case  
- Defines message category and behavior  

Examples:

- combat\_update  
- command\_accept  
- handshake\_init  

The `type` field uniquely determines the payload schema.

---

### 5.2 timestamp (Required)

- Integer  
- Server authoritative  
- Represents server tick number OR epoch time (defined globally in protocol spec)  

Clients MUST treat this as authoritative ordering.  

Clients MUST NOT trust their own clock for sequencing.

---

### 5.3 correlation\_id (Conditionally Required)

- UUID string  
- Required for:

     - command\_submit

     - command\_accept

     - command\_reject

     - Errors tied to a specific command  

May be null or omitted for:

- Broadcast events  
- Tick updates  
- System notifications  

---

### 5.4 payload (Required)

- JSON object  
- Must never be null  
- Must not be a primitive root type  

Payload must never contain presentation formatting.

---

## 6. Message Namespacing

Message types must follow:

<domain>\_<action>

Domains:

- system\_\*
- command\_\*
- combat\_\*
- mining\_\*
- exploration\_\*
- economy\_\*
- error\_\*

Examples:

- combat\_update
- mining\_yield
- system\_tick\_update

---

## 7. Command-Response Model

### 7.1 Command Submission (Client → Server)

{

     "type": "command\_submit",

     "timestamp": 0,

     "correlation\_id": "uuid-123",

     "payload": {

       "command": "fire\_weapon",

       "parameters": { ... }

     }

}

Client timestamp is ignored by the server.

---

### 7.2 Command Accept (Server → Client)

{

     "type": "command\_accept",

     "timestamp": 10452,

     "correlation\_id": "uuid-123",

     "payload": {

       "queued": true

     }

}

---

### 7.3 Command Reject (Server → Client)

{

     "type": "command\_reject",

     "timestamp": 10452,

     "correlation\_id": "uuid-123",

     "payload": {

       "reason": "Insufficient energy"

     }

}

---

## 8. Event Broadcast Model

Server-initiated events:

{

     "type": "combat\_update",

     "timestamp": 10453,

     "correlation\_id": null,

     "payload": {

       "tracking": 0.42,

       "heat": 35

     }

}

Broadcast events must not include correlation IDs.

---

## 9. Error Model

Errors must follow the envelope format:

{

     "type": "error\_invalid\_command",

     "timestamp": 10453,

     "correlation\_id": "uuid-123",

     "payload": {

       "reason": "Unknown command"

     }

}

Errors must not terminate session unless explicitly defined.

---

## 10. Extensibility Rules

- New fields may be added to payload  
- Clients must ignore unknown fields  
- Field removal requires major protocol version increment  
- Envelope field names are immutable  

Reserved top-level keys:

- type
- timestamp
- correlation\_id
- payload

No additional top-level keys permitted in v1.

---

## 11. Ordering Guarantees

- Messages are processed in arrival order per session  
- Tick-based events carry increasing timestamp values  
- Clients must use timestamp for ordering  

No guarantee of real-time network delivery ordering across sessions.

---

## 12. Security Rules

- Invalid envelope structure → rejection  
- Missing required fields → rejection  
- Non-JSON payload → rejection  
- Correlation ID spoofing → rejection  

Server validates every message before enqueue.

---

## 13. Performance Constraints

- Envelope parsing must be O(1)  
- No reflection-based dynamic schema resolution  
- Message dispatch must use type-based routing table  

---

## 14. TEXT Mode Adapter Behavior

TEXT mode:

- Receives structured envelope internally  
- Adapter renders ANSI output  
- Adapter never mutates envelope  
- Protocol semantics remain identical  

---

## 15. GUI Mode Behavior

GUI mode:

- Receives structured JSON directly  
- No ANSI formatting  
- No rendering instructions  
- Client responsible for presentation  

---

## 16. Non-Goals (v1)

- Binary serialization  
- Delta compression  
- WebSocket framing  
- Schema registry enforcement  
- Streaming partial payloads  

---

## 17. Future Extensions

- Binary protocol variant  
- Message batching  
- Delta state updates  
- Event subscription filtering  

---

# End of Document
