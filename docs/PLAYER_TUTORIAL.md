# Fatewalker Player Tutorial

Fatewalker is a Greek mythic fantasy about a soul that returns from death carrying a black thread tied to its fate. The gods are real, monsters still haunt ancient roads, and rebirth can carry the same soul into modern Athens and later ages. Choices, memories, and consequences persist across lives.

The first journey begins in Asterion Village beneath Mount Olympus. Damon helps explain why you have returned. The ten-lesson opening tutorial teaches navigation, inspection, conversation, combat, gear, and story tracking in stages. Use `tutorial` or `help <topic>` whenever you want a refresher. NPC hints are on demand, and room information does not repeat automatically after movement.

## 1. Room views and navigation

- Enter a room, then type `look` when you want its description and a concise list of people, merchants, and active threats. Login and movement do not automatically print the room view.
- Room summaries list NPC names and numbers, merchant labels, and threat names/health. They do not print full NPC or enemy descriptions or a path list.
- Use `look Damon`, `look Harpy`, or `look 1` to inspect a person or creature in detail. Use `look east` (or `look e`) to read the destination room's description without moving there.
- `exits` lists every available direction and destination room name.
- `map` shows a compact local diagram of connected rooms. `worldmap` shows the broader world.
- Automatic mapping is off by default. Use `automap on` to show the local map after movement, `automap off` to hide it, and `automap toggle` to switch modes. `map` always displays the local map when requested.
- Move with `north`, `south`, `east`, `west`, `up`, or `down`; short forms `n`, `s`, `e`, `w`, `u`, and `d` work too.
- `score`, `inv`, `eq`, and `quests` each show their own information without reprinting the room.

## 2. The ten opening lessons

1. **Orientation:** type `look` to read the opening room, then move north to the village lane.
2. **Read the room:** type `look` again to learn how room summaries identify people and threats.
3. **Inspect:** type `look Damon` or `look 1` for the guide's full description and interaction options.
4. **Conversation:** use `talk 1 hello`, then `talk 1 village`. The lesson asks you to try both a greeting and a topic.
5. **Hints:** type `hint` to request guidance, then `talk 1 road` to ask Damon about the mountain threat.
6. **Navigation:** use both `exits` and `map`. Damon will not let you continue north until you have tried both tools.
7. **Combat:** travel to the foothills, type `look` for the threat summary, then `look Harpy` to inspect the creature before `attack 1`.
8. **Inventory:** use `inv` to review carried items. Possessing an item is different from equipping it.
9. **Equipment:** use `eq` to review equipped gear. Later, `wield <item>` equips a weapon and `wear <item>` equips armor.
10. **Story tracking:** use both `quests` and `journal` to review objectives and their story context. `score` shows your character's progress.

Lessons advance only after their required actions. Tutorial progress is saved with the character. Once the opening is complete, you can explore freely; the guide remains available through `tutorial` and `help <topic>`.

## 2. Combat without getting stuck

Creatures in dangerous rooms are encounters. Use `attack` (also `kill` or `hit`) to strike; `attack 1` selects the numbered threat shown by `look`, and you can also use its name. Some characters later unlock divine powers and can use `cast <power name>`.

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
- `talk 1 <topic>` selects NPC 1 explicitly; use the NPC number shown by `look` when a room has multiple people.
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
