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
