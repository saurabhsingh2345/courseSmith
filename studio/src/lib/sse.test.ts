import { describe, expect, it, vi } from "vitest";
import {
  backoffDelay,
  createSSEParser,
  SSEClient,
  type StudioEvent,
} from "./sse";

describe("createSSEParser", () => {
  it("parses a single complete message with id", () => {
    const p = createSSEParser();
    const msgs = p.push('id: 7\ndata: {"type":"log","seq":7}\n\n');
    expect(msgs).toEqual([{ id: "7", data: '{"type":"log","seq":7}' }]);
  });

  it("buffers partial chunks across pushes", () => {
    const p = createSSEParser();
    expect(p.push("id: 1\nda")).toEqual([]);
    expect(p.push("ta: hello\n")).toEqual([]);
    const msgs = p.push("\nid: 2\ndata: world\n\n");
    expect(msgs).toEqual([
      { id: "1", data: "hello" },
      { id: "2", data: "world" },
    ]);
  });

  it("ignores comment heartbeats", () => {
    const p = createSSEParser();
    expect(p.push(": ping\n\n")).toEqual([]);
    expect(p.push(": ping\n\ndata: x\n\n")).toEqual([{ data: "x" }]);
  });

  it("joins multi-line data with newlines", () => {
    const p = createSSEParser();
    const msgs = p.push("data: line1\ndata: line2\n\n");
    expect(msgs).toEqual([{ data: "line1\nline2" }]);
  });

  it("handles CRLF separators", () => {
    const p = createSSEParser();
    const msgs = p.push("id: 3\r\ndata: crlf\r\n\r\n");
    expect(msgs).toEqual([{ id: "3", data: "crlf" }]);
  });

  it("parses event field and value without leading space", () => {
    const p = createSSEParser();
    const msgs = p.push("event:log\ndata:x\n\n");
    expect(msgs).toEqual([{ event: "log", data: "x" }]);
  });

  it("returns multiple messages from one chunk", () => {
    const p = createSSEParser();
    const msgs = p.push("data: a\n\ndata: b\n\ndata: c\n\n");
    expect(msgs.map((m) => m.data)).toEqual(["a", "b", "c"]);
  });
});

describe("backoffDelay", () => {
  it("doubles from base", () => {
    expect(backoffDelay(0)).toBe(500);
    expect(backoffDelay(1)).toBe(1000);
    expect(backoffDelay(2)).toBe(2000);
    expect(backoffDelay(3)).toBe(4000);
  });
  it("caps at max", () => {
    expect(backoffDelay(10)).toBe(15000);
    expect(backoffDelay(100, 500, 8000)).toBe(8000);
  });
  it("clamps negative attempts to the base delay", () => {
    expect(backoffDelay(-3)).toBe(500);
  });
});

function sseResponse(body: string): Response {
  const stream = new ReadableStream<Uint8Array>({
    start(controller) {
      controller.enqueue(new TextEncoder().encode(body));
      controller.close();
    },
  });
  return new Response(stream, {
    status: 200,
    headers: { "Content-Type": "text/event-stream" },
  });
}

function deferred<T>() {
  let resolve!: (v: T) => void;
  const promise = new Promise<T>((r) => (resolve = r));
  return { promise, resolve };
}

describe("SSEClient", () => {
  it("dispatches typed events to type and wildcard handlers", async () => {
    const done = deferred<void>();
    const fetchFn = vi.fn(async () =>
      sseResponse(
        'id: 1\ndata: {"type":"log","line":"hi","seq":1,"at":"t"}\n\n' +
          'id: 2\ndata: {"type":"run-finished","seq":2,"at":"t"}\n\n',
      ),
    );
    const client = new SSEClient("/api/events", {
      fetchFn: fetchFn as unknown as typeof fetch,
      // Swallow reconnects after the stream ends.
      setTimeoutFn: () => 0 as unknown as ReturnType<typeof setTimeout>,
    });
    const logs: StudioEvent[] = [];
    const all: StudioEvent[] = [];
    client.on("log", (e) => logs.push(e));
    client.on("*", (e) => all.push(e));
    client.on("run-finished", () => done.resolve());
    client.start();
    await done.promise;
    client.close();

    expect(logs).toHaveLength(1);
    expect(logs[0].line).toBe("hi");
    expect(all.map((e) => e.type)).toEqual(["log", "run-finished"]);
    expect(client.lastEventId).toBe("2");
  });

  it("sends Last-Event-ID on reconnect and backs off exponentially", async () => {
    const secondConnect = deferred<void>();
    const delays: number[] = [];
    let call = 0;
    const fetchFn = vi.fn(async (_url: unknown, init?: RequestInit) => {
      call += 1;
      if (call === 1) {
        return sseResponse('id: 42\ndata: {"type":"log","seq":42,"at":"t"}\n\n');
      }
      const headers = init?.headers as Record<string, string>;
      expect(headers["Last-Event-ID"]).toBe("42");
      secondConnect.resolve();
      // Hang forever so no further reconnects happen.
      return new Response(new ReadableStream<Uint8Array>({ start() {} }), {
        status: 200,
      });
    });
    const client = new SSEClient("/api/events", {
      fetchFn: fetchFn as unknown as typeof fetch,
      baseDelayMs: 100,
      maxDelayMs: 800,
      setTimeoutFn: (fn, ms) => {
        delays.push(ms);
        fn();
        return 0 as unknown as ReturnType<typeof setTimeout>;
      },
    });
    client.start();
    await secondConnect.promise;
    client.close();

    expect(fetchFn).toHaveBeenCalledTimes(2);
    expect(delays[0]).toBe(100);
  });

  it("retries with growing delays on repeated failure and resets on success", async () => {
    const delays: number[] = [];
    const done = deferred<void>();
    let call = 0;
    const fetchFn = vi.fn(async () => {
      call += 1;
      if (call <= 3) throw new Error("connection refused");
      done.resolve();
      return new Response(new ReadableStream<Uint8Array>({ start() {} }), {
        status: 200,
      });
    });
    const client = new SSEClient("/api/events", {
      fetchFn: fetchFn as unknown as typeof fetch,
      baseDelayMs: 100,
      maxDelayMs: 10_000,
      setTimeoutFn: (fn, ms) => {
        delays.push(ms);
        fn();
        return 0 as unknown as ReturnType<typeof setTimeout>;
      },
    });
    client.start();
    await done.promise;
    client.close();

    expect(delays).toEqual([100, 200, 400]);
  });

  it("does not reconnect after close()", async () => {
    const opened = deferred<void>();
    const fetchFn = vi.fn(async () => {
      opened.resolve();
      return sseResponse("data: x\n\n");
    });
    let scheduled = 0;
    const client = new SSEClient("/api/events", {
      fetchFn: fetchFn as unknown as typeof fetch,
      setTimeoutFn: (fn, _ms) => {
        scheduled += 1;
        fn();
        return 0 as unknown as ReturnType<typeof setTimeout>;
      },
    });
    client.start();
    await opened.promise;
    client.close();
    const after = scheduled;
    await new Promise((r) => setTimeout(r, 10));
    expect(fetchFn.mock.calls.length).toBeLessThanOrEqual(after + 1);
    expect(client.status).toBe("closed");
  });

  it("reports status transitions", async () => {
    const done = deferred<void>();
    const fetchFn = vi.fn(async () => sseResponse("data: x\n\n"));
    const statuses: string[] = [];
    const client = new SSEClient("/api/events", {
      fetchFn: fetchFn as unknown as typeof fetch,
      setTimeoutFn: () => {
        done.resolve();
        return 0 as unknown as ReturnType<typeof setTimeout>;
      },
    });
    client.onStatus((s) => statuses.push(s));
    client.start();
    await done.promise;
    client.close();
    expect(statuses[0]).toBe("connecting");
    expect(statuses).toContain("open");
    expect(statuses[statuses.length - 1]).toBe("closed");
  });
});
