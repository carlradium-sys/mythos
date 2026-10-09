# Fatewalker Player Tutorial

This is a separate beginner's guide for learning the game before following the story. You can also type `tutorial` or `guide` in-game for a compact version. The terminal UI keeps room descriptions, exits, NPCs, and threats compact. The color legend is not repeated on every screen, and automatic maps are off by default so the important text stays visible.

## 1. The basic loop

- `look` (or `l`) describes your current room, names each exit destination, numbers nearby NPCs, and shows active threats with their HP.
- The room view focuses on **Paths**, **People to talk to**, and **Threats** so exits and targets are easy to spot.
- NPCs and enemies use color-coded labels, but the legend is not repeated on every room view. Player-presence detection is not implemented yet.
- `score`, `inv`, `eq`, and `quests` use separate headings and short labeled lines to make character information easier to scan.
- `exits` lists each available direction and the destination room's name.
- `map` shows a local, active map centered on your current room and its connected exits.
- Automatic mapping is off by default to reduce screen clutter. Use `automap on` to show a local map after movement, `automap off` to hide it, and `automap toggle` to switch modes. `map` always shows the current local map.
- `worldmap` shows the broader world overview.
- Move with `north`, `south`, `east`, `west`, `up`, or `down`; short forms `n`, `s`, `e`, `w`, `u`, and `d` work too.
- `score` (or `stats`) shows health, mana, level, experience, and other character details.
- `help` shows the getting-started guide. `help list` lists every topic; use `help inventory`, `help powers`, `help equipment`, `help map`, or `help automap` for detailed instructions.

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
- `hint` or `hint 2` asks a nearby NPC for a suggested conversation topic or next action.
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
2. **Look:** type `look` to inspect the lane and find Damon, the village guide.
3. **Conversation:** type `talk 1 hello`.
4. **Hint:** type `hint`, then ask about the road with `talk 1 road`.
5. **Combat:** travel north to the Foothills of Olympus, type `look` to inspect the Harpy, and use `attack 1` to fight. You cannot leave while it is alive.
6. **Inventory and gear:** after victory, type `inv`, then `eq` to inspect your equipment.
7. **Quests and story:** type `quests` or `journal`; use `score` to see your character progress. This completes the opening tutorial.
8. Continue north to the Oracle Path. Use `hint` with Pythia if you are unsure what to ask, then follow the story choices and quest guidance.
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
| Shop / buy | `shop`, `buy 1`, `buy <item fragment>` |
| Character stats | `score`, `stats` |
| Quests/story | `quests`, `journal`, `soul` |
| Powers | `powers`, `awaken <domain>`, `cast <power>` |
| Tutorial | `tutorial`, `guide` |
| Disconnect | `quit`, `exit` |
