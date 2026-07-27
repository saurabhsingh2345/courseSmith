// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { SnippetsPage } from "./SnippetsPage";
import { RunProvider } from "../state/RunContext";
import { ShortcutProvider } from "../state/ShortcutContext";
import {
  api,
  type CreateSnippetResponse,
  type SnippetSummary,
  type SnippetTemplateInfo,
} from "../api/client";

const TEMPLATES: SnippetTemplateInfo[] = [
  {
    name: "vscode",
    title: "VS Code walkthrough",
    description: "An editor opens, code types itself in, and the terminal runs it for real.",
    example: "How for loops work in Python",
    shows_code: true,
    min_target_sec: 0,
    default_target_sec: 0,
  },
  {
    name: "whiteboard",
    title: "Whiteboard sketch",
    description: "A hand-drawn board that fills in as you talk.",
    example: "Why HTTP caching matters",
    shows_code: false,
    min_target_sec: 0,
    default_target_sec: 0,
  },
];

const SNIPPETS: SnippetSummary[] = [
  {
    id: "for-loops-in-python",
    title: "For Loops in Python",
    prompt: "how for loops work",
    template: "vscode",
    ready: true,
    video_url: "/artifacts/snippets/for-loops-in-python/final.mp4",
  },
];

function renderPage() {
  return render(
    <ShortcutProvider>
      <RunProvider>
        <SnippetsPage />
      </RunProvider>
    </ShortcutProvider>,
  );
}

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

function stubBaseCalls() {
  vi.spyOn(api, "snippetTemplates").mockResolvedValue(TEMPLATES);
  vi.spyOn(api, "snippets").mockResolvedValue(SNIPPETS);
  vi.spyOn(api, "runStatus").mockResolvedValue({ running: false });
}

it("renders the template gallery and existing snippets", async () => {
  stubBaseCalls();
  renderPage();

  expect(await screen.findByText("VS Code walkthrough")).toBeTruthy();
  expect(screen.getByText("runs code")).toBeTruthy();
  expect(await screen.findByText("For Loops in Python")).toBeTruthy();
  expect(screen.getByText("ready")).toBeTruthy();
});

// The primary action must not be blocked on a choice the user has no opinion
// about, so the first template is selected by default.
it("submits with the first template preselected", async () => {
  stubBaseCalls();
  const created: CreateSnippetResponse = {
    id: "new-clip",
    title: "New clip",
    prompt: "explain dictionaries",
    template: "vscode",
    ready: false,
    run_id: "run-1",
  };
  const create = vi.spyOn(api, "createSnippet").mockResolvedValue(created);
  vi.spyOn(api, "snippet").mockResolvedValue({ ...created, target_sec: 45 });

  renderPage();
  await screen.findByText("VS Code walkthrough");

  fireEvent.change(screen.getByLabelText("What should it teach?"), {
    target: { value: "explain dictionaries" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Generate clip" }));

  await waitFor(() => expect(create).toHaveBeenCalled());
  expect(create.mock.calls[0][0]).toMatchObject({
    prompt: "explain dictionaries",
    template: "vscode",
    target_sec: 45,
  });
});

it("keeps the generate button disabled until there is a prompt", async () => {
  stubBaseCalls();
  renderPage();
  await screen.findByText("VS Code walkthrough");

  const button = screen.getByRole("button", { name: "Generate clip" }) as HTMLButtonElement;
  expect(button.disabled).toBe(true);

  fireEvent.change(screen.getByLabelText("What should it teach?"), {
    target: { value: "loops" },
  });
  expect(button.disabled).toBe(false);
});

it("plan only asks the backend to stop after planning", async () => {
  stubBaseCalls();
  const create = vi.spyOn(api, "createSnippet").mockResolvedValue({
    id: "planned",
    title: "Planned",
    prompt: "loops",
    template: "vscode",
    ready: false,
  });
  vi.spyOn(api, "snippet").mockResolvedValue({
    id: "planned",
    title: "Planned",
    prompt: "loops",
    template: "vscode",
    ready: false,
    target_sec: 45,
  });

  renderPage();
  await screen.findByText("VS Code walkthrough");
  fireEvent.change(screen.getByLabelText("What should it teach?"), {
    target: { value: "loops" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Plan only" }));

  await waitFor(() => expect(create).toHaveBeenCalled());
  expect(create.mock.calls[0][0].plan_only).toBe(true);
});

// Picking a template is the moment someone has just been told what it does and
// has to invent a prompt that suits it. The example is already that prompt.
it("fills the prompt with a template's example when it is picked", async () => {
  stubBaseCalls();
  renderPage();
  await screen.findByText("Whiteboard sketch");

  const box = screen.getByLabelText("What should it teach?") as HTMLTextAreaElement;
  expect(box.value).toBe("");

  fireEvent.click(screen.getByRole("button", { name: /Whiteboard sketch/ }));
  expect(box.value).toBe("Why HTTP caching matters");

  // Browsing between templates keeps swapping the demo, because what is in the
  // box is still an example rather than anything the user wrote.
  fireEvent.click(screen.getByRole("button", { name: /VS Code walkthrough/ }));
  expect(box.value).toBe("How for loops work in Python");
});

// One typed character makes the box theirs. Switching templates after that must
// not throw the prompt away.
it("never overwrites a prompt the user typed", async () => {
  stubBaseCalls();
  renderPage();
  await screen.findByText("Whiteboard sketch");

  const box = screen.getByLabelText("What should it teach?") as HTMLTextAreaElement;
  fireEvent.change(box, { target: { value: "my own idea about closures" } });
  fireEvent.click(screen.getByRole("button", { name: /Whiteboard sketch/ }));

  expect(box.value).toBe("my own idea about closures");
});

// The title is optional, and an empty one must not be sent as "" — the pipeline
// treats an empty title as "let the model write one", and so should the form.
it("sends an edited title, and omits an empty one", async () => {
  stubBaseCalls();
  const created: CreateSnippetResponse = {
    id: "titled",
    title: "My Own Title",
    prompt: "explain closures",
    template: "vscode",
    ready: false,
  };
  const create = vi.spyOn(api, "createSnippet").mockResolvedValue(created);
  vi.spyOn(api, "snippet").mockResolvedValue({ ...created, target_sec: 45 });

  renderPage();
  await screen.findByText("VS Code walkthrough");

  fireEvent.change(screen.getByLabelText("What should it teach?"), {
    target: { value: "explain closures" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Generate clip" }));
  await waitFor(() => expect(create).toHaveBeenCalled());
  expect(create.mock.calls[0][0].title).toBeUndefined();

  fireEvent.change(screen.getByLabelText(/^Title/), {
    target: { value: "  My Own Title  " },
  });
  fireEvent.change(screen.getByLabelText("What should it teach?"), {
    target: { value: "explain closures" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Generate clip" }));
  await waitFor(() => expect(create).toHaveBeenCalledTimes(2));
  expect(create.mock.calls[1][0].title).toBe("My Own Title");
});

// A template that cannot be short must not be offered a short runtime. Offering
// one cost a real user three correction rounds and a day's token budget to find
// out that eight beats do not fit in twenty seconds.
it("only offers runtimes the chosen template can satisfy", async () => {
  vi.spyOn(api, "snippetTemplates").mockResolvedValue([
    ...TEMPLATES,
    {
      name: "story",
      title: "Directed short",
      description: "A one-to-two-minute piece.",
      example: "How a database index finds your row",
      shows_code: false,
      min_target_sec: 60,
      default_target_sec: 90,
    },
  ]);
  vi.spyOn(api, "snippets").mockResolvedValue(SNIPPETS);
  vi.spyOn(api, "runStatus").mockResolvedValue({ running: false });

  renderPage();
  await screen.findByText("Directed short");

  // The short runtimes are on offer for the default template...
  expect(screen.queryByRole("button", { name: "~20s" })).toBeTruthy();

  fireEvent.click(screen.getByRole("button", { name: /Directed short/ }));

  // ...and gone once a template that cannot be short is chosen.
  expect(screen.queryByRole("button", { name: "~20s" })).toBeNull();
  expect(screen.queryByRole("button", { name: "~45s" })).toBeNull();
  expect(screen.queryByRole("button", { name: "~75s" })).toBeTruthy();
  expect(screen.queryByRole("button", { name: "~2 min" })).toBeTruthy();
});
