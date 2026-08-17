// Template families, mirroring snippet_categories.go.
//
// A family is which gallery offers a template, not what the template is for —
// the categories are still the vocabulary a card is filed under, and the replica
// batch spans several of them. Kept here rather than inline in the two pages so
// the strings agree with Go in one place.

/** The catalog offered on the snippets page. Empty, so no registration site changed. */
export const FamilyCore = "";

/** The batch cut to the reference videos' visual grammar. */
export const FamilyReplica = "replica";

/**
 * The house style the replica batch assumes.
 *
 * These templates are drawn for a near-black stage with a lot of air and
 * standing chrome. Rendered in the default skin they still work, but they read
 * as a different production — so the gallery fixes the skin rather than offering
 * it as one more choice nobody has an opinion about on arrival.
 */
export const ReplicaSkin = "broadcast" as const;

/** The batch built to teach computer science itself — the CS foundations course. */
export const FamilyFoundations = "foundations";

/**
 * The house style the foundations batch assumes.
 *
 * These templates are diagram-dense — bit rows, block diagrams, packet maps —
 * and they are composed against the editorial skin's hard left axis, where the
 * headline holds one edge and the picture is allowed to be unbalanced. As with
 * replica, the gallery fixes the skin rather than offering it as a choice.
 */
export const FoundationsSkin = "editorial" as const;

/** The batch whose subjects are named products wearing their own logos. */
export const FamilyShowroom = "showroom";

/**
 * The house style the showroom batch assumes, and the only light one.
 *
 * The split here is harder than replica's or foundations'. Those batches read as
 * a different production in the wrong skin; these templates do not work in it at
 * all. Their whole claim is that the logo on the card is the real one, and a real
 * logo on a near-black stage has to be recoloured to be visible — at which point
 * the viewer is no longer recognising anything.
 */
export const ShowroomSkin = "showroom" as const;

/** The house styles a clip can be cut in. Mirrors videoskin.go. */
export type Skin = "default" | "broadcast" | "minimal" | "editorial" | "showroom";

/**
 * Which families a theme may cast from. Mirrors comboPools in combo_pool.go.
 *
 * The whole piece is cut in one house style, so a combo's template picker has to
 * offer the same set the server will accept — otherwise the page shows a choice
 * that comes back as a validation error, which teaches people the picker lies.
 * The core family is in every pool: those templates predate the split and are
 * drawn skin-neutral, so they inherit whichever style they are rendered in.
 *
 * Duplicated from Go rather than fetched, and that is a real trade. It is four
 * lines that go stale silently if a family is added; the alternative is an
 * endpoint and a loading state on a control that must render instantly. If this
 * grows a third entry, fetch it.
 */
export const ComboPools: Record<Skin, string[]> = {
  default: [FamilyCore],
  minimal: [FamilyCore],
  broadcast: [FamilyCore, FamilyReplica],
  editorial: [FamilyCore, FamilyFoundations],
  showroom: [FamilyCore, FamilyShowroom],
};

/** The themes a combo can be cut in, in picker order. */
export const ComboSkins: { value: Skin; label: string; hint: string }[] = [
  { value: "default", label: "Default", hint: "The catalog's own look — the core templates" },
  { value: "broadcast", label: "Broadcast", hint: "Near-black stage, large type, standing chrome — adds the replica batch" },
  { value: "minimal", label: "Minimal", hint: "Flat, one accent, no furniture — the core templates" },
  { value: "editorial", label: "Editorial", hint: "Hard left axis for diagram-dense teaching — adds the foundations batch" },
  { value: "showroom", label: "Showroom", hint: "Light: paper, white cards, real product logos — adds the showroom batch" },
];
