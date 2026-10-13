# Fatewalker Demo Acceptance Checklist

This is the release gate for calling a build a **playable demo**, rather than merely a server that starts. Run it from a clean account and again from a returning account. Record failures with the command, room, expected result, and actual result.

## 1. First connection and tutorial

- [ ] Register an account, create a character, disconnect, and log back in.
- [ ] Confirm the same character and story progress are restored.
- [ ] Read the opening scene as prose; it should establish a place, a person, and a reason to care rather than only list commands.
- [ ] Use `look`, `look <person>`, `examine <object>`, `listen <place>`, and `exits`; each should answer a distinct question.
- [ ] Confirm the tutorial's practice enemies do not attack first.
- [ ] Defeat the Practice Satyr, then use `look` several times. It must remain defeated for the rest of that visit and after reconnecting.
- [ ] Verify the route is understandable: return south to the courtyard, go east to the storehouse, and north through the marble gate when ready.
- [ ] Exercise the storehouse chest: open it, inspect its contents, take the bandage, use it, and verify its healing effect.
- [ ] Confirm tutorial guidance is available on request and does not keep interrupting ordinary exploration.

## 2. Movement and room interaction

- [ ] `exits`, `map`, `worldmap`, and the automatic map agree with the actual room connections.
- [ ] Invalid directions leave the character in place and explain that the route is unavailable.
- [ ] `look` does not reset enemies, duplicate objects, repeat rewards, or advance unrelated quests.
- [ ] `examine`, `read`, `search`, `listen`, and `iscan` distinguish scenery, hidden clues, carried items, and container contents.
- [ ] Items can be picked up, dropped, stored, and retrieved without losing item stats.
- [ ] Story-critical items cannot be accidentally stored away from the quest path.
- [ ] Repeating an interaction gives a sensible response without duplicating rewards or story memories.

## 3. Story and quest continuity

- [ ] The scratched coin can be found, examined, and optionally returned to the village well.
- [ ] The Harpy encounter is not an automatic attack on room entry; inspecting it provides useful context before combat.
- [ ] The Harpy's letter can be found and read, then carried into Asphodel and offered to the Unremembered Shade.
- [ ] The Manticore Den's sealed coffer can be discovered, opened, and its prayer strip read.
- [ ] Charon's memory quest and the Shade's response make sense as a connected story, not a disconnected fetch chain.
- [ ] The Oracle's trust/defiance choice is persistent and mutually exclusive; it unlocks the appropriate branch without silently overwriting the choice.
- [ ] The Oracle's Whisper completes only after asking Pythia about the thread or prophecy; entering a room alone does not substitute for the conversation.
- [ ] The River of Memory completes only after using `listen water` at the ancient Styx, not on arrival.
- [ ] Quest progress survives disconnect/reconnect.
- [ ] `current quests` lists active quests with stable numeric IDs; `current quests track <ID>` selects one; `current quest` shows its checklist and a useful next lead; `current quests untrack` clears it.
- [ ] Every quest checklist has meaningful wording and actionable hints, and completing one quest does not falsely complete another.

## 4. Combat and progression

- [ ] A peaceful creature remains peaceful until provoked; an explicitly aggressive creature can initiate combat.
- [ ] Combat descriptions state what happened, show damage in parentheses, and remain readable at a narrow terminal width.
- [ ] `attack`, `cast`, `guard`, `flee`, and `rest` give clear feedback when usable and when not usable.
- [ ] `ql 1 <target>` uses the first configured quicklist power and reports missing slots, targets, or mana clearly.
- [ ] Defeating an enemy awards XP, loot, quest progress, and reputation at most once.
- [ ] Healing, mana use, level-ups, equipment stats, merchant prices, and faction greetings agree with the character sheet.
- [ ] Rebirth requires the intended threshold, preserves soul-bound memories/relics/choices, resets life-specific progression correctly, and places the character in the appropriate era.

## 5. Persistence and multiplayer

- [ ] A character can disconnect and reconnect without losing location, inventory, equipment, quest flags, memories, choices, or container storage.
- [ ] Creating multiple characters on one account keeps their data separate.
- [ ] `who` has no duplicate stale entries; disconnecting removes a player from the roster.
- [ ] Room emotes and speech reach players in the same room; world chat reaches online players; `tell` reaches only the intended recipient.
- [ ] Two sessions can fight or move independently without corrupting one another's state or terminal output.
- [ ] Opening or closing a container in one session does not reveal or hide its contents for another character.
- [ ] Dropping an ordinary item makes it visible in the shared room, and concurrent pickup attempts give it to exactly one character.
- [ ] An account cannot have two simultaneous sessions; after disconnecting, reconnecting works without stale state overwriting newer progress.

## 6. Terminal presentation and help

- [ ] Long narrative lines wrap at a comfortable width without splitting words or counting ANSI color sequences as visible characters.
- [ ] Room headings, NPCs, threats, objects, and exits are visually distinct but not noisy.
- [ ] Help topics match implemented commands and explain syntax with examples.
- [ ] The game never requires the player to guess a hidden command just to continue the main story.
- [ ] Errors offer a next step; they do not silently do nothing when a useful explanation is possible.

## 7. Automated release checks

Run from the repository root:

```sh
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/fatewalker
```

A demo candidate is not considered release-ready until these checks pass and the manual story path above has been completed at least once from a fresh character. Passing automated tests is necessary, but does not replace the playthrough.
