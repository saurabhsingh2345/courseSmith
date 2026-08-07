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

/** The house styles a clip can be cut in. Mirrors videoskin.go. */
export type Skin = "default" | "broadcast" | "minimal" | "editorial";
