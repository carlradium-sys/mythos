# Fatewalker Player Tutorial

Fatewalker is a Greek mythic fantasy about a soul that returns from death carrying a black thread tied to its fate. The gods are real, monsters still haunt ancient roads, and rebirth can carry the same soul into modern Athens and later ages. Choices, memories, and consequences persist across lives.

The first journey begins in Asterion Village beneath Mount Olympus. Damon helps explain why you have returned. The opening tutorial is deliberately paced: it presents one action at a time, then gives the next instruction after you complete the current one. Use `tutorial` or `help <topic>` whenever you want a refresher.

## 1. Room views and navigation

- When you first enter a room—at login or after moving—the room is described once. Other commands do not repeat the room view.
- Type `look` whenever you want to see the current room description again.
- Room summaries list NPC names/numbers, merchant labels, and active threat names/health, without full descriptions or a path list.
- Use `look Damon`, `look Harpy`, or `look 1` to inspect a person or creature in detail. Use `look east` (or `look e`) to read a destination room's description without moving.
- `exits` lists every available direction and destination room name.
- `map` shows a compact local diagram; `worldmap` shows the wider world.
- Automatic mapping is off by default. Use `automap on` to show the local map after movement, `automap off` to hide it, and `automap toggle` to switch modes.
- Move with `north`, `south`, `east`, `west`, `up`, or `down`; short forms `n`, `s`, `e`, `w`, `u`, and `d` work too.
- `score`, `inv`, `eq`, and `quests` show their own information without repeating the room.

## 2. The sixteen opening lessons

Each lesson gives one action. Complete it to receive the next instruction.

1. Type `look`.
2. Move north to the village lane.
3. Inspect Damon with `look Damon`.
4. Greet him with `talk Damon hello` (or use the number selector: `talk 1 hello`).
5. Ask about the village with `talk Damon village` (or `talk 1 village`).
6. Request guidance with `hint`.
7. Ask about the road with `talk Damon road` (or `talk 1 road`).
8. Use `exits` to read the detailed route list.
9. Use `map` to view nearby room connections.
10. Move north to the foothills.
11. Inspect the Harpy with `look Harpy`.
12. Fight with `attack Harpy` (or use the number selector: `attack 1`).
13. Use `inv` to review carried items.
14. Use `eq` to review equipped gear.
15. Use `quests` to review objectives.
16. Use `journal` to read the story context.

The tutorial requires inspection before the Harpy fight, and the defeated Harpy stays defeated while you remain in the room. Tutorial progress is saved with the character. Once complete, explore freely and ask NPCs about topics; hints remain on demand.

## 2. Combat without getting stuck

Creatures in dangerous rooms are encounters. Prefer `attack <enemy name>`, such as `attack Harpy`; `attack 1` is also available as a number selector. Use `attack` (also `kill` or `hit`) to strike. Some characters later unlock divine powers and can use `cast <power name>`.

After a creature is defeated, it stays defeated while you remain in that room. You can inspect inventory, check quests, talk, or issue other commands without the creature instantly returning. Leaving the room and entering it again can start a new encounter.

- `flee` abandons an active fight.
- `rest` restores health and mana when you are safe.
- Watch the HP values printed during combat.

## 3. Inventory and equipment

- `inv` (or `inventory`, `i`) lists everything you carry.
- `eq` (or `equipment`, `gear`) shows your currently equipped weapon and armor.
- `wield <item>`, `equip <item>`, and `wear <item>` equip carried gear.

Item names support case-insensitive partial matching. For example, `wield heph` can find **Sword of Hephaestus**. If several items match, the game lists numbered choices. Enter `wield 1` (or `equip 1`) to choose the first result. If a match is unclear, use `inv` to see the full item names.

## 4. Story and conversations

- `talk` greets the first person in the room.
- `hint` or `hint 2` asks a nearby NPC for a suggested conversation topic or next action. Hints are optional and never repeat automatically.
- `talk <topic>` asks that person about a topic, such as `talk fate`.
- Prefer `talk <NPC name> <question>`, such as `talk Damon can you tell me more about the village?` or `talk Pythia what choices do I have?`. The conversation system matches meaningful keywords and topic aliases, so polite wording and full questions are fine; you do not have to type one exact phrase. You can also use a number selector, such as `talk 1 road`; NPC names and numbers are shown by `look`.
- `talk choices` asks Pythia about the choice she offers.
- `shop` lists a merchant's wares with numbers; `buy 1` buys the first listed item, while `buy <part of name>` also works.
- `choose trust` or `choose defy` makes that story choice. Choices may affect reputation, memories, oaths, and later scenes.
- `quests` shows quest progress.
- `journal` (also `quest` or `story`) gives the current story direction.
- `soul` (also `legacy`) shows persistent memories and soul traits.

## 5. Divine powers and progression

- Defeating creatures and completing quests grants XP.
- At level 3, choose one domain with `awaken storm`, `awaken tide`, `awaken ember`, or `awaken aegis`.
- Use `powers` to see divine powers and their unlock levels.
- Use `cast <power name>` during combat when you have an unlocked power and enough mana.
- At level 10, `rebirth` opens the next life. The first rebirth leads to modern Athens. Later rebirths offer more destinations.

## 6. The hand-held opening

The game starts in **Asterion Village**, not at Olympus. Follow the prompts in order:

1. **Movement:** type `north` to enter the Village Lane.
2. **Look:** type `look` to inspect the lane and find Damon, the village guide. Room descriptions show exits, NPCs, and active threats; you do not need to type `look` repeatedly to receive tutorial instructions.
3. **Conversation:** type `talk 1 hello`.
4. **Hint:** type `hint`, then ask about the road with `talk 1 road`. The game teaches this once; afterward hints are only shown when you ask for them.
5. **Combat:** travel north to the Foothills of Olympus, type `look` to inspect the Harpy, and use `attack 1` to fight. You cannot leave while it is alive.
6. **Inventory and gear:** after victory, type `inv`, then `eq` to inspect your equipment.
7. **Quests and story:** type `quests` or `journal`; use `score` to see your character progress. This completes the opening tutorial.
8. The apprentice has a separate forge east of the foothills. Visit him after the Harpy encounter if you want to inspect or buy equipment.
9. Continue north to the Oracle Path. Ask Pythia about choices when you are ready; use `hint` only if you want a nudge.
9. Use `journal` and `quests` as you explore and grow to level 10 before rebirth.

## Command quick reference

| Goal | Commands |
|---|---|
| Look around | `look`, `l` |
| Nearby room map | `map` |
| Automatic map after movement | `automap on`, `automap off`, `automap toggle` |
| Detailed command help | `help inventory`, `help powers`, `help equipment`, `help map` |
| Exit destinations | `exits` |
| World overview | `worldmap` |
| Move | `north` / `n`, `south` / `s`, `east` / `e`, `west` / `w` |
| Fight / target | `attack`, `attack 1`, `attack <enemy name>` |
| Inventory | `inv`, `inventory`, `i` |
| Equipped gear | `eq`, `equipment`, `gear` |
| Equip by partial name | `wield heph`, `equip <name>`, `wear <name>` |
| Talk to numbered NPC | `talk 1 <topic>` |
| Ask an NPC for guidance | `hint`, `hint 2` |
| Visit the village smith | Travel east from the foothills to Theron's Forge |
| Shop / buy | `shop`, `buy 1`, `buy <item fragment>` |
| Character stats | `score`, `stats` |
| Quests/story | `quests`, `journal`, `soul` |
| Powers | `powers`, `awaken <domain>`, `cast <power>` |
| Tutorial | `tutorial`, `guide` |
| Disconnect | `quit`, `exit` |
