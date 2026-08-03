import { SnippetWorkbench } from "../components/SnippetWorkbench";
import { FamilyReplica, ReplicaSkin } from "../lib/families";

// ReplicaPage is the second front door onto the snippet machinery, offering the
// batch drawn frame-by-frame against the reference explainer videos.
//
// It is a separate page rather than a category on the snippets gallery for two
// reasons, and neither is tidiness.
//
// The first is that these templates carry an assumption the others do not: they
// are composed for a near-black stage with a lot of air and standing chrome, and
// dropped into a default-skin course they read as a different production. The
// gallery fixes the skin, which is only honest if the gallery is also the thing
// you chose to open.
//
// The second is that the snippets catalog is already thirty-two cards. Adding
// twelve more to the same scroll makes the screen worse for everybody who came
// for the ones that were already there — the cost of a bigger catalog lands on
// people who never asked for it.

export function ReplicaPage() {
  return (
    <SnippetWorkbench
      family={FamilyReplica}
      skin={ReplicaSkin}
      heading="Make a replica"
      blurb="The same one-prompt clip, drawn in the reference explainer grammar: a near-black stage, one precise diagram, and a headline with a single word carrying the argument."
      galleryHeading="Which picture does the work?"
    >
      <p className="mt-3 rounded-lg border border-ink-800 bg-ink-900 p-3 text-[13px] leading-relaxed text-ink-400">
        These are cut in the{" "}
        <span className="font-mono text-[12px] text-ink-300">broadcast</span> skin and colour by
        meaning rather than by branding — gold is the quantity under discussion, red the ceiling it
        runs into, blue the alternative being weighed. Every clip from this gallery uses the same
        three, so the second one a viewer watches is already legible.
      </p>
    </SnippetWorkbench>
  );
}
