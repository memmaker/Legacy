# Legacy — Codebase Documentation

## Overview

**Legacy** is a tile-based role-playing game written in Go. It blends classic CRPG depth (dense world, dialogue, quests, party management) with roguelike sensibilities (quick character creation, loot variety, procedural dungeons). The game is built on the [Ebiten v2](https://ebitengine.org/) 2D game engine and uses [LDtk](https://ldtk.io/) as its level editor.

Design goals (from `DESIGN.txt`):
- A coffee-break RPG sandbox with a plot.
- Multiple skill-based approaches to every problem (violence, magic, stealth, persuasion, money, environment).
- No grinding, no level scaling, no useless skills.

---

## Technology Stack

| Dependency | Purpose |
|---|---|
| `github.com/hajimehoshi/ebiten/v2` v2.7.8 | 2D game engine (rendering, input, windowing) |
| `github.com/tidwall/gjson` v1.17.0 | JSON path querying (LDtk map data) |
| `github.com/Knetic/govaluate` v3.0.0 | Expression evaluation for dialogue conditionals |
| LDtk (`.ldtk` project file) | Hand-crafted level design |
| Custom `recfile` package | Record-file format used for save data and dialogue |

Go version: **1.21**

---

## Repository Structure

```
Legacy/
├── main.go           — Entry point; GridEngine struct, game loop, map loading, input routing
├── game.go           — Party menu, inventory, equipment, ranged/magic/skill menus
├── combat.go         — Turn-based CombatState; player and AI turns, hit/damage resolution
├── dialogue.go       — Conversation rendering, NPC speech window, scrollable text
├── generator.go      — Procedural dungeon map generator (accretion algorithm)
├── groundation.go    — Map transitions (wells, stairs, mirrors, portals)
├── helper.go         — General utility helpers for the engine
├── init.go           — Engine initialisation: renderer setup, UI overlay, party creation
├── input.go          — Keyboard/mouse input handling; command dispatch
├── interface.go      — Modal stack management, tooltip system
├── menus.go          — Context menus, examine, take/drop, use/interact, NPC actions
├── saveload.go       — Save/load serialisation (party + all visited maps)
├── stable.go         — Debug/cheat menu, XP table display
├── statusbar.go      — Status bar rendering (HP, mana, gold, food, clock, buffs)
│
├── game/             — Core game-logic package (engine-agnostic)
├── renderer/         — Rendering subsystem (tile atlas, grid renderer, animator)
├── ui/               — UI widgets (menus, conversation modal, inventory, tooltips)
├── gridmap/          — Generic grid map with FOV, pathfinding, tile/actor/item/object layers
├── geometry/         — 2D point, rect, FOV primitives
├── dungen/           — Dungeon generators (accretion, BSP)
├── ldtk_go/          — LDtk project/level/layer reader
├── recfile/          — Record-file reader/writer (save format, dialogue files)
├── ega/              — EGA colour palette constants
├── gocoro/           — Coroutine helpers for animations
├── util/             — Table layout, menu item helpers
├── bmpfonts/         — Bitmap font index helpers
│
├── assets/           — Game assets
│   ├── Legacy.ldtk   — LDtk map project (all hand-crafted levels)
│   ├── charset.png   — UI / character tileset
│   ├── entities.png  — Entity / NPC sprite atlas
│   ├── world.png     — World tile atlas
│   ├── dialogues/    — Dialogue script files (recfile format)
│   ├── npc/          — NPC definition files (recfile format)
│   └── scrolls/      — In-game readable scroll text files
│
├── saves/            — Save-game directory (runtime, not committed)
├── DESIGN.txt        — Design notes, TODO list, feature ideas
├── STORY.txt         — Narrative, world lore, quest outlines
└── SCALE.txt         — Balance targets (level cap, damage, HP, XP)
```

---

## Architecture

### Engine (`GridEngine`)

`main.go` defines the top-level `GridEngine` struct that implements the `ebiten.Game` interface. It owns:

- The active `gridmap.GridMap` (current map) and a cache of all maps visited this session.
- The `game.Party` (player's party), `game.Flags` (world-state flags), and `game.PlayerKnowledge`.
- A `CombatState` for turn-based combat.
- The renderer, animator, and UI modal stack.
- The LDtk project used to load hand-crafted levels.

### Game Logic (`game/`)

Engine-agnostic structs and rules. The `game.Engine` interface (`game/engine.go`) decouples all game objects from the top-level engine, allowing game logic to call back into the engine without direct coupling.

### Rendering (`renderer/`)

A **dual-grid renderer** displays a small grid (UI/text layer) overlaid on a larger world grid. Three texture atlases are used:
- `AtlasCharacters` — charset tiles (UI, borders, text glyphs)
- `AtlasWorld` — terrain tiles
- `AtlasEntities` / `AtlasEntitiesGrayscale` — actor and object sprites

### Maps (`gridmap/`, `ldtk_go/`, `generator.go`)

Maps are generic over actor, item, and object types. Hand-crafted maps are loaded from the LDtk project. Dungeon levels (`!gen_dungeon_*`) are procedurally generated using an accretion algorithm and optionally a BSP algorithm.

### Serialisation (`saveload.go`, `recfile/`)

Save games are stored in `saves/<slot>/`. The save format is a custom **recfile** (record-file) text format. A save contains:
- `party.rec` — party members, inventory, gold, food, flags, knowledge.
- `maps/<name>.rec` + `maps/<name>.bin` — state of every visited map (actors, items, objects, tile overrides).

---

## Core Game Systems

### Attributes

Every actor has seven primary **SPECIAL**-style attributes (scale 1–10, default 5) plus Level:

| Attribute | Effect |
|---|---|
| **Strength** | Base melee damage |
| **Perception** | Base ranged damage; ranged hit chance |
| **Endurance** | HP gained on level-up |
| **Charisma** | Social interactions |
| **Intelligence** | Magic effectiveness |
| **Agility** | Base armor; movement allowance; initiative |
| **Luck** | Critical hits and loot quality |

Derived attributes (computed from primaries + modifiers):
- **Movement Allowance** — tiles moved per turn (based on Agility minus armor encumbrance)
- **Initiative** — turn order in combat (`(Agility + Perception) × Level`)
- **Base Armor** — equals Agility plus equipment bonuses
- **Base Melee / Ranged Damage** — from Strength / Perception plus equipment

Temporary **modifiers** can be attached to any attribute or derived attribute (from status effects, equipment, spells).

### Skills

Skills have four levels: **Beginner → Advanced → Expert → Master** (stored as 1–4).

| Category | Skills |
|---|---|
| **Melee Combat** | Melee Combat, Backstab, Tackle |
| **Ranged Combat** | Ranged Combat |
| **Art of Theft** | Lockpicking, Pickpocket, Sneak, Glasscutting |
| **Social Skills** | Deceive, Persuade, Intimidate, Bluff, Spot Lies |
| **Outdoor Survival** | Survival, Hunting, Herbalism |
| **Athletics** | Climbing, Swimming |
| **Languages** | Common, Animals, Monsters |
| **Perception** | Spot Hidden, Danger Sense, Assess |
| **Other** | Tool Usage, Repair |

Skill checks use a **difficulty table** (8 levels from Trivial to Impossible). The success probability is adjusted by the relative difference between the actor's skill level and the check difficulty.

### Combat

Combat is **turn-based**. When the player attacks an NPC (or is attacked), `CombatState` activates:

1. All combatants are sorted by initiative.
2. On the player's turn: move, melee-attack an adjacent enemy, use a ranged weapon, cast a spell, or use active skills.
3. On the AI's turn: enemies pathfind toward their target and attack.
4. Combat ends when all opponents are dead, flee, or the player flees.

Hit resolution:
- **Melee**: chance based on attacker vs. defender melee skill levels.
- **Ranged**: chance based on attacker ranged skill vs. defender Agility/Perception.
- **Damage**: `max(baseDamage − armor, 1)`.

Backstab is available when the player sneaks and attacks from behind.

### Magic System

Spells require a **scroll** in the caster's inventory and consume **mana**. Spells implemented:

| Spell | Mana Cost | Effect |
|---|---|---|
| Nom De Plume | 0 | Rename the caster |
| Create Food | 10 | Create 10 food rations |
| Bird's Eye | 10 | Reveal a wider area of the map for several turns |
| Healing Word of Tauci | 10 | Heal caster for Level × 10 HP |
| Icebolt | — | Ranged damage (ice) |
| Fireball | 10 | Area-of-effect fire damage (radius 3), Level × 10 per tile |
| Raise as Undead | 10 | Animate a nearby corpse as an undead ally (range 4) |

Spells can be used in or out of combat. Each spell has an AI-evaluated **combat utility score** so NPCs can decide whether to cast it.

### Items & Equipment

**Equipment slots:**
- Weapons: Right Hand, Left Hand, Ranged
- Armor: Helmet, Breast Plate, Shoes
- Accessories: Robe, Ring (×2), Amulet (×2)
- Scroll slot (for equipped spells)

**Item types:** Weapon, Armor, Scroll, Potion, Key, Tool, Throwable, Chest (container), Flavor Item (readable/decorative), Pseudo-item, Tombstone, Lightsource.

Items can have embedded **actions** (context-menu interactions), **on-hit procs**, and **pickup events** that trigger game events.

### Status Effects

Status effects are applied to actors and may:
- Modify attributes or derived attributes (`AttributeModifier` / `DerivedAttributeModifier`).
- Modify skills (`SkillModifier`).
- Tick every real-time update (`RealTimeEffect`).
- Trigger each combat turn (`CombatEffect`).
- Respond to incoming damage (`OnDamageEffect`).
- Respond to the party resting (`OnRestEffect`).

Named effects include: `sleeping`, `undead`, `weak`, `holy bonus`, `blessed`.

### Dialogue System

Dialogue is stored in **recfile** files under `assets/dialogues/`. Each file defines `ConversationNode` entries keyed by trigger keywords. Features:
- **Conditionals** evaluated via `govaluate` expressions (flags, skills, items, gold).
- **Skill checks** with success/fail branches.
- **Effects** (set flags, give items, award XP, start combat).
- **Forced choices** (multiple-choice prompts).
- First-time vs. repeat conversation paths.
- Dialogue choices logged automatically to the in-game **Journal**.

### Flags & World State

`game.Flags` is a simple `map[string]int` used as a global event bus. Flags track:
- NPC conversation history (`talked_to_<npc>`).
- Quest progress.
- Criminal offenses.
- One-shot encounter triggers.

### Party System

The party holds up to ~4 members. The first member is the **avatar** (player character). Additional members are recruited NPCs that follow the avatar. Party-level resources:
- **Gold** (shared currency).
- **Food** (consumed on rest; starts at 10 rations).
- **Lockpicks** (consumed on lockpick attempts).
- A shared **multi-page inventory**.
- A **key ring** tracking named keys.

The party tracks a "steps before rest" counter. When it reaches zero, the party must rest (consuming food).

Party members can be split off for separate control via the **Split** menu option.

### Time System

The world has a **WorldTime** clock (minutes + days):
- Moving one step inside a level costs 1 minute.
- Moving one step on the world map costs 20 minutes.
- Days are 24 hours × 60 minutes. Years are 360 days.
- Time is displayed in the status bar.

### Line of Sight & Exploration

The map uses a **Field of View** (FOV) calculation (radius-6 square by default). Tiles outside the FOV are not rendered; tiles that have been seen but are outside the current FOV are rendered in a dimmed/remembered state.

### Searching

The party can **Search** a 3×3 area around them to reveal secret doors, hidden items, and hidden actors (using the Perception / Spot Hidden skill).

### Dungeon Generation

Procedural dungeons use an **accretion** algorithm (`dungen/accretion.go`): rooms are grown by repeatedly attaching new cells to the existing structure until the map is sufficiently filled. A **BSP** generator (`dungen/bsp.go`) is also available.

### Save / Load

Saves are stored in `saves/<slot>/` as recfile + binary pairs. The save captures:
- Party state: all member stats, skills, inventory, gold, food, flags, player knowledge.
- Map state: all maps loaded during the session, including current actor positions, item positions, and door/container states.

---

## UI Widgets (`ui/`)

| Widget | Purpose |
|---|---|
| `ConversationModal` | NPC dialogue, multiple-choice prompts, scrollable pages |
| `Menu` / `DialogueMenu` | Context menus, party menu, action menus |
| `Inventory` | Multi-page inventory grid |
| `Equipment` | Equipment slot viewer |
| `ScrollableText` | Full-screen scrollable text (books, scrolls) |
| `TextInput` | Single-line text prompt (used for mirror teleport, rename spell) |
| `Tooltip` | Delayed hover tooltip on map cells |
| `IconWindow` | Single-icon display panel |
| `MultiPage` | Tabbed/paged container |
| `Buttons` | Clickable UI button primitives |

A **modal stack** allows UI layers to be pushed and popped. Only the top modal receives input.

---

## Notable Game Objects (`game/`)

| File | Object |
|---|---|
| `door.go` | Locked/unlocked doors; key-based and lockpick-based opening |
| `chest.go` | Loot containers with level-scaled random loot |
| `key.go` | Keys that open specific named doors |
| `weapon.go` | Melee and ranged weapons with damage stats and on-hit effects |
| `armor.go` | Armor pieces with defense values and encumbrance |
| `potion.go` | Consumable potions (drink menu with effects) |
| `scroll.go` | Spell scrolls; equipping one enables the associated spell |
| `tool.go` | Tools used with the `Tool Usage` skill |
| `vehicle.go` | Rideable vehicles (can traverse water/land/mountains, affect step time) |
| `shrine.go` | Interactive shrines (meditation, stat bonuses) |
| `mirror.go` | Magic mirrors; player types a destination name to teleport |
| `well.go` | Map transition wells |
| `fireplace.go` | Decorative / interactive fireplaces |
| `barbecue.go` | Interactive cooking objects |
| `lightsource.go` | Torches and other light-emitting objects |
| `tombstone.go` | Readable tombstones |
| `flavor_item.go` | Decorative readable items (books, notes, signs) |
| `pseudo_items.go` | Abstract items (gold, food) that represent party resources |

---

## Known Locations (Maps)

Hand-crafted maps in the LDtk project:

| Map | Description |
|---|---|
| `Bed_Room` | Starting area (player spawn) |
| `WorldMap` | Overworld map |
| `Tauci_Castle` | Main castle area including the Throne Room |
| `Tauci_Mines_Level_1` | Mine dungeon with ladder transitions |
| `Tauci_Woods` | Forest area outside Tauci |
| `UI_Overlay` | Special LDtk level used to define the HUD tile overlay |

Mirror teleport destinations (typed by the player):

| Phrase | Destination |
|---|---|
| `celador` | WorldMap spawn |
| `tauci king` | Tauci Castle – Throne Room |
| `tauci mines` | Tauci Mines Level 1 – ladder up |
| `tauci woods` | Tauci Woods – entrance |

---

## Known Issues & Planned Features

From `DESIGN.txt`:

**Annoyances / bugs:**
- Saving/loading is rough; dialogue missing from NPCs after loading.
- NPC patrol continues during conversation/combat.
- NPCs cannot use skills or spells.
- Selling items not fully implemented.
- Well transitions need fixing.

**Planned content:**
- More spell effects: random teleport, enchant weapon/armor, summon monster, invisibility, fire propagation, lightning bolt, poison cloud.
- Wands and potions as consumable spells with charges.
- Unique attack patterns per weapon type.
- A full leveling system with trainers.
- Vendors (armoury, weapon smith, alchemy, general store, black market, bank).
- XP awards for non-combat activities.
- Quick slots UI.
- Better targeting mode.

**Balance targets** (from `SCALE.txt`):
- Max player level: **20** (~500k total XP).
- Max HP at level 20: **~100**.
- End-game weapon damage: **100–150**; legendary: **150–200**.
- End-game armor: **80–100**; legendary: **100–120**.
