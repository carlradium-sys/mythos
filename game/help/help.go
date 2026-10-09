package help

import "strings"

type Topic struct {
	Name string
	Text string
}

var topics = map[string]Topic{
	"start": {"start", `FATEWALKER: GETTING STARTED

You are not following a linear quest. You are entering a living myth.

Your first life begins in Ancient Greece. Explore, fight, learn, awaken,
make choices, and uncover why the Fates have marked you for rebirth.

BASIC LOOP
  look                 Examine the room and discover threats.
  hint                 Ask a nearby NPC what to ask or do next.
  north/south/east/west Travel through discovered paths.
  attack               Fight with your equipped weapon.
  powers               See divine powers and their unlock levels.
  awaken <domain>      Choose your first divine resonance at level 3.
  score                Review your character.
  inventory             Review carried items and their bonuses.
  wield / wear <item>  Equip a weapon or armor.
  map                  Show the local map now.
  automap on/off       Show or hide the local map after movement.
  help <topic>         Read a detailed help file (try inventory or powers).
  journal              See the current story thread without a forced quest path.
  quests               Review active, completed, and choice-locked quests.
  soul                 Review memories, oaths, scars, echoes, and faction standing.
  invoke               Use your once-per-life manifestation.
  rebirth              At level 10, cross the Styx and begin another life.

First-time visits to places grant exploration XP, so discovery matters as
much as combat. You can revisit places, but a room's discovery reward is one-time.

There is no single correct way to play. The story gives you mysteries and
consequences; the sandbox gives you room to decide what they mean.`},
	"movement": {"movement", `MOVEMENT & EXPLORATION

Use north, south, east, west, up, down, in, or out. Short forms also
work for cardinal and vertical directions: n, s, e, w, u, d.

LOOK is important. Rooms can contain enemies, discoveries, NPCs, quests,
and future story hooks. Automatic mapping is off by default to keep the room view compact. Use automap on to show a map after movement, automap off to hide it, or automap toggle to switch modes. Type map to show it at any time; worldmap shows the wider world.

Combat can lock a path until the threat is defeated or escaped.

EXPLORATION TIP
If a room feels important, read its description. Fatewalker is designed so
that the world itself tells part of the story.`},
	"combat": {"combat", `COMBAT

ATTACK
  attack

Your weapon determines the style of the attack. Combat descriptions can
target different body locations and may trigger special effects.

Every damaging action ends with its damage in parentheses:
  Your bronze sword slashes the harpy's wing. (14 damage)

CAST
  cast Stormspark
  cast Thunderfall

Divine powers cost mana and unlock as you grow.

FLEE
  flee

REST
  rest

Rest restores health and mana when no enemy is actively fighting you.
Some battles are meant to be escaped. Survival is part of the sandbox.

VICTORY
Defeated enemies grant XP and gold. Use gold in shops for gear.`},
	"powers": {"powers", `DIVINE POWERS

At level 3, use:
  awaken storm
  awaken tide
  awaken ember
  awaken aegis

Your choice is permanent for the current character.

Each domain has a progression of powers. Use powers to list the abilities
available to your chosen domain, including each power's unlock level and
mana cost. You must meet the unlock level and have enough mana to cast.

During an active fight, use cast <power name>, for example:
  cast Stormspark
  cast Thunderfall

Power names are matched without regard to capitalization. If you have not
chosen a domain, use awaken <domain> at level 3. Your domain choice is
permanent for this character, so review the list before deciding.

The important idea: your power is not simply borrowed from a god. It is
the shape your soul is becoming.`},
	"inventory": {"inventory", `INVENTORY

Use:
  inv
  inventory
  i

Your inventory lists carried items, their rarity, and their damage and armor
bonuses. [EQUIPPED] marks gear currently in use; [SOUL RELIC] marks a relic
that persists through rebirth.

EQUIPPING ITEMS
  wield <item name>   Equip a weapon.
  equip <item name>   General equipment command.
  wear <item name>    Equip armor.

Item names support case-insensitive partial matching. For example:
  wield heph

If more than one item matches, the game displays numbered choices. Select
one with the corresponding numbered equip command. Use eq to see what is
currently equipped.`},
	"equipment": {"equipment", `EQUIPMENT

Use:
  eq
  equipment
  gear

This view shows your current weapon and armor.

CHANGE YOUR GEAR
  wield <item name>   Equip a weapon.
  wear <item name>    Equip armor.
  inv                 Review items you carry.

Partial item names work (for example, wield heph). If several items match,
choose from the numbered results. Equipment bonuses contribute to your
combat attack and defense values.`},
	"map": {"map", `MAPS & NAVIGATION

  map                 Show the local map centered on your current room.
  automap             Check whether automatic mapping is on or off.
  automap on          Show the local map after every successful movement.
  automap off         Stop showing the map automatically.
  automap toggle      Switch automatic mapping on or off.
  exits               List exits and their destination room names.
  worldmap            Show the broader world overview.

Automatic mapping is OFF by default to keep the room view compact. When
enabled, a local map appears after each successful movement. Type map at
any time to show the current room and its connected destinations.`},
	"automap": {"automap", `AUTOMATIC MAP

Automatic mapping displays the local map after each successful movement.
It is OFF by default to keep the terminal view compact.

Commands:
  automap on
  automap off
  automap toggle
  automap              Show the current on/off setting.
  map                  Display the local map immediately.

This setting applies to your current session. A new connection starts with
automatic mapping disabled.`},
	"items": {"items", `ITEMS & RARITY

Item tiers:
  Common      White
  Uncommon    Green
  Rare        Blue
  Epic        Purple
  Legendary   Gold
  Mythic      Bright red/divine

Inventory shows each item's tier, damage, and armor. Use equip <item> to change your weapon or armor.
Soul relics are marked in your inventory and persist through rebirth. The
Styxglass Shard is the first relic: recover it by completing The River of
Memory at the Styx. Once per encounter, its river-ward absorbs up to 8 damage.
The ward refreshes when a new enemy encounter begins. Repeating a reward
cannot create a second copy.
An Oracle's Thread from the trust branch absorbs up to 4 damage once per
encounter. The defiance branch grants an Unwritten Ember, adding 6 damage to
your first successful weapon strike each encounter. Finding the Delphi
Sanctum also completes the Trials of the Seer and grants the Laurel of the
Seer, which adds 3 damage to your first successful weapon strike each
encounter. Defensive relics do not stack; the strongest available ward is used.`},
	"story": {"story", `STORY & SANDBOX

Fatewalker is built around two promises:

1. You can wander.
2. Your wandering matters.

The main story provides major mysteries, characters, bosses, and turning
points. Optional exploration can uncover alternate routes, rare enemies,
lost relics, hidden history, factions, and consequences that can change
later lives.

You should never feel that you are merely walking from quest marker to
quest marker. The world should give you reasons to explore without taking
away your freedom.`},
	"rebirth": {"rebirth", `REBIRTH & THE STYX

At level 10, rebirth becomes possible.

Rebirth is not a prestige button and not a simple reset. A life ends, but
parts of the soul survive: memories, scars, oaths, relationships, favors,
curses, relics, bloodlines, and divine resonance.

Life II opens the modern Greek world. Later lives will introduce new eras
and new mechanics.

Each life also manifests a once-per-life gift. Use 'invoke' to call on it:
  Thread Sense       — sense paths and available story threads.
  Echo Sight         — recall a soul-memory and recover mana.
  Chronal Pulse      — recover health by bending time.
  Moon's Shelter     — recover health and mana beneath lunar protection.
  Fateweaver's Knot  — sever a curse or receive a lasting favor.

The gift changes with the destination of rebirth and resets for each new life.
The long-term goal is to make every life feel like a chapter of one soul's
story rather than another unrelated character.

WORLD STRUCTURE
Ancient Greece grows from Olympus and Delphi into Athens and the realms below.
Olympus itself is a destination, not merely background scenery. The Underworld
contains the Styx, Asphodel, judgment, Cerberus, and the road toward Tartarus.

Life II changes the same mythology into modern Athens: museums, metro tunnels,
rooftops, hidden temples, and a modern Styx. Later lives will open additional
eras rather than forcing every player through the same third life.`},
	"journal": {"journal", `THE FATEWALKER JOURNAL

Your journal tracks the story without turning the world into a checklist.

The First Thread:
  You began in Asterion, a small village beneath Mount Olympus.
  A black thread binds itself to your sword.
  Something beyond the village knows your name.

CURRENT DIRECTION
  Complete the village lessons, then explore the foothills and discover
  why the creatures seem to recognize you.

Remember: a journal objective is a thread, not a command. You can ignore it,
explore elsewhere, hunt for equipment, or pursue another mystery. The story
will wait—and sometimes the world will change while you are away.

`},
	"quests": {"quests", `QUESTS

Story quests are persistent parts of your character. Progress is saved automatically.

  quests    Show active and completed story quests.
  journal   Show the broader story direction.

Some quests only appear after a pivotal choice. Trusting Pythia and defying
her prophecy open different paths, and those consequences remain after rebirth.

QUEST CHAINS ACROSS LIVES
  The Oracle's Whisper must be completed before The River of Memory can
  resolve. Recovering the Styx memory records a permanent soul flag and echo.
  After rebirth, that memory unlocks Echoes in Glass at the modern Styx.
  Myrto may recognize what you did in an earlier life.

Quests reward exploration and combat without forcing a linear path.`},
	"talk": {"talk", `NPC DIALOGUE

Use:
  talk
  talk thread
  talk fate
  talk choices      Show the Oracle’s first major choice.
  choose trust      Trust Pythia; gain Delphi reputation and a lasting oath.
  choose defy       Reject prophecy; lose Delphi reputation and gain a scar.

Story choices are recorded on your character and survive logout and rebirth.
Some choices can only be made once; the world remembers what you decide.

Use soul (or legacy) to inspect the marks your decisions leave behind.`},
	"soul": {"soul", `SOUL LEGACY

Use:
  soul
  legacy

Review persistent memories, oaths, scars, favors, curses, echoes, faction
standing, and recorded story choices. These marks can survive rebirth.
Soul relics appear in inventory with a [soul relic] marker and endure across
lives alongside your memories and other legacy marks.

Faction standing changes through story choices and selected quests. Friendly
factions may offer better prices; hostile factions may charge more.`},
	"shop": {"shop", `MERCHANTS

Some NPCs sell equipment.

  shop             Show the current merchant's wares.
  buy <item>       Purchase an item.

Gold persists with your character. Equipment purchased in one life remains
part of that character's inventory unless a future soul system changes it.

Faction standing affects prices: standing of +3 or higher grants a 10%
discount, while standing of -2 or lower adds a 10% surcharge.`},
	"buy": {"buy", `BUYING

Stand near a merchant and use:
  shop
  buy <item>

Purchased equipment is added to your inventory and can be equipped normally.`},
	"tutorial": {"tutorial", `THE FIRST THREAD — STORY TUTORIAL

You died beside the black river and woke again near Asterion, with a dark
thread tied to your soul. The gods are real, monsters haunt old roads, and
rebirth can carry you into another age. Your choices remain with you.

The opening tutorial has ten deliberate lessons:
  1. Type look to read your current room, then move north to the village lane.
  2. Use look again to read the new room's summary.
  3. Inspect Damon with look Damon or look 1 for his full description.
  4. Talk 1 hello, then talk 1 village to practice a greeting and topic.
  5. Use hint, then talk 1 road to learn about the mountain threat.
  6. Use exits for the detailed route list and map for nearby connections.
  7. At the foothills, look Harpy before using attack 1.
  8. Use inv to inspect carried items.
  9. Use eq to inspect equipped gear.
 10. Use both quests and journal to review objectives and story context.

Room summaries do not print paths or full NPC descriptions. Use look <target>
for details, exits for the full route list, and map for a local diagram.
Room information is not printed automatically after movement; use look when
you want to inspect your surroundings. Hints appear only when requested.
Use tutorial or help <topic> whenever you need a refresher. Progress is saved.`},
	"hint": {"hint", `NPC HINTS

Use:
  hint       Ask the first nearby NPC for a hint.
  hint 2     Ask NPC number 2 (use look to see NPC numbers).

Hints are optional and only appear when you request them; room descriptions
and shop views do not repeat tutorial instructions.

Hints suggest a conversation topic or a practical next step. Examples:
  Damon: talk 1 road
  Pythia: talk 1 choices, then choose trust or choose defy
  Theron: talk 1 thread, then deal with the Harpy in the foothills
  Myrto: talk 1 ancient, or talk 1 museum when the memory is ready

Hints are guidance, not quest completion. Use quests and journal for
your current story progress.`},
}

func Get(name string) (Topic, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" { name = "start" }
	t, ok := topics[name]
	return t, ok
}

func Names() []string {
	return []string{"start","movement","combat","powers","inventory","equipment","items","map","automap","story","journal","quests","rebirth","talk","hint","soul","invoke","shop","buy","tutorial"}
}
