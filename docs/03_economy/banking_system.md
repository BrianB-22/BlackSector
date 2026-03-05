# Banking System Specification

## Version: 0.1
## Status: Draft
## Owner: Economy
## Last Updated: 2026-03-05

---

# 1. Purpose

Defines the banking system for BlackSector: how players store credits safely, earn interest, transfer funds, and pay other players.

The bank separates a player's **wallet** (credits on their person, vulnerable in combat) from their **bank balance** (credits deposited at ports, protected from robbery and combat outcomes).

---

# 2. Design Principles

- Bank balance is never accessible during combat. Robbers and bribes can only reach wallet credits.
- Each port operates its own bank. Interest rates vary by port — risk and reward are correlated.
- Transfers require a physical dock. Money cannot move over IRN.
- Players should always know their full financial picture with a single command.
- Interest is a passive reward that encourages engagement with the economy and exploration.

---

# 3. Wallet vs. Bank

Players have two places their credits exist:

## 3.1 Wallet

- Stored as `credits` on the `players` table
- Always accessible — no docking required
- Vulnerable: robbery, bribes, and ransom demands draw from wallet only
- Used for all purchases (ships, upgrades, drones, cargo)

## 3.2 Bank Accounts

- One account per port that has a bank (`has_bank: true`)
- Stored in `player_bank_accounts` table, keyed by `(player_id, port_id)`
- Protected: combat outcomes cannot touch bank balances
- Inaccessible in space — requires docking at the holding port to withdraw
- Earns interest at the port's configured rate

**Players can hold accounts at multiple ports simultaneously.**

---

# 4. Interest Rates

Each port with a bank has a configured `interest_rate_percent` (annual equivalent, applied per interest period).

Interest rates correlate with security zone:

| Zone            | Typical Rate     | Notes                                               |
| --------------- | ---------------- | --------------------------------------------------- |
| Federated Space | 1–2% per period  | Guaranteed safe. Almost no return.                  |
| High Security   | 2–4% per period  | Stable. Low risk.                                   |
| Medium Security | 4–7% per period  | Moderate risk. Better return.                       |
| Low Security    | 7–12% per period | High risk port. Elevated return.                    |
| Black Sector    | 12–20% per period| Extreme risk. No law enforcement. Highest yield.    |

Interest rates are defined per-port in world config. They are not dynamic in v1.

## 4.1 Interest Accrual

Interest is calculated and credited once per **interest period** (configured in `server.json` as `bank_interest_period_ticks`).

Formula:
```
interest_earned = floor(balance × interest_rate_percent / 100)
```

Applied to each `player_bank_accounts` row at the end of each period.

Interest is credited directly to the bank account (not the wallet). Players do not need to be docked to receive it.

---

# 5. Operations

All banking operations require the player to be **docked at a port with a bank**.

## 5.1 Deposit

Transfer credits from wallet to the current port's bank account.

```
bank deposit <amount>
bank deposit all
```

- Deducts `amount` from `players.credits`
- Adds `amount` to `player_bank_accounts.balance` for current port
- Instant. No fee.
- Fails if wallet balance < amount

## 5.2 Withdraw

Transfer credits from the current port's bank account to wallet.

```
bank withdraw <amount>
bank withdraw all
```

- Deducts `amount` from `player_bank_accounts.balance` for current port
- Adds `amount` to `players.credits`
- Instant. No fee.
- Fails if account balance < amount

## 5.3 Transfer Between Accounts

Move credits from the current port's account to another port's account (your own). Must be docked at the source port.

```
bank transfer <amount> <port_name>
```

- Deducts from current port's account
- Credits to named port's account (creates account row if it doesn't exist)
- Instant settlement
- Fails if source balance < amount
- Cannot transfer to a port that has no bank

## 5.4 Send to Another Player

Send credits from wallet to another player's wallet. Must be docked. Cannot be done over IRN.

```
bank send <amount> <player_name>
```

- Deducts `amount` from sender's `players.credits`
- Adds `amount` to recipient's `players.credits`
- Recipient is notified via `message_received` on next tick
- Fails if sender wallet balance < amount
- Fails if player_name does not exist
- Cannot be initiated while undocked
- **Cannot be sent over IRN** — requires physical presence at a port

## 5.5 Status Overview

Display complete financial picture: wallet balance, all bank accounts, interest rates, and total net worth.

```
bank
```

Example output:
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  FINANCIAL STATUS — nova

  Wallet:              4,250 Cr    (on hand — vulnerable in combat)

  Bank Accounts:
  ─────────────────────────────────────────────────────────────
  Port                  Zone         Balance     Rate / Period
  ─────────────────────────────────────────────────────────────
  Federated Station α   Federated    12,000 Cr   1.5%
  Vega Trading Hub      High Sec      8,500 Cr   3.2%
  Outer Reach Depot     Low Sec       5,000 Cr   9.0%
  ─────────────────────────────────────────────────────────────
  Total Banked:        25,500 Cr
  Total Net Worth:     29,750 Cr

  Next interest period in: 320 ticks
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

# 6. Combat Interaction

## 6.1 Robbery

When a player is boarded or submits to a robbery demand, only wallet credits are taken. Bank balances are inaccessible during combat and cannot be targeted.

```
Robbery outcome:
  Credits taken from wallet: 4,250 Cr
  Bank balance:              UNTOUCHED
```

## 6.2 Bribe / Ransom

When a player pays a bribe or ransom to escape combat, the amount is deducted from wallet credits. If wallet is insufficient to pay the demanded amount:

- Player may partially pay (if attacker accepts)
- Player may refuse and accept combat consequences

The server does not auto-debit bank accounts to cover bribe/ransom shortfalls. This is intentional: bank accounts require a deliberate dock action.

## 6.3 Implication for Players

The recommended strategy is to keep minimal credits in wallet while traveling in dangerous space, and park the bulk in a bank. Low Security and Black Sector ports offer higher rates but require traveling to withdraw.

---

# 7. Data Model

## 7.1 player_bank_accounts table

```sql
CREATE TABLE player_bank_accounts (
  account_id        TEXT PRIMARY KEY,
  player_id         TEXT NOT NULL REFERENCES players(player_id),
  port_id           INTEGER NOT NULL REFERENCES ports(port_id),
  balance           INTEGER NOT NULL DEFAULT 0,
  opened_at_tick    INTEGER NOT NULL,
  last_interest_tick INTEGER NOT NULL DEFAULT 0,
  UNIQUE (player_id, port_id)
);
```

## 7.2 bank_transactions table

Audit log for all banking activity.

```sql
CREATE TABLE bank_transactions (
  transaction_id    TEXT PRIMARY KEY,
  player_id         TEXT NOT NULL REFERENCES players(player_id),
  port_id           INTEGER REFERENCES ports(port_id),
  transaction_type  TEXT NOT NULL,  -- deposit | withdraw | transfer_out | transfer_in | interest | send | receive
  amount            INTEGER NOT NULL,
  balance_after     INTEGER NOT NULL,
  counterparty_id   TEXT REFERENCES players(player_id),
  counterparty_port_id INTEGER REFERENCES ports(port_id),
  tick              INTEGER NOT NULL
);
```

## 7.3 ports table additions

Two new columns on the `ports` table:

```sql
ALTER TABLE ports ADD COLUMN has_bank INTEGER NOT NULL DEFAULT 0;
ALTER TABLE ports ADD COLUMN interest_rate_percent REAL NOT NULL DEFAULT 0.0;
```

`has_bank` defaults to `0`. World config sets which ports have banks.

---

# 8. Server Configuration

New fields in `server.json`:

```json
"bank": {
  "interest_period_ticks": 1000,
  "max_accounts_per_player": 20
}
```

`interest_period_ticks`: how often interest is applied (default 1000 ticks).
`max_accounts_per_player`: prevents unlimited account sprawl (default 20).

---

# 9. Protocol Messages

## Interest credited (Server → Client)

Delivered to player on next login or active session tick after interest accrual.

```json
{
  "type": "bank_interest_credited",
  "timestamp": 5000,
  "correlation_id": null,
  "payload": {
    "port_id": 12,
    "port_name": "Outer Reach Depot",
    "amount": 450,
    "new_balance": 5450,
    "rate_percent": 9.0
  }
}
```

## Payment received (Server → Client)

```json
{
  "type": "bank_payment_received",
  "timestamp": 5001,
  "correlation_id": null,
  "payload": {
    "from_player": "ghost",
    "amount": 2000,
    "wallet_after": 6250,
    "message": "good trade"
  }
}
```

---

# 10. Player Commands Summary

```
bank                             — show wallet, all accounts, rates, net worth

bank deposit <amount>            — wallet → current port bank
bank deposit all

bank withdraw <amount>           — current port bank → wallet
bank withdraw all

bank transfer <amount> <port>    — current port bank → another port's bank (your accounts)

bank send <amount> <player>      — wallet → another player's wallet (docked only, not over IRN)
```

---

# 11. Relationship to Existing Systems

- **Wallet** (`players.credits`): unchanged column. Now understood as "on-hand credits" subject to combat risk.
- **Robbery / Bribe**: See `docs/04_combat/` — those systems deduct from `players.credits` only.
- **Death (standard mode)**: Insurance payout credited to wallet. Minimum credits floor applies to wallet. Bank accounts are fully retained through death.
- **Death (permadeath mode)**: All accounts wiped on reset. Bank balances are lost.
- **IRN**: Bank transactions are not transmittable over IRN. `bank send` requires docking. No bank commands are available in the IRN command set.
- **Port system**: See `docs/03_economy/port_system.md` — `has_bank` and `interest_rate_percent` are added to port config.

---

# 12. Non-Goals (v1)

- Loans or credit lines
- Shared accounts between players (guilds, partnerships)
- Bank failures or port bankruptcy events
- Dynamic interest rate fluctuation
- ATM-style access at ports without a full bank
- Withdrawal fees or transaction taxes
- Interest compounding (simple interest only in v1)

---

# 13. Future Extensions

- Dynamic rates tied to port economic health
- Loan system with repayment schedule and default consequences
- Black market hawala-style transfers (off-book, no transaction log)
- Guild/corp shared treasury accounts
- Interest compounding
- Rate negotiation for large depositors

---

# End of Document
