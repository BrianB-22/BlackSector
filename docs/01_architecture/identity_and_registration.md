# Identity \& Registration Specification

## Version: 0.1
## Status: Draft
## Owner: Core Architecture

## Last Updated: 2026-03-02

---

# 1. Purpose

Defines the user identity, registration, and authentication model for SpaceGame.

This specification governs:

- Account creation
- Credential storage
- Authentication flow
- Session binding
- Identity persistence
- Security controls

Identity is separate from simulation state but required before gameplay access.

---

# 2. Scope

## IN SCOPE

- User registration flow
- Credential hashing
- Login authentication
- Session binding
- Account persistence
- Role assignment (Player/Admin)
- Registration rate limiting

## OUT OF SCOPE

- OAuth integration
- Third-party SSO
- Email verification system (v1 optional)
- Account recovery flows (future)

---

# 3. Design Principles

- Identity is server-controlled.
- Passwords are never stored in plaintext.
- Authentication must occur before command submission.
- Identity must be auditable.
- Account system must not block tick engine.
- Registration must resist abuse.

---

# 4. Account Model

Each account contains:

- account\_id (UUID)
- username (unique, case-insensitive)
- password\_hash
- password\_salt (if applicable)
- role (Player | Admin)
- created\_at (UTC timestamp)
- last\_login\_at (UTC timestamp)
- account\_status (Active | Suspended | Disabled)

Accounts are persisted independently of world state.

---

# 5. Credential Storage

Passwords must be:

- Hashed using a modern adaptive hashing algorithm (e.g., bcrypt or argon2)
- Salted automatically by algorithm
- Never logged
- Never returned to client

Minimum password policy (v1 baseline):

- 8+ characters
- Configurable complexity rules

---

# 6. Registration Flow

## 6.1 Registration Request

Client submits:

- username
- password

Application layer validates:

- Username uniqueness
- Username format rules
- Password policy compliance
- Rate limit compliance

---

## 6.2 Registration Processing

If valid:

1\. Generate account\_id

2\. Hash password

3\. Persist account record

4\. Log registration event (INFO)

5\. Return success

If invalid:

- Return structured error
- Log WARN if suspicious pattern

---

# 7. Authentication Flow

## 7.1 Login Request

Client submits:

- username
- password

---

## 7.2 Authentication Processing

1\. Lookup account by username

2\. Verify password hash

3\. Check account\_status

4\. Generate session\_id

5\. Bind session to account\_id

6\. Log successful login

On failure:

- Return generic authentication error
- Do not reveal which field failed
- Log failed attempt (WARN)

---

# 8. Session Binding

Upon successful authentication:

- Session marked authenticated
- account\_id attached to session context
- role loaded into session
- Command submission enabled

Unauthenticated sessions may:

- Only perform registration
- Only perform login
- Not submit gameplay commands

---

# 9. Rate Limiting

To prevent abuse:

- Limit registration attempts per IP
- Limit failed login attempts per IP
- Lock account temporarily after repeated failures
- Log excessive attempts

Rate limiting enforced at application layer.

---

# 10. Account Status Management

Account statuses:

Active  

Suspended (temporary lock)  

Disabled (permanent lock)  

Suspended accounts:

- Cannot login
- Can be reactivated by admin

Disabled accounts:

- Permanently blocked

All status changes logged.

---

# 11. Admin Role

Admin accounts:

- Assigned manually
- May execute admin-only commands
- Must be logged at elevated verbosity

Admin command execution must:

- Be logged with account\_id
- Include timestamp
- Include parameters used

---

# 12. Persistence Model

Account data stored separately from simulation snapshot.

Recommended structure:

- accounts table or file
- independent of world snapshot
- loaded at server startup
- cached in memory

Account persistence must:

- Not block tick engine
- Use async write
- Ensure atomic updates

---

# 13. Security Requirements

- No plaintext passwords
- No logging of credentials
- No client-side hash acceptance
- No trust of client timestamp
- All authentication server-side only

---

# 14. Logging Requirements

Must log:

- Successful registration (INFO)
- Failed registration (WARN)
- Successful login (INFO)
- Failed login (WARN)
- Account suspension (WARN)
- Admin login (INFO)

Logs must include:

- timestamp
- severity
- account\_id (if known)
- IP address (if available)

Sensitive data must be redacted.

---

# 15. Testing Requirements

Automated tests required for:

- Registration success
- Duplicate username rejection
- Password hashing validation
- Failed login attempts
- Account suspension
- Rate limiting enforcement

Authentication logic must have full unit test coverage.

---

# 16. Non-Goals (v1)

- Email verification
- Password reset workflow
- Multi-factor authentication
- Social login
- Federated identity

---

# 17. Future Extensions

- MFA support
- Email verification
- Password reset tokens
- API tokens
- Role-based fine-grained permissions

Identity expansion must not compromise simulation determinism.

---

# End of Document
