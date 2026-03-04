# GUI Adapter - Client Model Specification
## Version: 0.1
## Status: Draft
## Owner: Core Architecture
## Last Updated: 2026-03-04

---

# 1. Purpose

Defines the architectural model for graphical user interface (GUI) clients that connect to the headless simulation server.

The GUI client will be introduced in a future version of the system. This specification establishes the rules required to support graphical clients without altering the server’s simulation logic.

The GUI client consumes structured protocol messages and renders them locally using graphical assets and interface components.

The server does not render graphical interfaces.

---

# 2. Scope

IN SCOPE:

- GUI client connection model
- structured protocol message consumption
- graphical rendering responsibilities
- visual asset reference usage
- client-side UI architecture expectations

OUT OF SCOPE:

- GUI framework selection
- rendering engine implementation
- asset pipelines
- user interface design details
- animation systems

---

# 3. Design Principles

The GUI client must follow the core architectural principles:

- The server remains headless.
- All UI rendering occurs client-side.
- The server emits structured protocol messages only.
- The protocol defines data, not presentation.
- GUI clients interpret structured messages to render visuals.

The protocol acts as the **stable API** between simulation and graphical interface.

---

# 4. Architecture Model

The GUI client operates through the following pipeline:

Simulation Engine  
↓  
Structured Protocol Messages  
↓  
Transport Layer (TCP / SSH / future transports)  
↓  
GUI Client  
↓  
Local Rendering Engine  

The GUI client receives protocol messages and constructs the graphical interface locally.

---

# 5. Structured Message Consumption

GUI clients receive protocol messages in structured JSON format.

Example message:

```
{
  "type": "combat_update",
  "timestamp": 10453,
  "correlation_id": null,
  "payload": {
    "tracking": 0.42,
    "heat": 35,
    "shields": 78
  }
}
```

The client interprets the payload and updates graphical elements accordingly.

No UI formatting instructions are included in protocol messages.

---

# 6. Visual Asset References

The server provides **visual identifiers** rather than graphical assets.

Example:

```
"visual": "planet-desert-01"
```

The GUI client resolves these identifiers using its local asset library.

This ensures:

- stable protocol messages
- deterministic visual mapping
- minimal network bandwidth usage

Visual identifiers are defined in `visual_asset_reference.md`.

---

# 7. Rendering Responsibilities

The GUI client is responsible for:

- map rendering
- object visualization
- HUD displays
- menus and interface widgets
- animations
- camera control

The server provides data but does not control layout or presentation.

---

# 8. Client Interface Model

Typical GUI client interface components may include:

- system map
- ship status dashboard
- navigation display
- combat interface
- inventory and cargo panels
- exploration scanner display
- market interface

These components are derived from structured protocol messages.

---

# 9. GUI Generator Compatibility

The GUI adapter is designed to support future GUI generation systems.

Structured protocol messages can be used by automated tools to generate graphical dashboards or visualization interfaces.

Potential applications include:

- automated UI generation
- debugging visualization tools
- replay viewers
- development dashboards

The protocol remains independent of specific GUI implementations.

---

# 10. Integration with Protocol Layer

GUI clients operate in **GUI interface mode**, negotiated during the protocol handshake.

Example handshake request:

```
{
  "type": "handshake_response",
  "protocol_version": "1.0",
  "interface_mode": "GUI",
  "client_capabilities": []
}
```

Once GUI mode is selected:

- the server sends structured JSON messages
- no ANSI formatting is transmitted
- all rendering occurs client-side

---

# 11. State Management

GUI clients maintain a local representation of game state reconstructed from server events.

State updates are applied through protocol messages such as:

- ship_state_update
- navigation_update
- market_update
- combat_update

The server remains authoritative.

Clients must never assume authority over simulation state.

---

# 12. Security Considerations

GUI clients must obey the same security constraints as CLI sessions.

Clients cannot:

- modify world state directly
- inject unauthorized commands
- bypass command validation

All gameplay actions must be submitted through protocol commands.

---

# 13. Non-Goals (v1)

The GUI client specification does not define:

- specific graphical frameworks
- web client implementation
- asset rendering technology
- real-time graphics pipelines
- VR or AR interfaces

These decisions are intentionally deferred.

---

# 14. Future Extensions

Possible future enhancements include:

- desktop graphical clients
- web-based GUI clients
- replay visualization tools
- graphical debugging dashboards
- AI monitoring interfaces

All implementations must remain compatible with the protocol architecture.

---

# End of Document