# Fatewalker Player Tutorial

## The First Thread: one-time guided prologue (about 30 minutes)

Every newly created character begins in an isolated, safe tutorial zone. This is a one-time part of character creation, not a normal world region. The game saves each completed lesson. Logging out and returning resumes the prologue; it does not reset it. After the final gate opens and you leave, the tutorial is permanently closed for that character and has no route back from the main world.

Read the room text and each instruction rather than racing through commands. A first playthrough should take roughly 25–35 minutes if you read descriptions, talk to the guide, and read the combat help.

### The 39 guided lessons

1. **Orient yourself:** type `look`. Read the room description slowly; it establishes where you are and what matters.
2. **Read your character sheet:** type `score`. Learn health, mana, level, experience, life, and era.
3. **Inspect your pack:** type `inv`. Inventory means everything carried, not necessarily everything equipped.
4. **Move to the shelter:** type `north`. Directions follow exits, not coordinates.
5. **Inspect the guide:** type `look Ianthe`. Inspecting a named person reveals more than the room summary.
6. **Greet the guide:** type `talk Ianthe hello`. Dialogue is topic-based and accepts meaningful keywords.
7. **Ask about movement:** type `talk Ianthe movement`. Learn the difference between previewing a route and taking it.
8. **List exits:** type `exits`. This gives precise route names and destinations.
9. **Read the local map:** type `map`. The map is an overview; exits is the precise route list.
10. **Preview a destination:** type `look east`. You can inspect a connected room without entering it.
11. **Enter the courtyard:** type `east`. Read the new room heading and description.
12. **Inspect the training shade:** type `look Training Shade`. Identify a creature before engaging it.
13. **Fight the training shade:** type `attack Training Shade`. Combat text reports damage. It is passive until attacked.
14. **Verify your pack:** type `inv`. Checking inventory or looking around must never resurrect a slain tutorial monster.
15. **Check equipped gear:** type `eq`. Your weapon affects attacks; armor contributes to defense.
16. **Practice recovery:** type `rest`. Rest restores health and mana when you are safe.
17. **Enable automatic mapping:** type `automap on`. The local map appears after successful movement.
18. **Disable automatic mapping:** type `automap off`. Control how much map output you want.
19. **Enter the grove:** type `north`. The next creature is wary but not aggressive.
20. **Inspect the satyr:** type `look Satyr`. Inspection is information gathering, not an attack.
21. **Choose to engage:** type `attack Satyr`. It defends itself only after you strike first.
22. **Confirm persistence:** type `look`. The slain satyr stays defeated for this character.
23. **Return to the courtyard:** type `south`. Retracing a route is normal exploration.
24. **Enter the storehouse:** type `east`. Review your equipment and help systems before leaving.
25. **Review inventory:** type `inv`. Read item names, rarity, and bonuses.
26. **Review equipment:** type `eq`. Know which weapon and armor are active.
27. **Read combat help:** type `help combat`. Help files explain mechanics in more depth than short prompts.
28. **View the wider map:** type `worldmap`. Later destinations may be locked by life, level, or story progression.
29. **Return to the courtyard:** type `west`. Move through exits rather than teleporting.
30. **Return to the shelter:** type `west`. NPCs shown by `look` can be targeted by name or number.
31. **Ask about combat:** type `talk Ianthe combat`. Ask an NPC for advice instead of guessing.
32. **Review quests:** type `quests`. Quests track objectives, progress, rewards, and locked story branches.
33. **Read the journal:** type `journal`. The journal offers direction, not a forced checklist.
34. **Check progress again:** type `score`. Level 3 opens domain awakening; level 10 opens rebirth.
35. **Survival and combat choices:** type `east`. Learn that `attack` uses your weapon, `cast <power>` spends mana, `use <item>` consumes a supply, `guard` reduces the next incoming hit, and `flee` gives the enemy one final strike before escape.
36. **Economy and crafting:** type `east`. Victories and discoveries can earn gold; merchants sell supplies and gear, and some enemies drop materials. At Theron's Forge, `craft` lists recipes; finished gear must be equipped.
37. **Approach the gate:** type `north`. It remains sealed until the review lessons are complete.
38. **Preview the real world:** type `look north`. This reveals Asterion Village without moving you.
39. **Depart:** type `north`. The gate is one-way. Completion is saved, and the tutorial cannot be revisited.

### Tutorial combat rules

Tutorial creatures have persistent defeat flags keyed to their tutorial room. Once slain, the flag prevents that creature from spawning again for that character, including after logout and reconnect. Tutorial creatures do not initiate attacks; they become aggressive only after the player attacks them. These rules are intentionally safer than ordinary hostile encounters in the main world.

### After the prologue

Use `look` to read a room, `look <target>` to inspect a person or creature, `exits` to list routes, `map` for local navigation, `inv` for carried items, `eq` for equipped gear, `quests` for objectives, and `journal` for story direction. Use `help list` to see detailed help topics. The sandbox opens up after the one-way gate, but the tutorial remains closed.

---

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

## 2. Combat in the main world

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
