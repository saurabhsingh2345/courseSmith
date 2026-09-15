# L71 — "Week 3 Day 1: Building Custom Claude Code Plugins and Marketplaces" (12:46)

**Entirely screen recording.** The longest lecture of Day 1 and the most structural.

## What he does, in order

| approx | on screen | what happens |
|---|---|---|
| 0:00 | talk | Why: a team can share plugins; **you can run your own marketplace inside your company** and colleagues add it to their Claude Code |
| 0:40 | file tree | New top-level folder — **`independent-reviewer/`**. It could live in a different repo entirely |
| 1:10 | file tree | Inside it, a folder with a **fixed** name: **`.claude-plugin/`**, holding **`plugin.json`** — the manifest |
| 1:50 | `plugin.json` | `{ name, description, version }` — "carry out an independent review of all changes since last commit" |
| 2:30 | talk | **A plugin directory can have exactly four subfolders**: `commands/`, `skills/`, `agents/` (sub-agents), and **`hooks/`** |
| 3:00 | new file | `hooks/` → **`hooks.json`**. He **cuts** the hook JSON out of `.claude/settings.json` and pastes it here — "we don't want it in both places" |
| 4:00 | talk | You *can* launch Claude pointed at one plugin, but it's inflexible. **The proper way is a marketplace** |
| 4:30 | file tree | A **top-level** `.claude-plugin/` (distinct from the one inside the plugin) containing **`marketplace.json`** — name, owner, email, and a **list of plugins** with their paths |
| 6:00 | `/plugin` | The menu. Marketplaces → only `claude-plugins-official` so far → **Add marketplace** → he types **`./`** for the local directory |
| 7:00 | menu | It reads the marketplace. `edtools` is added. Browse plugins → `independent-reviewer` → **space to select, enter to install** → **"install for all collaborators on this repository"** (project scope) |
| 8:00 | terminal | **Restart required.** Back in, `/plugin` → Installed shows `independent-reviewer` above the others |
| 8:40 | test | Deletes the README **and** `review.md`, then asks for a concise project readme. Claude writes it → **stop hook fires from the plugin, not from `settings.json`** (which he emptied) → Codex runs → `review.md` appears |
| 10:30 | `review.md` | It flags the new hook and worries about a **possible recursive loop**. He thinks no such risk exists |
| 11:00 | recap | The whole structure, said once more slowly |
| 11:40 | **sub-agent pros and cons** | **Pros**: parallelism · self-correcting negative feedback · context efficiency · one focused task done well. **Cons**: more moving parts and mysterious failures · **boundary/interface problems** · **compounding errors** when a divvied-up task goes wrong · **and they can cost MORE, not less** — more churn, more reviews, more turns |
| 12:46 | end | Tomorrow: sandboxing and remote execution. **"73% complete"** |

## For our version

- **This is the best structural lecture in Day 1** and it is entirely file-and-menu
  work — perfect for footage. The nested `.claude-plugin/` at two levels is the
  one genuinely confusing thing; **one small diagram of the folder tree** is
  justified, everything else is screen.
- **`cursor-agent -p -f`** in `hooks.json`.
- **Ours publishes the marketplace to our own GitHub**, so viewers can add it for
  real — he only does `./` locally. That is a leverage point: our version is
  strictly more useful.
- **Keep the recursive-loop worry.** A review agent reviewing the hook that
  spawned it is a genuinely interesting edge and it happened unprompted.
- **The pros-and-cons close is the one honest cost beat in Day 1** — especially
  "sub-agents can cost more, not less". Keep it whole; it pairs with the spend
  cap we wrote into W1 L13.
- **Footage: ~90%.**
