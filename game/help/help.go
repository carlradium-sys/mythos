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
  north/south/east/west Travel through discovered paths.
  attack               Fight with your equipped weapon.
  powers               See divine powers and their unlock levels.
  awaken <domain>      Choose your first divine resonance at level 3.
  score                Review your character.
  inventory             Review your equipment.
  equip <item>          Equip a weapon or armor.
  map                  See your current position.
  journal              See the current story thread without a forced quest path.
  quests               Review active, completed, and choice-locked quests.
  soul                 Review memories, oaths, scars, echoes, and faction standing.
  invoke               Use your once-per-life manifestation.
  rebirth              At level 10, cross the Styx and begin another life.

There is no single correct way to play. The story gives you mysteries and
consequences; the sandbox gives you room to decide what they mean.`},
	"movement": {"movement", `MOVEMENT & EXPLORATION

Use north, south, east, or west. Short forms also work: n, s, e, w.

LOOK is important. Rooms can contain enemies, discoveries, NPCs, quests,
and future story hooks. The map shows where you are in the known world.

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

Some battles are meant to be escaped. Survival is part of the sandbox.

VICTORY
Defeated enemies grant XP. Stronger enemies can also reveal rare equipment
and story clues.`},
	"powers": {"powers", `DIVINE POWERS

At level 3, use:
  awaken storm
  awaken tide
  awaken ember
  awaken aegis

Your choice is permanent for the current character.

Each domain has a progression of powers. The first begins at level 5,
stronger manifestations arrive later, and future systems will let your
choices shape how a domain evolves.

The important idea: your power is not simply borrowed from a god. It is
the shape your soul is becoming.`},
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
cannot create a second copy.`},
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
  You awakened at the Gates of Olympus with no memory of your death.
  A black thread binds itself to your sword.
  Something beyond the gates knows your name.

CURRENT DIRECTION
  Explore the foothills and discover why the creatures seem to recognize you.

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

The tutorial is part of the world, not a separate training room.

You begin at the Gates of Olympus with no explanation for why you are
alive. Follow the prompts, but you are free to leave the intended path.

The tutorial teaches:
  look / movement
  combat
  inventory and gear
  XP and leveling
  divine awakening
  map use
  the first major choice

If you ignore the tutorial and explore, the world does not punish you.
You may simply discover the story in a different order.`},
}

func Get(name string) (Topic, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" { name = "start" }
	t, ok := topics[name]
	return t, ok
}

func Names() []string {
	return []string{"start","movement","combat","powers","items","story","journal","quests","rebirth","talk","soul","invoke","shop","buy","tutorial"}
}
