# The First Thread: Fatewalker Prologue

The First Thread is the opening chapter of Fatewalker, not a test of whether a player can memorize commands. The player wakes beside the black water with an unfamiliar thread around their wrist, meets Ianthe at the Wayfarer's Shelter, and gradually learns that the village and mountain have been unsettled by their return from death.

The prologue is saved to the character. Logging out should preserve progress; reconnecting should resume the story rather than restart encounters or repeat completed beats. The tutorial area is a safe place to explore, and its practice creatures wait to be attacked. The gate eventually opens onto Asterion Village and cannot be used to return to the prologue.

## Story beats

### 1. The Shore Between Heartbeats

The player wakes without a clear memory of death. Black water, the dark thread, and a voice from the shelter establish the mystery. Encourage the player to look around, read their character sheet, and inspect what they carry, but do not treat these actions as disconnected tests.

### 2. The Wayfarer's Shelter

Ianthe has prepared ordinary travel supplies because she expects the player to leave. Her concern is personal, and she knows more about the thread than she initially admits. Conversation should support natural questions and follow-ups, not require one exact phrase. The player can inspect the map, preview a path, and learn how exits connect.

### 3. The Courtyard

The training shade is made of dusk and ash. It does not attack first. The player can inspect it, check their equipment, and decide when to start the first fight. Combat teaches targeting, damage feedback, and the distinction between passive and aggressive creatures. Looking around after victory must not respawn the defeated shade.

### 4. The Quiet Practice Grove

A young satyr has been asked to keep the player company. He is wary and a little rude, but not initially hostile. This encounter exists to demonstrate intent: creatures are not automatically enemies merely because they occupy a room. If the player attacks, the satyr defends himself; if the player chooses not to, the scene should remain peaceful.

### 5. The Storehouse and the Gate

The storehouse contains useful objects and signs of previous travelers. A practice chest holds a real healing bandage, so players can scan the room, open a container, look inside, pick up an item, and examine its description before reviewing inventory and equipment. The gate offers a view of Asterion Village and the first open-world choices. Crossing it closes the prologue for that character.

## Interaction commands

- `look`: read the current room, including visible people, threats, and loose objects.
- `examine <name>` or `look <name>`: inspect a person, creature, object, or connected room.
- `talk <name> <question>`: ask a nearby character about a subject in natural language.
- `hint`: request an optional nudge. Hints should not repeat automatically.
- `exits`, `map`, and `worldmap`: understand the nearby routes and wider world.
- `scan`: list visible people and creatures. Scanning does not start combat.
- `iscan`: list visible objects and distinguish containers, takeable items, and scenery.
- `open <container>`, `close <container>`, and `look in <container>`: reveal and inspect container contents. Closed containers keep their contents hidden.
- `get <name>` or `get <item> from <container>`: pick up an available object; `get all` collects visible takeable objects.
- `examine <name>`, `search <name>`, and `read <name>`: investigate details, search an open container, or read a written clue.
- `offer <item> to <object>`: make a story offering when the situation calls for one; `offer memory <moment>` records a player-authored memory in Asphodel.
- `inv` reviews carried items. `drop <name>` leaves ordinary non-quest items in the room. `put <item> in <container>` stores an ordinary item in an open container; `get <item> from <container>` retrieves it.
- `eq`, `wield <name>`, and `wear <name>`: review and change equipment.
- `attack <name>`: begin combat when the player chooses to fight.
- `help <topic>`, `quests`, and `journal`: look up rules and recover the larger story context.

## Narrative and implementation notes

The prologue may introduce mechanics in stages, but its presentation should not label every action as a numbered lesson or imply that the only valid way to play is to follow a checklist. Each room needs its own atmosphere and reason to exist. NPCs should have opinions, incomplete knowledge, and motives beyond teaching commands. Optional objects and conversations should reward curiosity without blocking the main story.

Tutorial defeat flags and collected-object flags belong to the individual character, not the shared room or current connection. Passive tutorial creatures should remain passive until attacked. Main-world creatures may be aggressive, wary, curious, indifferent, or willing to bargain according to the encounter design.

The game should wrap prose to a predictable terminal width without counting ANSI escape sequences as visible characters. Room headings, descriptions, objects, people, threats, and exits should form distinct blocks. The command prompt belongs on the input line, not on a line by itself. Color should improve readability but never be required to understand the output.

This document describes the intended narrative standard. The tutorial is still being migrated from its original command-gated progression into a more flexible story flow; the visible prose should continue improving without breaking saved character progress.
