/**
 * SSE client for /api/events with Last-Event-ID resume and exponential
 * backoff reconnect.
 *
 * Native EventSource cannot send a custom Last-Event-ID header after we
 * recreate it for backoff, so this uses fetch + a small incremental SSE
 * parser. The parser and backoff schedule are pure and unit tested.
 */

export type StudioEventType =
  | "run-started"
  | "stage-started"
  | "stage-finished"
  | "stage-skipped"
  | "stage-failed"
  | "log"
  | "run-finished"
  | "run-failed"
  | "quota";

export interface StudioEvent {
  type: StudioEventType;
  run_id?: string;
  course?: string;
  lesson?: string;
  stage?: string;
  line?: string;
  error?: string;
  seq: number;
  at: string;
}

export interface SSEMessage {
  id?: string;
  event?: string;
  data: string;
}

/** Incremental server-sent-events wire parser. Feed chunks, get messages. */
export function createSSEParser(): { push(chunk: string): SSEMessage[] } {
  let buffer = "";
  return {
    push(chunk: string): SSEMessage[] {
      buffer += chunk;
      const messages: SSEMessage[] = [];
      // Messages are separated by a blank line. Normalize CRLF.
      for (;;) {
        const idx = buffer.search(/\r\n\r\n|\n\n|\r\r/);
        if (idx === -1) break;
        const sepLen = buffer.slice(idx, idx + 4) === "\r\n\r\n" ? 4 : 2;
        const raw = buffer.slice(0, idx);
        buffer = buffer.slice(idx + sepLen);
        const msg: SSEMessage = { data: "" };
        const dataLines: string[] = [];
        for (const line of raw.split(/\r\n|\n|\r/)) {
          if (line === "" || line.startsWith(":")) continue; // comment/heartbeat
          const colon = line.indexOf(":");
          const field = colon === -1 ? line : line.slice(0, colon);
          let value = colon === -1 ? "" : line.slice(colon + 1);
          if (value.startsWith(" ")) value = value.slice(1);
          if (field === "data") dataLines.push(value);
          else if (field === "id") msg.id = value;
          else if (field === "event") msg.event = value;
        }
        msg.data = dataLines.join("\n");
        if (msg.data !== "" || msg.id !== undefined || msg.event !== undefined) {
          messages.push(msg);
        }
      }
      return messages;
    },
  };
}

/** Exponential backoff with cap: base * 2^attempt, clamped to max. */
export function backoffDelay(attempt: number, base = 500, max = 15000): number {
  if (attempt < 0) attempt = 0;
  const delay = base * Math.pow(2, attempt);
  return Math.min(delay, max);
}

export type SSEStatus = "connecting" | "open" | "reconnecting" | "closed";

type EventHandler = (event: StudioEvent) => void;
type StatusHandler = (status: SSEStatus) => void;

export interface SSEClientOptions {
  /** Injectable for tests. Defaults to global fetch. */
  fetchFn?: typeof fetch;
  baseDelayMs?: number;
  maxDelayMs?: number;
  /** Injectable timer for tests. Defaults to setTimeout. */
  setTimeoutFn?: (fn: () => void, ms: number) => ReturnType<typeof setTimeout>;
}

export class SSEClient {
  private url: string;
  private fetchFn: typeof fetch;
  private baseDelayMs: number;
  private maxDelayMs: number;
  private setTimeoutFn: (fn: () => void, ms: number) => ReturnType<typeof setTimeout>;

  private handlers = new Map<string, Set<EventHandler>>();
  private statusHandlers = new Set<StatusHandler>();
  private abort: AbortController | null = null;
  private timer: ReturnType<typeof setTimeout> | null = null;
  private attempt = 0;
  private stopped = true;

  lastEventId: string | null = null;
  status: SSEStatus = "closed";

  constructor(url: string, opts: SSEClientOptions = {}) {
    this.url = url;
    this.fetchFn = opts.fetchFn ?? ((...args) => fetch(...args));
    this.baseDelayMs = opts.baseDelayMs ?? 500;
    this.maxDelayMs = opts.maxDelayMs ?? 15000;
    this.setTimeoutFn = opts.setTimeoutFn ?? ((fn, ms) => setTimeout(fn, ms));
  }

  /** Subscribe to one event type, or "*" for all. Returns unsubscribe. */
  on(type: StudioEventType | "*", handler: EventHandler): () => void {
    let set = this.handlers.get(type);
    if (!set) {
      set = new Set();
      this.handlers.set(type, set);
    }
    set.add(handler);
    return () => set.delete(handler);
  }

  onStatus(handler: StatusHandler): () => void {
    this.statusHandlers.add(handler);
    return () => this.statusHandlers.delete(handler);
  }

  start(): void {
    if (!this.stopped) return;
    this.stopped = false;
    this.attempt = 0;
    void this.connect();
  }

  close(): void {
    this.stopped = true;
    if (this.timer !== null) {
      clearTimeout(this.timer);
      this.timer = null;
    }
    this.abort?.abort();
    this.abort = null;
    this.setStatus("closed");
  }

  private setStatus(status: SSEStatus): void {
    this.status = status;
    for (const h of this.statusHandlers) h(status);
  }

  private dispatch(event: StudioEvent): void {
    for (const h of this.handlers.get(event.type) ?? []) h(event);
    for (const h of this.handlers.get("*") ?? []) h(event);
  }

  private async connect(): Promise<void> {
    if (this.stopped) return;
    this.setStatus(this.attempt === 0 ? "connecting" : "reconnecting");
    this.abort = new AbortController();
    const headers: Record<string, string> = { Accept: "text/event-stream" };
    if (this.lastEventId !== null) headers["Last-Event-ID"] = this.lastEventId;
    try {
      const res = await this.fetchFn(this.url, {
        headers,
        signal: this.abort.signal,
      });
      if (!res.ok || !res.body) throw new Error(`SSE HTTP ${res.status}`);
      this.setStatus("open");
      this.attempt = 0;
      const parser = createSSEParser();
      const reader = res.body.getReader();
      const decoder = new TextDecoder();
      for (;;) {
        const { done, value } = await reader.read();
        if (done) break;
        for (const msg of parser.push(decoder.decode(value, { stream: true }))) {
          if (msg.id !== undefined) this.lastEventId = msg.id;
          if (msg.data === "") continue;
          let parsed: StudioEvent;
          try {
            parsed = JSON.parse(msg.data) as StudioEvent;
          } catch {
            continue;
          }
          this.dispatch(parsed);
        }
      }
      // Stream ended cleanly: server closed; reconnect.
      this.scheduleReconnect();
    } catch {
      if (this.stopped) return;
      this.scheduleReconnect();
    }
  }

  private scheduleReconnect(): void {
    if (this.stopped) return;
    const delay = backoffDelay(this.attempt, this.baseDelayMs, this.maxDelayMs);
    this.attempt += 1;
    this.setStatus("reconnecting");
    this.timer = this.setTimeoutFn(() => {
      this.timer = null;
      void this.connect();
    }, delay);
  }
}
