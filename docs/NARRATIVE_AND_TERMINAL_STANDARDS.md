# Fatewalker Narrative and Terminal Standards

This document records the writing and presentation standards used while expanding the first playable demo.

## Narrative principle

Fatewalker is a world first and a command interface second. A tutorial step may introduce a useful command, but the scene must still work as a small piece of fiction: a character wants something, a place has a history, and the player is allowed to notice details that are not immediately useful.

Avoid instructional voice inside the world. Do not describe a room as a place where the player is supposed to learn inventory. Describe a storeroom that has been used by real travelers, with evidence of who passed through it. Explain the inventory command in help; let the room show why a traveler cares about what they carry.

## Voice

- Use natural, varied sentences. Avoid repeating "Type X to learn Y" in every room or NPC response.
- NPCs have their own vocabulary, knowledge, fears, and reasons for speaking. They should not all act as tutorial interfaces.
- Lore may be incomplete or contradictory. Mark rumor as rumor and religious belief as belief; do not flatten every mystery into an authoritative encyclopedia answer.
- Favor concrete sensory details and meaningful objects over strings of generic fantasy adjectives.
- Give the player optional discoveries and conversations. A missed clue should not normally make the main story impossible to finish.
- Keep help files direct, searchable, and complete. In-world prose and technical instructions have different jobs.

## Tutorial pacing

The First Thread is a prologue, not a 39-item examination. Guided mechanics may still be introduced in stages, but the player should experience continuity between them. Keep the persistent progress state for safe reconnects, avoid automatic combat in every room, and let looking, talking, and inspecting remain low-risk actions. Story beats should give the player time to explore and return.

## Terminal output

- Use CRLF line endings for Telnet output.
- Wrap prose to a predictable width without counting ANSI escape sequences as visible characters.
- Keep headings, room prose, people, threats, exits, and command prompts visually distinct.
- Reset ANSI styling at message boundaries; the game must remain readable without color.
- Use blank lines to separate semantic blocks, not between every line.
- Avoid repeating the entire room after every command. Movement shows the new room; look deliberately repeats the full view.
- Keep lists concise, and provide help and journal to expand detail on demand.

## Reference inspiration

Aardwolf is a reference for comprehensive help and readable room presentation. Erion is a reference for mission continuity and reminders. Fatewalker should combine those strengths with original Greek-inspired lore, persistent choices, and a world that changes across rebirths; it should not copy their text or distinctive quest content.
