# Fatewalker: Beyond the Styx

Fatewalker is a Greek-mythology MUD built around reincarnation.

Your first life begins in Ancient Greece. You grow stronger, awaken a divine domain, gain powers, survive monsters, and eventually cross the Styx. Rebirth is not simply a level reset: each life opens a new era, new systems, and new possibilities.

## Current playable foundation

- TCP MUD server on port 4000
- Ancient Greece starter route
- Room-to-room movement
- Active ANSI-color map
- Harpy, Satyr, and Manticore encounters
- Detailed location-aware combat narration
- Critical hits and rare severing events
- XP and level progression
- Tiered loot colors: Common through Mythic
- Four divine domains: Storm, Tide, Ember, Aegis
- Powers unlock at different levels
- Mana and divine progression
- Death recovery through the Styx
- First reincarnation into modern Athens
- Inventory and equipment foundation

## Commands

`look`, `map`, movement commands, `attack`, `cast <power>`, `powers`, `awaken <domain>`, `inventory`, `score`, `flee`, `rebirth`, `help`, `quit`.

## World vision

Fatewalker is intended to become an all-encompassing mythological world. Ancient Greece is only the beginning: Olympus is explorable, Delphi and Ancient Athens have their own stories, and the road below leads to the Styx, Asphodel, judgment, Cerberus, and Tartarus. Life II reframes the same mythology through modern Athens, with hidden temples, museums, metro tunnels, rooftops, and a modern Styx. Later rebirths are not limited to a single third life; they can open new eras and branches, including future versions of Athens, new roads to Olympus, lunar prophecy, and distant mythic futures.

The goal is familiarity without repetition. Places, gods, monsters, symbols, and consequences can return in new forms as history advances. A player should eventually be able to look at the same mountain, city, river, or divine figure across multiple lives and realize that it is both familiar and profoundly changed.

## Design direction

The long-term goal is a living reincarnation game rather than a conventional fantasy MUD. Each life should add a new layer to the player's identity and abilities.

Planned systems include persistent soul traits, divine relationships, faction reputation, more eras, equipment affixes, tactical combat states, quests, NPC relationships, bosses, procedural world events, and meaningful choices about what survives each reincarnation.

The repository is currently named `mythos` for continuity, while the game and Go module are named `fatewalker`.

## Persistent accounts and characters

Fatewalker now uses server-managed persistence rather than player save commands. Players register or log into an account, then select an existing character or create another. Character state—including level, inventory, rebirth history, soul legacy, and quest progress—is written automatically by the server.

The account system is the foundation for multiple characters per player and future account-level systems such as character rosters, shared unlocks, reincarnation records, and persistent world consequences.
