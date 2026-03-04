# Handshake Flow Specification
## Version: 0.1
## Status: Draft
## Owner: Core Architecture
## Last Updated: 2026-03-04

---

# 1. Purpose

Defines the connection negotiation process between client and server.

---

# 2. Connection Flow

Client connects via transport.

Server sends handshake_init.

Client responds with handshake_response.

Server accepts or rejects.

---

# 3. Successful Flow

Server → handshake_init  
Client → handshake_response  
Server → handshake_ack  

Session becomes ACTIVE.

---

# 4. Rejection Flow

Server → handshake_reject

Connection closed.

---

# End of Document