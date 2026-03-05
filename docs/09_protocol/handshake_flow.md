# Handshake Flow

See `handshake_protocol.md` for the full handshake specification including dual-port architecture, message formats, rejection flows, and session conflict handling.

## Quick Reference

```
SSH port (2222) — TEXT mode

  [SSH auth]
  Server → handshake_init   (interface_mode: TEXT)
  Client → handshake_response
  Server → handshake_ack    (session_id, tick_interval)
  [Session ACTIVE]


TCP port (2223) — GUI mode

  Server → handshake_init   (interface_mode: GUI)
  Client → handshake_response
  Server → handshake_ack    (session_id, tick_interval)
  [Session ACTIVE]


Rejection (either port):

  Server → handshake_reject
  [Connection closed]
```

Interface mode is determined by which port the client connects to — it is not negotiated.
