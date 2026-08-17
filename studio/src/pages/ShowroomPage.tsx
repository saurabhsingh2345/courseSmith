import { SnippetWorkbench } from "../components/SnippetWorkbench";
import { FamilyShowroom, ShowroomSkin } from "../lib/families";

// ShowroomPage is the front door onto the batch whose subjects are real products.
//
// Every other gallery here offers templates that DRAW their subject — a bar, a
// ring, a boundary, a table — which is why they can all be themed and timed the
// same way, and why none of them can put something on screen that a viewer
// RECOGNISES. A row of products set in the house font is a row of words.
//
// These three fetch the real logo and keep the real colour, and that is why they
// are a separate page rather than three more cards on the snippets scroll: they
// are the only templates in the catalog cut on paper. The skin is not offered as a
// choice because there is no version of these clips on a dark stage — the mark
// would have to be repainted to be visible, and a repainted mark is not the thing
// anybody recognises.

export function ShowroomPage() {
  return (
    <SnippetWorkbench
      family={FamilyShowroom}
      skin={ShowroomSkin}
      heading="Make a showroom clip"
      blurb="The same one-prompt clip, about things that exist by name: white cards on paper, each wearing the product's own logo in the product's own colour."
      galleryHeading="How many things, and what are you claiming?"
    >
      <p className="mt-3 rounded-lg border border-ink-800 bg-ink-900 p-3 text-[13px] leading-relaxed text-ink-400">
        These are cut in the{" "}
        <span className="font-mono text-[12px] text-ink-300">showroom</span> skin — the only light one
        — and they are the only templates that reach the open web while planning: the logo is fetched
        from Simple Icons and the brand's own colour comes with it. Name the products in your prompt
        and give their real names; anything with no logo to be found gets a drawn glyph instead, which
        is a plainer card rather than a hole.
      </p>
    </SnippetWorkbench>
  );
}
