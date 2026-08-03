import { SnippetWorkbench } from "../components/SnippetWorkbench";
import { FamilyCore } from "../lib/families";

// SnippetsPage is the short-form front door: write a prompt, pick how it should
// look, get a finished clip.
//
// It is deliberately not the Compose page with different labels. Composing a
// lesson is a decision about a course; making a snippet is a decision about a
// *look*, so the template gallery is the primary control and the prompt sits
// inside it. Everything else — the run log, the stage progress — is borrowed
// from the machinery the lesson path already has.
//
// The screen itself is SnippetWorkbench, shared with the replica gallery. This
// page is the copy and the family; nothing here decides behaviour.

export function SnippetsPage() {
  return (
    <SnippetWorkbench
      family={FamilyCore}
      heading="Make a snippet"
      blurb="One prompt, one look, one finished clip. No lesson to write and no course to pick."
    />
  );
}
