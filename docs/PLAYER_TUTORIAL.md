# Fatewalker Player Tutorial

This is a separate beginner's guide for learning the game before following the story. You can also type `tutorial` or `guide` in-game for a compact version. The terminal UI uses short section headings, spacing, and color-coded labels to make room details easier to scan.

## 1. The basic loop

- `look` (or `l`) describes your current room, names each exit destination, numbers nearby NPCs, and shows active threats with their HP.
- The room view is grouped into **Paths**, **People to talk to**, **Other travelers**, and **Threats** so exits and targets are easy to spot.
- Color legend: **green** `[NPC]` means an interactable NPC; **blue** is reserved for other players; **yellow** `[ENEMY]` means a hostile creature. Labels remain understandable without color. Player-presence detection is not implemented yet, and the room view says so rather than pretending it can see other players.
- `score`, `inv`, `eq`, and `quests` use separate headings and short labeled lines to make character information easier to scan.
- `exits` lists each available direction and the destination room's name.
- `map` shows a local, active map centered on your current room and its connected exits.
- `worldmap` shows the broader world overview.
- Move with `north`, `south`, `east`, `west`, `up`, or `down`; short forms `n`, `s`, `e`, `w`, `u`, and `d` work too.
- `score` (or `stats`) shows health, mana, level, experience, and other character details.
- `help` shows the available help topics. `help list` lists topics.

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

## 6. Suggested first journey

1. Start at the Gates of Olympus and type `look`.
2. Travel `north` to the Foothills of Olympus.
3. Defeat the Harpy with `attack`.
4. After the victory, try `inv`, `eq`, `score`, and `quests`. The Harpy should not reappear until you leave and re-enter the room.
5. Travel `north` to the Oracle Path.
6. Defeat the Satyr, then speak to Pythia. Try `talk choices`, then choose `trust` or `defy`.
7. Continue exploring, follow `journal` and `quests`, and grow to level 10 before rebirth.

## Command quick reference

| Goal | Commands |
|---|---|
| Look around | `look`, `l` |
| Nearby room map | `map` |
| Exit destinations | `exits` |
| World overview | `worldmap` |
| Move | `north` / `n`, `south` / `s`, `east` / `e`, `west` / `w` |
| Fight / target | `attack`, `attack 1`, `attack <enemy name>` |
| Inventory | `inv`, `inventory`, `i` |
| Equipped gear | `eq`, `equipment`, `gear` |
| Equip by partial name | `wield heph`, `equip <name>`, `wear <name>` |
| Talk to numbered NPC | `talk 1 <topic>` |
| Shop / buy | `shop`, `buy 1`, `buy <item fragment>` |
| Character stats | `score`, `stats` |
| Quests/story | `quests`, `journal`, `soul` |
| Powers | `powers`, `awaken <domain>`, `cast <power>` |
| Tutorial | `tutorial`, `guide` |
| Disconnect | `quit`, `exit` |
