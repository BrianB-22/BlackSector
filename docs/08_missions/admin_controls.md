# Mission Admin Controls Specification

## Version: 0.1

## Status: Draft

## Owner: Mission Systems

## Last Updated: 2026-03-05

---

# 1. Purpose

This document defines how a server administrator manages mission content on their Black Sector instance.

Admins control which mission files are loaded, which missions are active, and when changes take effect — without requiring a server restart.

---

# 2. Scope

IN SCOPE:

* mission file directory management
* hot-reload of mission definitions
* enabling and disabling individual missions
* validation feedback
* active mission state inspection

OUT OF SCOPE:

* mission content authoring (see content_schema.md)
* player-facing mission UI
* mission evaluation logic (see mission_framework.md)

---

# 3. Design Principles

* admins control all mission content on their instance
* no mission runs without admin approval
* changes can be made at runtime without server restart
* validation errors are surfaced clearly
* community files can be dropped in and activated immediately

---

# 4. Mission File Directory

All mission files reside in:

```
config/missions/
```

Subdirectories are supported for organization:

```
config/missions/combat/
config/missions/trade/
config/missions/exploration/
config/missions/community/
```

The server scans all `.json` files recursively at startup and on hot-reload.

To add a new mission pack: place the `.json` file in the directory and trigger a reload.
To remove a mission pack: delete or move the file and trigger a reload.

---

# 5. Admin Commands

Admin commands are issued via the server console or admin API.

---

## mission reload

Rescans `config/missions/` and reloads all mission definitions.

```
mission reload
```

Behavior:

* newly added files are loaded and validated
* modified files are reloaded
* removed files are unregistered
* active MissionInstances using a removed definition continue to completion using their snapshot
* validation errors are printed per-file; valid files still load

---

## mission list

Lists all currently loaded missions with their status.

```
mission list
```

Output columns:

* mission_id
* name
* enabled (true/false)
* active instances (count of players currently on this mission)
* source file

---

## mission enable [mission_id]

Enables a loaded but disabled mission, making it available to players.

```
mission enable emergency_ore_run
```

---

## mission disable [mission_id]

Disables an active mission. Players currently on this mission complete it normally; no new instances are offered.

```
mission disable emergency_ore_run
```

---

## mission info [mission_id]

Displays full definition details for a specific mission.

```
mission info emergency_ore_run
```

Output includes: name, description, objectives summary, rewards, security zones, repeatable flag, active instance count.

---

## mission validate [file_path]

Validates a mission file without loading it into the registry. Useful for testing community files before activation.

```
mission validate config/missions/community/new_pack.json
```

Output: validation result per mission definition in the file, with specific error messages for any failures.

---

## mission status [player_id]

Displays the active mission state for a specific player.

```
mission status player_12345
```

Output includes: mission_id, current objective, progress, ticks elapsed, ticks remaining before expiry.

---

## mission reset [player_id] [mission_id]

Resets a player's mission instance. Use to resolve stuck states.

```
mission reset player_12345 emergency_ore_run
```

The instance is removed. The mission returns to AVAILABLE state for that player.

---

# 6. Enable / Disable Within Files

Individual missions within a file can be pre-disabled using the `enabled` field in the JSON definition:

```json
{
  "mission_id": "work_in_progress",
  "enabled": false,
  ...
}
```

This allows an admin to include a file with multiple missions while keeping specific ones inactive.

Enabled state from the JSON is the default. Admin commands override the JSON value at runtime (in memory only — the file on disk is not modified).

---

# 7. Hot-Reload Behavior

Hot-reload does not affect currently active MissionInstances.

On reload:

* new missions become available immediately after validation
* removed missions are unregistered but existing instances run to completion
* updated missions (same mission_id, new version) — existing instances continue on old definition; new acceptances use updated definition
* disabled missions: existing instances complete; no new instances are created

Players are not notified of reload events. Mission availability changes silently from the player perspective.

---

# 8. Community Mission Packs

Community mission files follow the same JSON schema as any other mission file.

Admin workflow for community content:

1. Download or receive a community `.json` file
2. Run `mission validate` to inspect content before activating
3. Place approved file in `config/missions/community/`
4. Run `mission reload`
5. Use `mission list` to confirm the missions loaded correctly
6. Use `mission disable` to turn off any missions within the pack that are not suitable for the instance

Admins are fully responsible for all mission content active on their server.

---

# 9. Logging

All admin mission actions are logged to the server log:

* `mission reload` — lists files loaded, files rejected, counts
* `mission enable / disable` — records which admin issued the command and timestamp
* `mission reset` — records player, mission, and issuing admin
* validation errors — file name, mission_id, error description

---

# 10. Performance Constraints

Hot-reload must complete within 500ms for up to 1,000 mission definitions.

Reload must not block the tick engine. Reload executes between tick cycles.

---

# 11. Security Considerations

Admin commands must require authenticated admin session.

Players cannot:

* trigger mission reload
* enable or disable missions
* inspect other players' mission state

Mission files are read-only to the server at runtime. The server does not write back to mission JSON files.

---

# 12. Non-Goals (v1)

Not included in initial release:

* web-based admin dashboard for mission management
* per-player mission whitelists or blacklists
* automatic community pack discovery or download
* mission scheduling (activate mission X at time Y)

---

# 13. Future Extensions

Possible expansions:

* scheduled mission activation windows (time-limited server events)
* mission pack versioning and dependency resolution
* admin dashboard UI
* permission tiers (moderator vs full admin for mission control)

---

# End of Document
