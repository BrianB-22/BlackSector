# Rate Limiting Specification
## Version: 0.1
## Status: Draft
## Owner: Core Architecture
## Last Updated: 2026-03-04

---

# 1. Purpose

Protects server from abuse and excessive command submission.

---

# 2. Limits

Server enforces:

Max commands per tick per session  
Max malformed messages per minute  

---

# 3. Enforcement

If exceeded:

error_rate_limited returned.

Repeated violations may terminate session.

---

# End of Document