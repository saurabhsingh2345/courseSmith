import { SnippetWorkbench } from "../components/SnippetWorkbench";
import { FamilyFoundations, FoundationsSkin } from "../lib/families";

// FoundationsPage is the third front door onto the snippet machinery, offering
// the batch built to teach computer science itself — binary, architecture,
// memory, operating systems, networking, algorithms, git.
//
// It is a separate page for the same two reasons replica is. The batch carries
// a skin assumption: these templates are diagram-dense and composed against the
// editorial skin's hard left axis, and dropped into a centred default-skin
// course they read as a different production. And the snippets catalog is
// already past thirty cards; thirty more on the same scroll would tax everyone
// who came for the ones that were already there.
//
// What makes this batch a family rather than thirty more looks: most of its
// templates draw pictures whose truth is checkable — the binary shown must
// equal the decimal claimed, the truth table must match the gate, the carry
// must actually carry — and their validators do the arithmetic.

export function FoundationsPage() {
  return (
    <SnippetWorkbench
      family={FamilyFoundations}
      skin={FoundationsSkin}
      heading="Make a foundations clip"
      blurb="The same one-prompt clip, drawn for teaching how computers actually work: bit rows that decode, packets that travel, stacks that breathe, commit graphs that diverge."
      galleryHeading="Which picture teaches it?"
    >
      <p className="mt-3 rounded-lg border border-ink-800 bg-ink-900 p-3 text-[13px] leading-relaxed text-ink-400">
        These are cut in the{" "}
        <span className="font-mono text-[12px] text-ink-300">editorial</span> skin — the headline
        holds the left edge and the diagram takes the rest of the frame. Where a template shows a
        computation, the pipeline checks the computation: a clip from this gallery does not ship a
        wrong sum, a wrong truth table, or a bit pattern that does not decode to what it claims.
      </p>
    </SnippetWorkbench>
  );
}
