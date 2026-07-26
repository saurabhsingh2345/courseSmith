import { tokens } from "../theme/tokens";
import { motion } from "../theme/motionTokens";
import { ShowcasePlayer } from "../components/ShowcasePlayer";
import { AdaptiveQuizDemo } from "../components/AdaptiveQuizDemo";

// ShowcasePage (workstream H) is the stakeholder "aha" surface: it renders the
// design system and motion language that drive every course — the tokens from
// workstreams B/E made visible and interactive.
//
// Scaffold status: token + motion showcase are live (driven by the real token
// values). The live lesson-video preview, diagram gallery, and BKT panel are
// placeholders pending @remotion/player and the tutor service (workstream D).

function Section({ title, subtitle, children }: { title: string; subtitle?: string; children: React.ReactNode }) {
  return (
    <section className="mb-10">
      <h2 className="text-ink-100 text-lg font-semibold">{title}</h2>
      {subtitle ? <p className="text-ink-500 mb-3 text-xs">{subtitle}</p> : null}
      {children}
    </section>
  );
}

function Swatch({ name, value }: { name: string; value: string }) {
  return (
    <div className="flex flex-col gap-1">
      <div className="h-14 w-full rounded-md border border-ink-700" style={{ backgroundColor: value }} />
      <div className="text-ink-300 font-mono text-[10px]">{name}</div>
      <div className="text-ink-500 font-mono text-[10px]">{value}</div>
    </div>
  );
}

function ColorTokens() {
  const c = tokens.colors;
  const entries: [string, string][] = [
    ["brand.light", c.brand.light],
    ["brand.dark", c.brand.dark],
    ["brand.saturated", c.brand.saturated],
    ["success", c.semantic.success],
    ["error", c.semantic.error],
    ["warning", c.semantic.warning],
    ["info", c.semantic.info],
  ];
  return (
    <div className="grid grid-cols-3 gap-4 sm:grid-cols-4 md:grid-cols-7">
      {entries.map(([name, value]) => (
        <Swatch key={name} name={name} value={value} />
      ))}
    </div>
  );
}

function MotionTokens() {
  const timing = motion.timing;
  const rows: [string, number][] = [
    ["fast", timing.fast],
    ["normal", timing.normal],
    ["slow", timing.slow],
    ["verySlow", timing.verySlow],
  ];
  return (
    <div className="flex flex-col gap-3">
      {rows.map(([name, seconds]) => (
        <div key={name} className="flex items-center gap-3">
          <div className="text-ink-300 w-20 font-mono text-xs">{name}</div>
          {/* An animated bar whose fill loops over exactly this duration, so the
              motion tokens are shown as felt time, not just numbers. */}
          <div className="bg-ink-800 h-3 flex-1 overflow-hidden rounded-full">
            <div
              className="showcase-bar h-full rounded-full"
              style={{
                background: "var(--color-brand)",
                animationDuration: `${seconds}s`,
                animationTimingFunction: motion.easing.subtle,
              }}
            />
          </div>
          <div className="text-ink-500 w-16 font-mono text-[10px]">{seconds}s</div>
        </div>
      ))}
      <style>{`
        @keyframes showcase-fill { from { width: 0% } to { width: 100% } }
        .showcase-bar { animation-name: showcase-fill; animation-iteration-count: infinite; animation-direction: alternate; }
      `}</style>
    </div>
  );
}

function EasingTokens() {
  const e = motion.easing;
  const rows: [string, string][] = [
    ["entrance", e.entrance],
    ["exit", e.exit],
    ["subtle", e.subtle],
  ];
  return (
    <div className="flex flex-col gap-3">
      {rows.map(([name, curve]) => (
        <div key={name} className="flex items-center gap-3">
          <div className="text-ink-300 w-20 font-mono text-xs">{name}</div>
          <div className="relative h-8 flex-1">
            <div
              className="showcase-dot absolute top-1/2 h-4 w-4 -translate-y-1/2 rounded-full"
              style={{ background: "var(--color-brand)", animationDuration: "1.6s", animationTimingFunction: curve }}
            />
          </div>
          <div className="text-ink-500 w-56 truncate font-mono text-[10px]">{curve}</div>
        </div>
      ))}
      <style>{`
        @keyframes showcase-slide { from { left: 0 } to { left: calc(100% - 1rem) } }
        .showcase-dot { animation-name: showcase-slide; animation-iteration-count: infinite; animation-direction: alternate; }
      `}</style>
    </div>
  );
}

export function ShowcasePage() {
  return (
    <div className="mx-auto max-w-4xl p-6">
      <h1 className="text-ink-100 mb-1 text-2xl font-bold">Design & motion showcase</h1>
      <p className="text-ink-500 mb-8 text-sm">
        The tokens and motion language every course is built from. Edit them once
        (Go owns motion; <span className="font-mono">tokens.ts</span> owns colour) and every lesson updates.
      </p>

      <Section title="Colour tokens" subtitle="Brand + semantic palette, theme-aware via CSS variables.">
        <ColorTokens />
      </Section>

      <Section title="Motion timing" subtitle="Duration tiers — watch each bar fill over exactly its token duration.">
        <MotionTokens />
      </Section>

      <Section title="Easing curves" subtitle="The three shared easings, animated.">
        <EasingTokens />
      </Section>

      <Section title="Lesson video preview" subtitle="The real renderer components, playing live via @remotion/player — the same code that renders final.mp4.">
        <div className="grid gap-4 sm:grid-cols-2">
          <ShowcasePlayer demo="exec" />
          <ShowcasePlayer demo="d3" />
        </div>
      </Section>

      <Section title="Adaptive insights" subtitle="Live BKT mastery + FSRS review schedule from the coursesmith-tutor service (workstream D).">
        <AdaptiveQuizDemo />
      </Section>
    </div>
  );
}
