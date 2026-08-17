package pipeline

// Bios: what a template tells a director about itself.
//
// The catalog already carries copy for a PERSON — Title, Description and
// Example answer "is this the look I want?" for somebody scrolling a gallery,
// and they answer it well. A director choosing between eighty-one templates
// needs something else, and it is the thing that was missing: not what a
// template looks like, but what it must be FILLED with, and when it must not be
// picked at all.
//
// The gap was expensive and it was visible in the finished videos. Every
// template has a validator that rejects material it cannot express, and a
// caster that cannot see those requirements picks looks it cannot fill: `gauge`
// on a subject with no ceiling, `costing` where nothing adds up, `trace` where
// nothing is shared. That failure lands late — after the cast, after a planning
// call, after three correction rounds — and its shipped form is a segment that
// got salvaged into something thin and did not belong in the piece.
//
// The old fix was a paragraph in the cast prompt naming four templates by hand.
// It worked for those four and told the model nothing about the other
// seventy-seven, which is how a hand-maintained list always ends: correct about
// what somebody remembered.
//
// == Why the bios are written here rather than at each registration site ==
//
// Category is declared at the registration site precisely so it cannot be
// forgotten, and the same argument would put Bio there. It is the wrong call for
// this field, for a reason particular to what a bio is for.
//
// A bio is read COMPARATIVELY. The director sees all eighty-one at once and its
// whole job is to tell them apart, so what matters is not that each bio is good
// but that they are pitched alike — the same grain of specificity, the same kind
// of noun. A bio written alone in its own file drifts, and inconsistent
// specificity is exactly what makes a model choose wrongly: given one
// requirement written as "a numeric ceiling with a unit and two to five things
// measured against it" and another as "some data", it will reach for the second
// every time, because the second sounds satisfiable and the first sounds like
// work.
//
// So they are written side by side, where a drifting one is visible. The
// "cannot be forgotten" property is kept by the guard rather than by the
// location: registerSnippetTemplate panics on a template with no bio, exactly
// as it panics on one with no category.

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// TemplateBio is a template's self-description, written for a machine that has
// to choose between eighty-one of them.
//
// Three fields, and each one exists because a specific bad cast happens without
// it. Needs is what the template's own validator will demand. Avoid is the
// subject it gets wrongly reached for. Roles is where in an arc it can sit.
type TemplateBio struct {
	// Needs is the concrete material this template must be filled with, written
	// as the thing a director would have to be able to name.
	//
	// Written as a REQUIREMENT, not as a description: "a numeric ceiling with a
	// unit and 2-5 things measured against it" rather than "numbers". The
	// difference is whether a director can check its own pick before committing
	// to it, which is the entire mechanism here — a template whose material you
	// cannot name is a template you have chosen wrongly, and that test only
	// works if the requirement is specific enough to fail.
	Needs string
	// Avoid is the subject this template is wrongly reached for, and why it
	// fails there. Empty only for the handful of templates that genuinely carry
	// anything.
	//
	// Separate from Needs because "what it requires" and "what it attracts" are
	// different failures. A director reading only Needs picks `gauge` for a
	// subject with a number in it and discovers at plan time that a number is not
	// a ceiling. The name is the trap, so the trap is named.
	Avoid string
	// Roles are the jobs in the arc this template can carry — RoleHook,
	// RoleDevelop, RolePayoff.
	//
	// The arc is already enforced (the piece opens on a hook, closes on a payoff,
	// and does not go backwards), and until now the caster satisfied it by
	// labelling whatever it had picked. That is the rule being obeyed in name:
	// `anatomy` declared a hook is still a clip that opens by taking a URL apart,
	// which puts nothing at stake. Declaring the affinity here makes the label
	// checkable against the template rather than against itself.
	Roles []string
	// Figures marks a template that cannot be planned without real, specific
	// numbers — the ones that must never be cast over a subject whose figures
	// were looked for and not found.
	//
	// This is the list that used to live as a hardcoded paragraph in the cast
	// prompt, naming four templates. It is a property of a template, so it lives
	// on the template, and adding a fifth data-hungry look now means setting a
	// bool rather than remembering a prompt exists.
	Figures bool
}

// CanCarry reports whether this template can do the given arc role.
func (b TemplateBio) CanCarry(role string) bool {
	return slices.Contains(b.Roles, role)
}

// templateBios is every registered template's bio, keyed by name.
//
// Grouped in catalog order rather than alphabetically, because that is how they
// are read: the reason to keep `gauge`, `metric`, `costing` and `budget` on
// adjacent lines is that their requirements have to be told apart, and they can
// only be told apart next to each other.
var templateBios = map[string]TemplateBio{

	// ---- Numbers & scale -------------------------------------------------
	//
	// Every template in this group can be planned only from figures that exist,
	// which is why nearly all of them set Figures. This is the group a director
	// over-reaches for: a subject that MENTIONS a quantity reads as a subject
	// that HAS one, and the difference is the whole cast.

	"budget": {
		Needs:   "a fixed total with a unit, and 3-6 named claims taken out of it that leave a remainder worth remarking on",
		Avoid:   "open-ended spend. If nothing is capped there is nothing to run out, and the clip draws a pot that never empties",
		Roles:   []string{RoleDevelop, RolePayoff},
		Figures: true,
	},
	"carry": {
		Needs:   "two binary operands and the addition worked out column by column, with the decimal equivalent as the check",
		Avoid:   "arithmetic you have not actually performed. The carries and the decimal check are drawn from the numbers, so a sum that does not add up is visible on screen",
		Roles:   []string{RoleDevelop},
		Figures: true,
	},
	"costing": {
		Needs:   "3-6 line items with real amounts that add up to a stated total",
		Avoid:   "a cost you can only estimate. The total is built in front of the viewer, so invented line items are stated as fact with a receipt's authority",
		Roles:   []string{RoleDevelop, RolePayoff},
		Figures: true,
	},
	"data": {
		Needs:   "a real series, spread or geographic distribution — actual values with units, enough of them to have a shape",
		Avoid:   "a subject with no dataset behind it. A chart of invented points is a false claim with axes on it",
		Roles:   []string{RoleDevelop},
		Figures: true,
	},
	"gauge": {
		Needs:   "one numeric ceiling with a unit, and 2-5 things measured against it — some of which fit and some of which do not",
		Avoid:   "any subject with no threshold in it. 'How AI lets people build faster' contains no ceiling and nothing to measure, and is the commonest miscast in the catalog",
		Roles:   []string{RoleDevelop},
		Figures: true,
	},
	"growth": {
		Needs:   "2-4 named complexity classes and one real input size to read their cost at",
		Avoid:   "a subject whose cost does not change with scale. Without a size to probe, every curve is a decoration",
		Roles:   []string{RoleDevelop},
		Figures: true,
	},
	"ladder": {
		Needs:   "the memory hierarchy's own rungs, with real latencies and capacities from register to network",
		Avoid:   "any hierarchy that is not the memory one. The rungs and their numbers are the subject, not a metaphor to borrow",
		Roles:   []string{RoleDevelop},
		Figures: true,
	},
	"latency": {
		Needs:   "3-6 operations with durations that span at least two orders of magnitude, with units",
		Avoid:   "durations that are all about the same size. A log axis with everything inside one decade draws a straight line and says nothing",
		Roles:   []string{RoleDevelop},
		Figures: true,
	},
	"metric": {
		Needs:   "3-6 real figures with units, each big or surprising enough to hold an almost empty frame on its own",
		Avoid:   "a subject you have no measurements for. This template puts one number alone on screen and counts it up, which states it as fact harder than any other look here",
		Roles:   []string{RoleHook, RoleDevelop},
		Figures: true,
	},
	"multiply": {
		Needs:   "a per-unit figure, the count it is multiplied by, and the product — all three known",
		Avoid:   "a total that is not genuinely a product. The whole effect is the small number becoming a large one, and it needs both factors",
		Roles:   []string{RoleHook, RoleDevelop},
		Figures: true,
	},
	"occupancy": {
		Needs:   "a countable fixed population and the number of it that is claimed",
		Avoid:   "a proportion with no population behind it. The grid draws individual units, so 'most of it' cannot be rendered",
		Roles:   []string{RoleDevelop},
		Figures: true,
	},
	"radix": {
		Needs:   "one value with its binary and hex forms, and place-value columns that sum back to it",
		Avoid:   "any subject that is not number bases themselves. The three columns are the lesson, not a way to show a number",
		Roles:   []string{RoleDevelop},
		Figures: true,
	},
	"ratio": {
		Needs:   "two measurements in the same unit and the proportion between them, stated as the line worth remembering",
		Avoid:   "two numbers in different units, or a proportion that is not striking. 'A third of' has to land harder than the pair did",
		Roles:   []string{RoleHook, RoleDevelop},
		Figures: true,
	},
	"scale": {
		Needs:   "3-5 quantities in one unit, each roughly ten times the last",
		Avoid:   "numbers of similar size. Pulling the camera back an order of magnitude requires orders of magnitude to pull back through",
		Roles:   []string{RoleHook, RoleDevelop},
		Figures: true,
	},

	// ---- Ideas & mental models -------------------------------------------
	//
	// The group a director should reach for when the subject is explanatory
	// rather than numeric. These carry a topic without needing data, which makes
	// them the honest answer whenever the figures were looked for and not found.

	"analogy": {
		Needs: "a familiar thing whose parts map one-to-one onto the real one — 3-5 pairs that each hold",
		Avoid: "an analogy that matches only in mood. A mapping that breaks on the second pair installs a wrong model that is harder to remove than the gap it filled",
		Roles: []string{RoleHook, RoleDevelop},
	},
	"bitfield": {
		Needs: "a real bit layout with named fields and their widths, and what each field decodes to",
		Avoid: "a number that is only a number. This draws layouts — sign and exponent, an address, permission bits — not values",
		Roles: []string{RoleDevelop},
	},
	"constellation": {
		Needs: "one central idea and 4-5 properties that genuinely radiate from it rather than following each other",
		Avoid: "a sequence. Anything whose parts come in an order wants a flow or a relay; this draws a thing and its aspects at once",
		Roles: []string{RoleDevelop, RolePayoff},
	},
	"encode": {
		Needs: "one character, its codepoint, and its actual UTF-8 bytes with the marker bits identified",
		Avoid: "any subject that is not text-as-data. The three stations are the lesson",
		Roles: []string{RoleDevelop},
	},
	"eras": {
		Needs: "3-5 named generations with their decades, one defining artifact each, and what each handed to the next",
		Avoid: "a history with no causation. Dates and names with nothing passing between them is trivia, and the arcs will have nothing to draw",
		Roles: []string{RoleHook, RoleDevelop},
	},
	"gates": {
		Needs: "one boolean gate and its complete truth table",
		Avoid: "logic in the loose sense — business rules, conditionals in code. This draws a circuit and fills a truth table from it",
		Roles: []string{RoleDevelop},
	},
	"illustration": {
		Needs: "a subject and one line worth saying. No data, no structure, no vocabulary — this is the look that can carry anything",
		Avoid: "",
		Roles: []string{RoleHook, RoleDevelop, RolePayoff},
	},
	"myth": {
		Needs: "a belief this audience genuinely holds, phrased in the words they would use, and what is true instead",
		Avoid: "a strawman. The whole effect depends on the viewer recognising the thought as their own, so a belief nobody actually holds reads as condescension",
		Roles: []string{RoleHook},
	},
	"stepper": {
		Needs: "an array of concrete values and the algorithm's actual comparisons and swaps over them, in order",
		Avoid: "an algorithm you have not stepped through by hand. The pointers and the counter are drawn from the real trace",
		Roles: []string{RoleDevelop},
	},
	"whiteboard": {
		Needs: "a subject that suits being thought through out loud, and 3-6 marks worth making in sequence",
		Avoid: "a finished structure. If the diagram is already settled, a drawn one reads as a worse version of it",
		Roles: []string{RoleDevelop},
	},

	// ---- Systems & process -----------------------------------------------
	//
	// Almost all develop-only, and that is a fact about the group rather than an
	// oversight: a clip that shows how something works has already assumed the
	// viewer wants to know, which is what a hook is for.

	"anatomy": {
		Needs: "one artefact and 3-6 labelled parts of it",
		Avoid: "several things at once. This holds a single object still and reaches into it; a comparison or a sequence belongs elsewhere",
		Roles: []string{RoleDevelop},
	},
	"blueprint": {
		Needs: "named blocks with ports, the wires between them, and one path worth lighting up",
		Avoid: "a process. This draws a standing structure — who is wired to whom — not what happens in what order",
		Roles: []string{RoleDevelop},
	},
	"breakdown": {
		Needs: "3-6 phases in order, each with detail worth opening on its own",
		Avoid: "phases that are only names. If a stage has nothing inside it, the opening lands on an empty panel",
		Roles: []string{RoleDevelop},
	},
	"callstack": {
		Needs: "a recursive function, its argument at each depth, the base case, and what unwinds back",
		Avoid: "iteration. A loop has no frames to push, so there is no stack to draw breathing",
		Roles: []string{RoleDevelop},
	},
	"capabilities": {
		Needs: "a boundary, the things denied outside it, and the one or two deliberately granted",
		Avoid: "a permission model with no denials. The frame is what is refused; a boundary that grants everything draws an empty ring",
		Roles: []string{RoleDevelop},
	},
	"cycle": {
		Needs: "3-6 stages that genuinely close a ring, and what is different on the next lap",
		Avoid: "a process that ends. If the last stage does not feed the first, this is a relay or a breakdown",
		Roles: []string{RoleDevelop},
	},
	"flow": {
		Needs: "named components and the traffic that moves between them",
		Avoid: "a subject with one component. Boxes and edges need at least three things connected to be worth drawing",
		Roles: []string{RoleDevelop},
	},
	"fork": {
		Needs: "two processes over shared memory and the specific write that splits a page",
		Avoid: "sharing with no copy. The moment being drawn is the split, so a subject that never diverges has no beat to land on",
		Roles: []string{RoleDevelop},
	},
	"handshake": {
		Needs: "two named parties and the ordered messages between them, each accomplishing something specific",
		Avoid: "a one-way call. Two lifelines with traffic in one direction is a flow diagram drawn the hard way",
		Roles: []string{RoleDevelop},
	},
	"history": {
		Needs: "a commit sequence that actually branches and merges",
		Avoid: "linear history. The lanes, the divergence and the join are the subject; a straight line of commits is a timeline",
		Roles: []string{RoleDevelop},
	},
	"journal": {
		Needs: "an append-only record with real entries, and the replay that rebuilds state from them",
		Avoid: "a log nobody reads back. Both halves are needed — the growing file and the rebuild",
		Roles: []string{RoleDevelop},
	},
	"journey": {
		Needs: "a route's named stops and what each one adds on the way out and back",
		Avoid: "a trip with one hop. The point is that many parties touch it and no single one holds the answer",
		Roles: []string{RoleDevelop},
	},
	"labcard": {
		Needs: "a task, the tools that must be installed, numbered steps, and what the screen looks like when it worked",
		Avoid: "an exercise with no observable end state. Without a success condition the viewer cannot tell whether they finished",
		Roles: []string{RoleDevelop, RolePayoff},
	},
	"layers": {
		Needs: "stacked named levels and one boundary a payload visibly crosses",
		Avoid: "levels with no crossing. The whole clip is what it costs to cross the line, so a static stack has no beat",
		Roles: []string{RoleDevelop},
	},
	"lookup": {
		Needs: "a key and the chain of tables that resolve it, each contributing part of the answer",
		Avoid: "a single lookup. One table answering in one step is a definition, not a resolution chain",
		Roles: []string{RoleDevelop},
	},
	"machine": {
		Needs: "a physical machine and 3-6 of its named parts, each with a one-line job",
		Avoid: "anything abstract. This draws a chassis with hardware in it; a software architecture wants blueprint",
		Roles: []string{RoleDevelop},
	},
	"multiplex": {
		Needs: "a pool of identical waiting sources, several going ready at once, and one worker taking them in a pass",
		Avoid: "work that genuinely needs parallelism. The point is that one thread is enough, which a subject requiring many contradicts",
		Roles: []string{RoleDevelop},
	},
	"pipeline": {
		Needs: "named stages, several items in flight at once, and the throughput arithmetic at the end",
		Avoid: "a sequence run one item at a time. Without overlap there is nothing in flight and the whole lesson is missing",
		Roles: []string{RoleDevelop},
	},
	"ranking": {
		Needs: "an ordered board with real entries and the specific change that re-sorts it",
		Avoid: "a static list. The subject is the movement — an eviction, a new entrant — not the order itself",
		Roles: []string{RoleDevelop},
	},
	"regions": {
		Needs: "a process's memory regions and what specifically runs out in one of them",
		Avoid: "memory in the abstract. The column is code, static data, heap and stack; a subject that is not address space has nothing to fill it",
		Roles: []string{RoleDevelop},
	},
	"relay": {
		Needs: "an ordered chain of hand-offs where the ORDER is the lesson, each stage doing one job",
		Avoid: "stages that could happen in any order. If the sequence does not matter, this draws a false constraint",
		Roles: []string{RoleDevelop},
	},
	"scheduler": {
		Needs: "several processes, their turns on one shared time axis, and where the context switches land",
		Avoid: "scheduling in the calendar sense. This is CPU time-sharing, and the lanes are processes competing for one resource",
		Roles: []string{RoleDevelop},
	},
	"states": {
		Needs: "a handful of states, the named event on each transition, and a path worth walking",
		Avoid: "a thing that can be in two states at once. The token is single, so anything concurrent is drawn as a lie",
		Roles: []string{RoleDevelop},
	},
	"table": {
		Needs:   "a real spec sheet with several rows, and the one row that actually decides the question",
		Avoid:   "a sheet where every row matters. The move is the strip-back to one line, so a table without a buried answer has no punchline",
		Roles:   []string{RoleHook, RoleDevelop},
		Figures: true,
	},
	"timeline": {
		Needs: "dated milestones in order, 3-6 of them, where the sequence is the point",
		Avoid: "events with no dates. The spine is chronological, and undated marks make it a list",
		Roles: []string{RoleDevelop},
	},
	"trace": {
		Needs:   "named actors, one shared value with a starting state, and the operations that change it in a specific interleaving",
		Avoid:   "a subject with no shared mutable state. No race, no lost update, no lock — nothing to drain a step at a time",
		Roles:   []string{RoleDevelop},
		Figures: true,
	},

	// ---- Code & screens ---------------------------------------------------

	"canvas": {
		Needs: "app nodes wired into a chain and a real record travelling it",
		Avoid: "code. This is a builder's canvas — no-code automations and integrations — not an editor",
		Roles: []string{RoleDevelop},
	},
	"footage": {
		Needs: "an actual recording that exists, and moments in it you can point at by timestamp",
		Avoid: "any claim you have no recording of. This template draws nothing itself, and it is the one look whose rule is about truth rather than shape — it will not ship a near miss",
		Roles: []string{RoleDevelop},
	},
	"mockup": {
		Needs: "a screen and the layers it assembles from, in the order they land",
		Avoid: "a screen you cannot describe element by element. The frame fills in piece by piece and needs the pieces",
		Roles: []string{RoleDevelop},
	},
	"promptloop": {
		Needs: "a goal held throughout, and 3+ rounds of ask, look at what came back, ask again",
		Avoid: "a single prompt and its answer. One round is a screenshot; the loop is the lesson",
		Roles: []string{RoleDevelop},
	},
	"shell": {
		Needs: "a command with its flags, the output it prints, and the one flag worth a caption",
		Avoid: "a command whose output you are guessing at. The terminal is authored rather than executed here, so a wrong output ships as if it were real",
		Roles: []string{RoleDevelop},
	},
	"spec": {
		Needs: "acceptance criteria that could be written before the build and checked off after",
		Avoid: "criteria that are not observable. 'Works well' cannot be ticked",
		Roles: []string{RoleDevelop},
	},
	"vscode": {
		Needs: "code that actually runs, and the output the interpreter really produced",
		Avoid: "code you have not run. The output is executed and verified, so a snippet that does not work fails the segment rather than faking it",
		Roles: []string{RoleDevelop},
	},
	"workspace": {
		Needs: "a project of several files with a program that runs across them",
		Avoid: "a single snippet. If it fits in one file, vscode does it better and cheaper",
		Roles: []string{RoleDevelop},
	},

	// ---- Choices & verdicts ----------------------------------------------
	//
	// This group holds most of the catalog's payoff capacity, which is why a
	// director short of a closer should look here first.

	"cards": {
		Needs: "2-5 things that have real names in the world, a line saying what each one is, and whether they are alternatives, a sequence, or just the players",
		Avoid: "abstractions. The cards wear the real logos, so a concept with no brand behind it draws a row of generic glyphs — that is what constellation and rundown are for",
		Roles: []string{RoleHook, RoleDevelop, RolePayoff},
	},
	"duel": {
		Needs: "two named products, the ONE axis the choice between them turns on, a number for each of them on it, and which one you would tell somebody to use",
		Avoid: "a choice that turns on several things at once. Two bars can only say one thing, and two bars of the same length say nothing — if they are level on the axis, use versus and compare them across five dimensions instead",
		Roles: []string{RoleDevelop, RolePayoff},
	},
	"spotlight": {
		Needs: "one named product and two to four things it is actually good FOR, each said in nine words or less",
		Avoid: "a feature list. \"Ten gigabyte context window\" is a spec and belongs on a showcase; this template wants what the spec is for. Also avoid it for a tool the course has not met yet if there is runtime for the long introduction — that is showcase",
		Roles: []string{RoleHook, RoleDevelop},
	},
	"opener": {
		Needs: "a subject that can be said in four to nine words, and one line saying what the viewer will be able to do afterwards",
		Avoid: "using it anywhere but the front. It is a title page: no list, no outcomes, no agenda — and a short subject kills it, because the big type is filled BY the words and three of them leave the frame empty",
		Roles: []string{RoleHook},
	},
	"changeplan": {
		Needs: "two to six real file paths and, for each, what CHANGES in it — not what the file is",
		Avoid: "a plan whose rows restate their own filenames. Also avoid it for a change in a single file: one row is an edit, not a plan, and patch draws that better",
		Roles: []string{RoleDevelop, RolePayoff},
	},
	"patch": {
		Needs: "one file and one to four small changes in it — the lines before, the lines after, and why each change was made",
		Avoid: "a large rewrite. The whole premise is that ONE change is readable at nearly twice the size a diff is normally set at, so five moved lines is the ceiling; a rewritten function is the code template showing the finished state",
		Roles: []string{RoleDevelop, RolePayoff},
	},
	"approval": {
		Needs: "one concrete action a tool is asking permission for, and two or three answers with what each one actually hands over",
		Avoid: "a vague ask. \"The agent wants to make changes\" is the dialog everybody clicks through; the frame only teaches if the request is specific enough to stop at",
		Roles: []string{RoleDevelop, RolePayoff},
	},
	"compare": {
		Needs: "two named options and the dimensions they genuinely differ on",
		Avoid: "two things that are not alternatives. A comparison implies a choice; without one it is two definitions in one frame",
		Roles: []string{RoleDevelop},
	},
	"decision": {
		Needs: "one question that separates the options, an axis split into bands, and an instruction at the end of each band",
		Avoid: "a choice that turns on several questions at once. The axis is single, so a multi-factor decision gets drawn as a simpler one than it is",
		Roles: []string{RoleDevelop, RolePayoff},
	},
	"mission": {
		Needs: "a goal, a spec checklist, the artifact to hand in, and an observable definition of done",
		Avoid: "an assignment with no deliverable. The card is a brief, and a brief without an artifact is advice",
		Roles: []string{RolePayoff},
	},
	"pitfall": {
		Needs: "the mistake people actually make at one specific step, the symptom that shows it happened, and the single move that fixes it",
		Avoid: "a general warning. Without a symptom the viewer cannot tell whether they have the problem",
		Roles: []string{RoleDevelop},
	},
	"showcase": {
		Needs: "one product, what it costs, and what it is genuinely good and bad at",
		Avoid: "a tool you have no price or weakness for. A card with only strengths reads as an advertisement",
		Roles: []string{RoleDevelop},
	},
	"stack": {
		Needs: "tiers of a build and the named tools that live in each",
		Avoid: "tools with no layering between them. If everything sits at the same level, this is a list",
		Roles: []string{RoleDevelop},
	},
	"toggle": {
		Needs: "a question the viewer arrived already asking, a one-word answer, and the caveats that complicate it",
		Avoid: "a question nobody is asking yet. The answer lands in the first three seconds, which only works if the viewer was waiting for it",
		Roles: []string{RoleHook},
	},
	"verdict": {
		Needs: "where a recommendation holds, where it breaks, and the call to leave on screen",
		Avoid: "a summary. This closes on an instruction; restating what was covered is what a constellation is for",
		Roles: []string{RolePayoff},
	},
	"versus": {
		Needs: "two contenders, 3-5 dimensions they are judged on, and a ruling on when to reach for which",
		Avoid: "a comparison with no forced choice. The spine implies the viewer must pick one, so two things used together belong elsewhere",
		Roles: []string{RoleHook, RoleDevelop, RolePayoff},
	},

	// ---- Presenting & pacing ---------------------------------------------
	//
	// The connective tissue. Most of these are openers by construction — a
	// lesson's contract, its floor, what came before — which is why this group
	// carries most of the catalog's hook capacity.

	"bridge": {
		Needs: "what the previous lesson established, and the specific gap this one fills",
		Avoid: "a lesson that stands alone. Sliding in a prior that does not exist opens on a hand-off from nowhere",
		Roles: []string{RoleHook},
	},
	"cast": {
		Needs: "a subject that lands better from a person's reaction than from a diagram, and an attitude to it per beat",
		Avoid: "anything structural. The frame is a character and a headline, so a diagram's worth of content has nowhere to go",
		Roles: []string{RoleHook, RoleDevelop, RolePayoff},
	},
	"chapter": {
		Needs: "the ordinal of the section starting, its name, and the sections either side of it",
		Avoid: "a piece with no sections. Without a path behind and ahead, the huge number refers to nothing",
		Roles: []string{RoleHook},
	},
	"checkpoint": {
		Needs: "a few steps the viewer performs now, and the condition that tells them it worked without anyone marking it",
		Avoid: "a check that needs grading. The whole point is self-verification",
		Roles: []string{RolePayoff},
	},
	"drill": {
		Needs: "one sharp question, 3-4 options where the wrong ones are misconceptions worth killing, and a one-line why",
		Avoid: "distractors that are obviously wrong. Striking through an implausible option teaches nothing",
		Roles: []string{RoleDevelop, RolePayoff},
	},
	"objective": {
		Needs: "2-4 things the viewer will be able to do afterwards, each with the evidence that would show it",
		Avoid: "outcomes phrased as topics. 'Understand recursion' has no evidence; 'trace a recursive call to its base case' does",
		Roles: []string{RoleHook},
	},
	"outcome": {
		Needs: "2-4 observable abilities, each with a reason it matters on the job",
		Avoid: "abilities with no stated payoff. This is the opener that has to earn attention, not just record a contract",
		Roles: []string{RoleHook},
	},
	"prereq": {
		Needs: "each thing the lesson assumes, whether the course taught it or the viewer brings it, and which are skippable",
		Avoid: "a lesson that assumes nothing. An empty floor is a slide saying you are ready",
		Roles: []string{RoleHook},
	},
	"quiz": {
		Needs: "a question, its answer, and why each wrong option tempted",
		Avoid: "a question with one plausible answer. Without tempting distractors there is nothing to think about in the pause",
		Roles: []string{RoleDevelop, RolePayoff},
	},
	"recap": {
		Needs: "what several earlier lessons established, each tagged with where it came from, and the one thread joining them",
		Avoid: "a recap of a single prior lesson — that is a bridge — or of ground this piece has not covered",
		Roles: []string{RoleHook},
	},
	"rundown": {
		Needs: "exactly N things that genuinely form a set, N between 3 and 5, each worth its own beat",
		Avoid: "a set you have to pad to reach N. The clip promises a count in its first second and delivers precisely that, so a weak fifth item is visible",
		Roles: []string{RoleHook, RoleDevelop},
	},
	"spine": {
		Needs: "the narration itself, or a subject to write it from. Every line becomes its own shot, so it carries connective material no other template has a home for",
		Avoid: "",
		Roles: []string{RoleHook, RoleDevelop, RolePayoff},
	},
	"story": {
		Needs: "an arc worth a minute or two — a character who acts, a change, and a resolution",
		Avoid: "a single fact. This is the catalog's only multi-shot film and it needs somewhere to travel",
		Roles: []string{RoleHook, RoleDevelop, RolePayoff},
	},
	"syllabus": {
		Needs: "4-8 modules as stops on a route, with the current one identified",
		Avoid: "a course with no shape yet. The route needs its whole length to draw progress along it",
		Roles: []string{RoleHook},
	},
}

// BioFor returns a template's bio.
func BioFor(name string) (TemplateBio, bool) {
	b, ok := templateBios[name]
	return b, ok
}

// checkTemplateBio is the guard registerSnippetTemplate runs beside the
// category and family ones.
//
// It panics for the same reason those do, and one more: a template with no bio
// is invisible to the director rather than merely hard to find. It would still
// appear in the catalog it is handed — with no requirements, no trap named and
// no role — which reads as the easiest template in the list to satisfy. A
// missing bio does not make a template unpickable; it makes it the DEFAULT pick,
// which is the worst failure available here.
func checkTemplateBio(t *SnippetTemplate) {
	bio, ok := templateBios[t.Name]
	if !ok {
		panic(fmt.Sprintf("snippet template %q has no bio — add one to templateBios in snippet_bio.go so a director can tell it apart from the other %d", t.Name, len(templateBios)))
	}
	if strings.TrimSpace(bio.Needs) == "" {
		panic(fmt.Sprintf("snippet template %q has a bio with no Needs — say what material it must be filled with, or a director cannot check its own pick", t.Name))
	}
	if len(bio.Roles) == 0 {
		panic(fmt.Sprintf("snippet template %q has a bio with no Roles — say whether it can hook, develop or pay off", t.Name))
	}
	for _, r := range bio.Roles {
		if _, ok := castRoleOrder[r]; !ok {
			panic(fmt.Sprintf("snippet template %q declares role %q, which is not one of %s, %s, %s", t.Name, r, RoleHook, RoleDevelop, RolePayoff))
		}
	}
	t.Bio = bio
}

// OrphanBios returns bios written for templates that do not exist.
//
// The registration guard catches a template with no bio; this catches the other
// direction, which the guard structurally cannot: a bio for a template that was
// renamed or removed sits in the table looking maintained. Exported because it
// is checked by a test rather than at init — a stale entry is harmless at
// runtime and should not stop the binary starting.
func OrphanBios() []string {
	var out []string
	for name := range templateBios {
		if _, ok := SnippetTemplates[name]; !ok {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

// TemplatesForRole returns the offered templates that can carry an arc role.
//
// Used by the director to answer the question the arc rules make unavoidable:
// the piece must open on a hook and close on a payoff, so before anything is
// cast it is worth knowing which looks can actually do those jobs. Without it a
// director satisfies the arc by mislabelling — calling an `anatomy` a hook — and
// the check passes on a clip that puts nothing at stake.
func TemplatesForRole(role string, pool []*SnippetTemplate) []*SnippetTemplate {
	out := make([]*SnippetTemplate, 0, len(pool))
	for _, t := range pool {
		if t.Bio.CanCarry(role) {
			out = append(out, t)
		}
	}
	return out
}
