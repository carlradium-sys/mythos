package help

import (
	"fmt"
	"strings"

	"fatewalker/game/power"
)

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
  look                 Read the room, its objects, and visible occupants.
  look <thing>         Inspect a person, creature, object, or exit.
  get <item> [from <container>]  Pick up a takeable object.
  get all              Pick up all visible takeable objects in the room.
  drop <item>          Leave a non-equipped, non-quest item in the current room.
  iscan                List objects you can inspect or handle.
  scan                 List people, creatures, and visible objects.
  who                  List connected players.
  hint                 Ask a nearby NPC what to ask or do next.
  north/south/east/west Travel through discovered paths.
  attack               Fight with your equipped weapon.
  powers               See divine powers and their unlock levels.
  awaken <domain>      Choose your first divine resonance at level 3.
  score                Review your character.
  inventory             Review carried items and their bonuses.
  wield / wear <item>  Equip a weapon or armor.
  use <item>           Consume a healing or mana-restoration item.
  craft <recipe>       Make gear at Theron's forge from monster materials.
  map                  Show the local map now.
  automap on/off       Show or hide the local map after movement.
  help <topic>         Read a detailed help file (try inventory or powers).
  journal              See the current story thread without a forced quest path.
  quests               Review active, completed, and choice-locked quests.
  current quests       List active quests with IDs and progress.
  current quest        Show the tracked quest, checklist, and next lead.
  current quests track <ID>  Track one quest by its displayed ID.
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
  attack Harpy       Attack by enemy name.
  attack 1           Number-selector alternative.

Your weapon determines the style of the attack. Combat descriptions can
target different body locations and may trigger special effects.

Aggressive creatures can strike as you enter their room. Passive creatures,
including the tutorial's practice enemies, wait until you attack them first.
A successful attack, divine cast, or consumable use gives an aggressive
enemy a chance to retaliate. A consumable is not a free combat turn.

AUTOMATIC OR MANUAL COMBAT
  pc                  Toggle automatic combat on/off instantly.
  pc on               Enable automatic follow-up weapon attacks.
  pc off              Disable automatic attacks; type attack each time.
  combatmode          Display the current setting.
  combatmode on/off   Set the mode explicitly.

COMBAT QUICKLIST
  ql                  List your four quick slots.
  ql set 1 Stormspark Assign a currently unlocked power.
  ql 1 Harpy          Cast slot 1 at the current target.
  ql clear 1          Clear a slot.

Quicklist assignments are saved per character. Slots only contain powers
unlocked in your current divine domain; casts still cost mana and use normal
combat retaliation. Basic attacks continue automatically in automatic mode.

Automatic combat is on by default. After your first attack, your character
continues making basic weapon attacks every 2.5 seconds while you are in
combat. You can still cast powers, use items, guard, or flee. The setting
is saved separately for each character.

Every damaging action ends with its damage in parentheses:
  Your bronze sword slashes the harpy's wing. (14 damage)

CAST
  cast Stormspark
  cast Thunderfall

Divine powers cost mana and unlock as you grow.

FLEE
  flee

Fleeing gives the enemy one final chance to strike before you escape. It
ends the encounter without victory XP, gold, or loot. If the strike defeats
you, normal recovery occurs instead. Controlled tutorial practice fights
cannot be fled; complete the lesson encounter to keep the prologue moving.

REST
  rest

Rest restores health and mana when no enemy is actively fighting you.
Some battles are meant to be escaped. Survival is part of the sandbox.

VICTORY
Defeated enemies grant XP and gold. Use gold in shops for gear.`},

	"combatmode": {"combatmode", `COMBAT MODE

  pc                 Toggle automatic combat on or off.
  pc on              Enable automatic follow-up weapon attacks.
  pc off             Require a typed attack for every weapon swing.
  combatmode         Show the current setting.
  combatmode on/off  Set the mode explicitly.

Automatic mode is the default. After you start a fight, weapon attacks
continue every 2.5 seconds until the enemy is defeated, you flee, or you
switch to manual mode. Divine powers and defensive actions remain available.
The preference is saved separately for each character.`},

	"quicklist": {"quicklist", `COMBAT QUICKLIST

  ql                       Display the four quick slots.
  ql set 1 Stormspark      Assign an unlocked divine power to slot 1.
  ql 1 Harpy               Cast the slot 1 power at the current enemy.
  ql clear 1               Empty slot 1.

Use slots 1 through 4. Assignments are saved with your character.
Only powers unlocked in your current domain can be assigned. Casting still
costs the power's normal mana and gives an aggressive enemy its normal turn.
With automatic combat enabled, your weapon attacks continue between powers.`},
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
bonuses. Consumables also show their healing or mana restoration values.
[EQUIPPED] marks gear currently in use; [SOUL RELIC] marks a relic that
persists through rebirth.

Use 'use <item name>' to consume a healing or mana-restoration item. An item
is not wasted if it would have no effect because you are already at full
health or mana. During aggressive combat, the enemy acts after the item use.

EQUIPPING ITEMS
  wield <item name>   Equip a weapon.
  equip <item name>   General equipment command.
  wear <item name>    Equip armor.

Item names support case-insensitive partial matching. For example:
  wield heph

If more than one item matches, the game displays numbered choices. Select
one with the corresponding numbered equip command. Use eq to see what is
currently equipped.

Use 'put <item> in <container>' to store an ordinary item in an open container.
Quest-critical items and soul relics are protected from being put away.`},
	"put": {"put", `STORING ITEMS IN CONTAINERS

Use:
  iscan
  open <container>
  put <item> in <container>
  look in <container>
  close <container>

Only open containers accept items. Closing a container hides its contents;
open it again to retrieve them with 'get <item> from <container>'. Consumables
retain their healing or mana properties when stored and retrieved. Soul relics
and story-critical quest items cannot be stored away, preventing accidental
quest softlocks.`},
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

WORLD OBJECTS
  look <thing>              Read a room object's description.
  examine <thing>           Inspect an object more closely.
  look in <container>       See what is visible inside a container.
  open <container>          Open a container and reveal hidden contents.
  close <container>         Close an open container.
  read <item>               Read a carried or visible written object.
  search <thing>            Search a visible object or open container for contents.
  get <item> from <source>  Collect a takeable item.
  get <item> <source>       Short form for the same action.
  get all                   Collect all visible takeable objects.
  drop <item>               Leave a non-equipped, non-quest item in the room.
  offer <item> to <object>  Make a story offering when a quest calls for one.
  offer memory <moment>     Offer a specific remembered experience in Asphodel.
  iscan                     List room objects and whether they can be taken.
  scan                      List visible people, creatures, and objects.

Objects are physical room fixtures, not just text in a description. Some are
scenery and cannot be collected; containers may hold hidden items until opened. Readable objects can reveal new quest steps when examined in context.
Your collected objects appear in inventory and can be examined there.
Soul relics are marked in your inventory and persist through rebirth. The
Styxglass Shard is the first relic: recover it by completing The River of
Memory at the Styx. Once per encounter, its river-ward absorbs up to 8 damage.
The ward refreshes when a new enemy encounter begins. Repeating a reward
cannot create a second copy.
An Oracle's Thread from the trust branch absorbs up to 4 damage once per
encounter. The defiance branch grants an Unwritten Ember, adding 6 damage to
your first successful weapon strike each encounter. Examining the broken inscription in the Delphi
Sanctum completes the Trials of the Seer and grants the Laurel of the Seer,
which adds 3 damage to your first successful weapon strike each encounter. Defensive relics do not stack; the strongest available ward is used.`},
	"communication": {"communication", "PLAYER COMMUNICATION\n\nWHO\n  who                 List connected players, their life, level, era, and room.\n\nROOM MESSAGES\n  say <message>       Speak to players in your current room.\n  emote <action>      Describe an action to players in the same room.\n\nPLAYER MESSAGES\n  tell <player> <text> Send a private message to an online player.\n  chat <message>      Send a world-wide message to connected players.\n  gossip <message>    Alias for chat.\n\nNPC dialogue remains separate: use talk <NPC> <topic>, or say <NPC> <topic>\nwhen you want to address someone in the room. Use who to see who is online."},
	"listen": {"listen", `LISTENING & ENVIRONMENTAL CLUES

Use listen <sound or feature> to pay attention to a place instead of simply passing through it.

Examples:
  listen well       Hear what the old well in Asterion conceals.
  listen to water   Follow a sound in the room when the water or river is nearby.
  listen water      At the ancient Styx, hear the river's question.
  listen water      In modern Athens, recognize the oars if you first heard them at the village well.

Some clues connect places across eras. The River of Memory advances when you listen at the ancient shore, not merely when you arrive. If a sound seems incomplete, remember where you first heard it and return when the world has changed. Listening does not start combat; it reveals details only when you are in the right place and pay attention to the right thing.`},
	"scan": {"scan", `SCANNING & ROOM OBJECTS

Use scan to list visible people, creatures, and objects that stand out.
Use iscan to focus on objects you can inspect, open, or pick up.

OBJECT INTERACTION
  look coin                  Read an object description.
  examine coin               Same as look coin.
  inspect campfire ash       Inspect a container or fixture.
  get coin                   Pick up a visible takeable item.
  get coin from ash          Specify the container holding an item.
  get coin ash               Short form accepted for the same action.
  get all                    Pick up all visible takeable objects.
  drop laurel poultice       Leave a carried item in the room.

Items listed in a room are not necessarily takeable. Lamps, altars, doors,
and other fixtures may be scenery; containers can hold discoverable items.
A carried object appears in inventory and can be examined there too.
Room objects are world objects, not merely text in the room description.`},
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
	"current": {"current", `CURRENT QUEST TRACKER

Use the tracker when you want a focused view of one story thread.

  current quests              List active quests and their numeric IDs.
  current quests track <ID>  Pin an active quest to your tracker.
  current quest               Show the tracked quest's story, checklist, and next lead.
  current quests untrack      Clear the pinned quest.

The checklist records discoveries already made, such as finding an object,
reading a letter, hearing a witness, or reaching a story location. The next
lead suggests a useful action without revealing every twist in the story.
Quest selection and progress are saved with your character. If you have not
chosen a quest yet, 'current quest' selects the first available active quest.

`},
	"talk": {"talk", `NPC DIALOGUE AND LORE

Talk to people as people, not as vending machines for quest keywords. The keyword is only how the MUD interface finds a subject; the response should reveal what this character has seen, believes, fears, or refuses to explain.

Prefer targeting people by name:
  talk Damon hello
  talk Damon what happened at the well
  talk Damon why does the Harpy steal letters
  talk Pythia what do you know about the thread
  talk Charon what does the river take
  talk Unremembered why does judgment frighten you
  talk Eleni what did the mirror show
  talk choices

You can also use a number selector, such as talk 1 hello or talk 1 choices. Natural-language questions work when they contain a recognized subject or topic; try different phrasing if an NPC does not understand. Use look to see who is present.

Listen for concrete details: objects, customs, dates, contradictions, repeated symbols, and things a character admits they cannot know. A witness can be honest and still be mistaken. Compare accounts rather than assuming the first explanation is the whole truth.

Use:
  talk thread
  talk fate
  talk memory
  talk choices      Hear the Oracle's first major choice.
  choose trust      Trust Pythia's warning; gain Delphi reputation and a lasting oath.
  choose defy       Reject her interpretation; gain a lasting scar.

Hear Pythia's warning about the thread or prophecy before choosing. She will
not accept trust or defiance until you understand what you are answering.
  offer memory <text>  Share a detailed personal memory with the Asphodel shade.

For the optional Underworld story, hear Charon's lesson, listen to the shade's
fragments, then offer a specific memory in your own words. A short phrase is
not enough; describe a moment that could matter even if its owner's name is lost.

Story choices are recorded on your character and survive logout and rebirth. Consequences may return as changed dialogue, soul memories, faction standing, quests, or evidence in a later era.

Use soul (or legacy) to inspect the marks your decisions leave behind.`},
	"offer": {"offer", `OFFERINGS & WITNESSES

Offer a physical item to a meaningful recipient when the story calls for it:
  offer coin to bowl       Return the examined scratched coin to Asterion's well.
  offer letter to shade    Give the read feather-folded letter to the Unremembered Shade.
  give letter to shade     Alias for the same item handoff.

The optional Underworld memory quest also asks you to share a real, specific memory with
the Unremembered Shade in Asphodel.

First, hear Charon's lesson at the Shores of the Styx:
  talk Charon memory

Then listen to the shade's fragments:
  talk Unremembered memory
  talk Unremembered token

Finally, offer a detailed memory in your own words:
  offer memory I remember someone singing off-key to comfort me when I feared the dark.

A short phrase is not enough. Describe a moment with enough detail that it
feels like a lived experience. The offered text is recorded in your soul
memories and becomes part of this character's story across later lives.`},
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
	"guard": {"guard", `GUARDING

Use 'guard' or 'defend' while an aggressive enemy is alive. You brace for
the next strike. If it hits, a successful guard reduces that hit's damage
by about half. If the enemy misses, the guard still protects you for that
exchange. Guarding spends your action and does not deal damage or restore
health. It is consumed by the next enemy turn and does not stack.

Use guard when your health is low, when you need to survive one more blow,
or before making a tactical choice. A guard cannot provoke a passive
creature; tutorial creatures remain passive until you attack them first.`},
	"progression": {"progression", `PROGRESSION AND DIVINE AWAKENING

Defeating enemies, discovering rooms, and completing quests grant experience.
Your level is derived from accumulated XP. Level-ups improve maximum health,
maximum mana, attack, and defense, and restore your health and mana.

At level 3, use 'awaken <domain>' to choose one path:
  storm — lightning and explosive force
  tide  — crushing water and deep currents
  ember — ancient fire and direct destructive power
  aegis — protective force turned into offense

Use 'powers' to see the abilities available to your domain and level. Powers
have unlock levels and mana costs. At level 10, 'rebirth' becomes available.
Your first rebirth leads to modern Athens; later lives offer more destinations.
Read 'help rebirth_paths' and 'soul' before choosing a later path.`},
	"economy": {"economy", `GOLD, MERCHANTS, AND CRAFTING

Victories and quests can reward gold. Find NPCs marked [MERCHANT], then use
'shop' to inspect stock and 'buy <item>' to purchase. A purchase requires
enough gold and adds the item to your inventory. Check 'inv' afterward.
Use 'wield <name>' for weapons and 'wear <name>' for armor.

Certain monsters leave crafting materials, marked [MATERIAL] in inventory.
At Theron's Forge, type 'craft' to see recipes. 'craft <recipe>' consumes
the listed ingredient and gold only when all requirements are met. Crafted
gear goes into your inventory; it must be equipped to affect combat.

Consumables are sold by merchants. Read 'help consumables' for health and
mana supplies. Consider whether a supply or a lasting equipment upgrade is
the better use of your early gold.`},
	"world": {"world", `WORLD EXPLORATION

Fatewalker is a connected world of rooms, routes, eras, and story gates.
Use north, south, east, west, up, and down; short forms n, s, e, w, u, and d
also work. 'exits' lists routes, 'map' shows nearby connections, and
'worldmap' gives a broader view.

A route can be blocked by a story condition or an unresolved encounter. If
movement fails, read the response rather than repeating the command blindly.
Some enemies attack as you enter a room; tutorial creatures wait for you to
initiate combat. First-time room discovery can grant experience, but a room
does not award the same discovery reward every time you revisit it.

Read descriptions carefully. Landmarks, unusual objects, NPC dialogue, and
small details can hint at quests, secrets, and the consequences of a choice.`},
	"invoke": {"invoke", `LIFE-GIFTS

Use 'soul' to inspect your life-gift and whether it is ready. 'invoke' or
'gift' calls on that gift once during the current life. The gift is distinct
from your chosen divine domain and from ordinary powers.

Each gift has a different effect, tied to the era or destination of your
current life. It refreshes after rebirth. If it has already answered you
once, the soul view reports that it is spent until your next life. Consider
saving it for a moment when its effect will be useful.`},
	"status": {"status", `CHARACTER STATUS

Use 'score' or 'stats' to review your character. HP is health; at zero you
recover at the Gates of Olympus. Mana fuels divine powers. XP is accumulated
experience. Level describes growth in this life, while Life and Rebirths
track how often your soul has begun again.

Attack contributes to weapon attacks. The equipped weapon supplies damage.
Defense helps you avoid attacks, and armor increases your defense value.
Gold pays for merchant goods and crafting. Domain identifies your awakened
divine path; an empty domain means you have not chosen one. Divinity tracks
the soul's growth across lives.

For details, combine 'score' with 'inv', 'eq', 'powers', 'quests', 'journal',
and 'soul'.`},
	"crafting": {"crafting", `CRAFTING AT THERON'S FORGE

Defeated monsters can leave materials. Materials are marked [MATERIAL] in
your inventory and are ingredients, not weapons or armor. Carry the required
ingredient and enough gold to Theron's Forge, east of the Olympus foothills,
then use 'craft <recipe>'.

RECIPES
  craft stormguard cuirass — 1 Harpy Feather + 60 gold; rare armor (+7)
  craft laurel spear      — 1 Satyr Horn + 55 gold; rare spear (23 damage)
  craft manticore mantle  — 1 Manticore Spine + 90 gold; epic armor (+10)

Crafting consumes one required material and the recipe's gold, then adds
the finished item to your inventory. It never consumes ingredients or gold
when the recipe requirements are not met. Equip the result with 'wear' or
'wield' as appropriate. Type 'craft' at the forge to see recipes. Crafting
is only available at Theron's Forge.`},
	"consumables": {"consumables", `CONSUMABLES AND COMBAT RESOURCES

Use 'inv' to see what you carry. Consumables display their healing or mana
restoration values. Use 'use <item name>' to consume one; 'drink' and
'consume' are aliases. A successful use removes one item from your inventory.

AMBROSIA DRAUGHT — Restores up to 45 health.
BLUE LOTUS TONIC — Restores up to 20 mana.
LAUREL POULTICE — Restores up to 20 health.

Restoration never exceeds your maximum health or mana. If you are already
at full values, the item is not consumed. Consumables can be used during a
fight, but an aggressive living enemy gets its turn immediately afterward.
That makes timing matter: healing is useful, but not a free action.

Visit a merchant and type 'shop' to see available supplies. Use 'buy <name>'
to purchase an item, then 'inv' to verify that it is in your pack.`},
	"tutorial": {"tutorial", `THE FIRST THREAD — GUIDED PROLOGUE

Newly created characters receive a one-time, 39-lesson tutorial before
entering Asterion Village. The prologue teaches room descriptions, movement,
NPC dialogue, exits, maps, combat, passive versus aggressive creatures,
inventory, equipment, recovery, help files, quests, the journal, and
character progression. Progress is saved between connections.

Use tutorial to repeat the current instruction. Use help <topic> for details.
Practice creatures are initially passive and defend themselves only after
being attacked. Once slain, they are persistently flagged as defeated and
will not respawn for that character.

The Gate of First Steps is one-way. When you leave for Asterion Village,
the tutorial is closed permanently for that character. Existing characters
are not forced into it; it is assigned when a new character is created.`},
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
	if name == "give" { name = "offer" }
	if name == "" { name = "start" }
	if t, ok := topics[name]; ok {
		return t, true
	}

	// Resolve spell help from the combat power catalogue so that the details
	// shown here remain aligned with the actual ability definitions.
	compact := strings.ReplaceAll(name, " ", "")
	for _, domain := range []power.Domain{power.Storm, power.Tide, power.Ember, power.Aegis} {
		for _, p := range power.ForDomain(domain) {
			if strings.ToLower(strings.ReplaceAll(p.Name, " ", "")) != compact {
				continue
			}
			return Topic{Name: p.Name, Text: fmt.Sprintf(
				"%s\n\nDOMAIN: %s\nUNLOCK LEVEL: %d\nMANA COST: %d\nBASE DAMAGE: %d\n\n%s\n\nCASTING\n  cast %s\n\nCOMBAT QUICKLIST\n  ql set 1 %s\n  ql 1 <target>\n\nThe power must be unlocked for your current domain and character level. Casting consumes its normal mana cost and follows normal combat rules.",
				p.Name, p.Domain, p.UnlockLevel, p.ManaCost, p.Damage, p.Description, p.Name, p.Name,
			)}, true
		}
	}
	return Topic{}, false
}

func Names() []string {
	return []string{"start","movement","combat","combatmode","quicklist","guard","progression","economy","world","status","powers","inventory","equipment","items","put","give","map","automap","story","journal","quests","current","rebirth","talk","offer","hint","soul","invoke","shop","buy","crafting","consumables","tutorial","Stormspark","Thunderfall","Judgment of the Sky","Undertow","Leviathan Surge","Sea of the First Deep","Cinderbrand","Promethean Star","Titanfire","Aegis Pulse","Olympian Bulwark","Sovereign Ward"}
}
