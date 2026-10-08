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
  map                  See your current position.
  journal              See the current story thread without a forced quest path.
  rebirth              At level 10, cross the Styx and begin another life.
  save <slot>          Persist your soul to a server-side save slot.
  load <slot>          Restore a saved soul.

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

Inventory shows each item's tier and damage. Fatewalker's long-term gear
system will expand this into weapons, armor, affixes, relics, and items
that carry meaning between lives.`},
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
	"save": {"save", `SAVING & LOADING\n\nYour character can be persisted on the game server.\n\n  save <slot>\n  load <slot>\n\nSlots use a safe name such as hero1 or athena_run. Save data includes your\nlevel, life, rebirth count, era, XP, divine domain, inventory, and other\ncharacter state. Combat encounters are not saved; loading clears active combat.`},
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
	return []string{"start","movement","combat","powers","items","story","journal","rebirth","save","tutorial"}
}
