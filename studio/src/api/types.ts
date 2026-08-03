
// No-code pieces. The one thing that separates this surface from a reel is
// that a segment carries evidence, so that is the field the page is built
// around.
export interface NoCodeEvidenceInfo {
  kind: "capture" | "fact";
  tool?: string;
  display?: string;
  of?: string;
  take?: string;
  fixture?: string;
  facts?: string[];
  /** True once the recording actually exists on disk. */
  recorded: boolean;
  /** The moments the recorder stamped — what the narration may talk about. */
  marks?: string[];
  duration_ms?: number;
}

export interface NoCodeSegmentInfo {
  id: string;
  template: string;
  prompt: string;
  skip?: boolean;
  template_title: string;
  template_category: string;
  evidence: NoCodeEvidenceInfo;
}

export interface NoCodeSummary {
  id: string;
  title: string;
  brief?: string;
  segments: number;
  skipped: number;
  /** How many segments stand on a recording, and how many of those exist. */
  captures: number;
  recorded: number;
  ready: boolean;
  video_url?: string;
  /** Where the render lands on disk, relative to the project root. Set whether
   *  or not it exists yet — "it will be written here" is as useful as "it is". */
  video_path?: string;
  created_at?: string;
}

export interface NoCodeDetail extends NoCodeSummary {
  segment_list: NoCodeSegmentInfo[];
  /** Exactly what a run of this piece will do, in order, so progress can be
   *  shown against a real list instead of a spinner. */
  stages?: string[];
  plan?: unknown;
}

export interface NoCodeCatalogEntry {
  name: string;
  title: string;
  description: string;
  category: string;
  /** Cannot be filled from a prompt — its frame is a recording or nothing. */
  needs_capture: boolean;
}

export interface Recordable {
  key: string;
  display: string;
  kind: "terminal" | "web" | "desktop";
  /** A description cannot drive it; it needs a checked-in take. */
  needs_take: boolean;
}

/** One take file a capture segment can name. */
export interface AvailableTake {
  /** What goes in the field: the base name, no extension. */
  name: string;
  /** The site or app key it drives, read out of the file itself. */
  tool: string;
  kind: "web" | "desktop";
  steps: number;
}

export interface NoCodeTakes {
  /** Where takes live, relative to the project root. */
  dir: string;
  takes: AvailableTake[];
}

export interface CreateNoCodeRequest {
  title: string;
  brief: string;
  segments: {
    template: string;
    prompt: string;
    kind: string;
    tool?: string;
    of?: string;
    take?: string;
    facts?: string[];
  }[];
}
