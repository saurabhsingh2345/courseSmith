import { describe, expect, it } from "vitest";
import {
  formatBytes,
  formatInt,
  formatMs,
  formatPct,
  formatSec,
  formatUsd,
} from "./format";

describe("formatMs", () => {
  it("formats zero", () => {
    expect(formatMs(0)).toBe("0:00");
  });
  it("formats under a minute", () => {
    expect(formatMs(9_500)).toBe("0:09");
    expect(formatMs(59_999)).toBe("0:59");
  });
  it("formats minutes and pads seconds", () => {
    expect(formatMs(61_000)).toBe("1:01");
    expect(formatMs(175_568)).toBe("2:55");
  });
  it("rolls hours into minutes", () => {
    expect(formatMs(3_723_000)).toBe("62:03");
  });
  it("clamps negative and non-finite input to 0:00", () => {
    expect(formatMs(-5000)).toBe("0:00");
    expect(formatMs(NaN)).toBe("0:00");
    expect(formatMs(Infinity)).toBe("0:00");
  });
});

describe("formatSec", () => {
  it("converts seconds", () => {
    expect(formatSec(90)).toBe("1:30");
  });
});

describe("formatBytes", () => {
  it("keeps small values in bytes", () => {
    expect(formatBytes(0)).toBe("0 B");
    expect(formatBytes(512)).toBe("512 B");
    expect(formatBytes(1023)).toBe("1023 B");
  });
  it("steps through units", () => {
    expect(formatBytes(1024)).toBe("1 KB");
    expect(formatBytes(1536)).toBe("1.5 KB");
    expect(formatBytes(5 * 1024 * 1024)).toBe("5 MB");
    expect(formatBytes(3.2 * 1024 * 1024 * 1024)).toBe("3.2 GB");
  });
  it("drops decimals for 3-digit values", () => {
    expect(formatBytes(150 * 1024)).toBe("150 KB");
  });
  it("clamps negative and non-finite input", () => {
    expect(formatBytes(-10)).toBe("0 B");
    expect(formatBytes(NaN)).toBe("0 B");
  });
});

describe("formatUsd", () => {
  it("formats normal amounts to cents", () => {
    expect(formatUsd(1.234)).toBe("$1.23");
    expect(formatUsd(0)).toBe("$0.00");
    expect(formatUsd(1200)).toBe("$1200.00");
  });
  it("keeps 4 decimals for sub-dollar amounts", () => {
    expect(formatUsd(0.07167615)).toBe("$0.0717");
    expect(formatUsd(0.9999)).toBe("$0.9999");
  });
  it("handles negatives", () => {
    expect(formatUsd(-0.5)).toBe("-$0.5000");
    expect(formatUsd(-2)).toBe("-$2.00");
  });
  it("treats non-finite as zero", () => {
    expect(formatUsd(NaN)).toBe("$0.00");
  });
});

describe("formatPct", () => {
  it("rounds ratios", () => {
    expect(formatPct(0)).toBe("0%");
    expect(formatPct(0.126)).toBe("13%");
    expect(formatPct(1)).toBe("100%");
  });
  it("treats non-finite as zero", () => {
    expect(formatPct(NaN)).toBe("0%");
  });
});

describe("formatInt", () => {
  it("adds separators", () => {
    expect(formatInt(388665)).toBe("388,665");
    expect(formatInt(91)).toBe("91");
  });
});
