# Protocol \& Interface Model Specification

## Version: 0.1
## Status: Draft
## Owner: Core Architecture

## Last Updated: 2026-03-02

---

# 1. Purpose

Defines how clients connect to the headless server and how interface modes are negotiated.

The server is UI-agnostic.

Transport and rendering are separated from simulation.

---

# 2. Design Principles

- Simulation never outputs raw UI formatting.
- All output structured internally.
- Text UI is an adapter.
- GUI client must negotiate capability.
- SSH is transport, not logic.

---

# 3. Connection Types

Supported transports:

- SSH (primary)
- Telnet (optional legacy)
- Direct TCP (future GUI)

Transport does not change simulation logic.

---

# 4. Interface Negotiation

Upon connection:

Server sends:

- ProtocolVersion
- SupportedInterfaceModes
- CapabilityFlags

Client responds with:

- DesiredInterfaceMode (TEXT | GUI)
- SupportedProtocolVersion
- CapabilityFlags

If GUI selected:

Server switches to structured message mode.

If TEXT selected:

Server enables terminal adapter.

---

# 5. TEXT Mode Requirements

TEXT mode must support:

- ANSI colors
- Cursor movement
- Rich formatting
- Box drawing
- Dynamic screen refresh
- Keyboard navigation

Text UI must feel modern and vibrant.

Text UI renders structured server messages.

---

# 6. GUI Mode Requirements

GUI mode:

- Receives structured data packets
- No ANSI codes
- No presentation formatting
- All UI rendered client-side
- Server sends pure data objects

Example:

{

     "type": "combat\_update",

     "tracking": 0.45,

     "heat": 32,

     "shields": 78

}

---

# 7. Message Model

All simulation output converted to structured internal messages.

Adapters transform messages into:

- ANSI-rendered text (TEXT mode)
- JSON or binary structured packets (GUI mode)

Simulation never outputs raw text.

---

# 8. Backward Compatibility

Protocol must:

- Be versioned
- Support negotiation fallback
- Allow GUI versions to verify compatibility

---

# 9. Non-Goals (v1)

- Embedded web server
- Browser client
- Mixed UI mode
- Client-side simulation

---

# End of Document
