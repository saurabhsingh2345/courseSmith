// Every register the film cuts between, other than the map.
//
// Three of these are the film's actual arguments rather than decoration: the
// decimal army, the rider's kit and the yam relay are the reasons the conquest
// worked, and each is a thing a diagram can show and footage cannot.

import React from 'react';
import {
  AbsoluteFill,
  OffthreadVideo,
  Img,
  staticFile,
  interpolate,
  useCurrentFrame,
  Easing,
  random,
} from 'remotion';
import {
  C,
  DISPLAY,
  NUM,
  BODY,
  Display,
  Body,
  Eyebrow,
  Rule,
  Glyph,
  Vignette,
  useWeave,
  ramp,
  holdFade,
} from './kit';
import {LAND_D, SPHERE_D, routeD, ROUTE, ROUTE_TOTAL_KM, projection, MAP_W, MAP_H} from './geo';

/** The grade that makes 1926 and 1928 stock read as one film. */
const GRADE =
  'sepia(0.58) saturate(1.22) contrast(1.09) brightness(0.9)';

// --------------------------------------------------------------------- footage

export const FootageScene: React.FC<{
  clip: string;
  fit: 'bleed' | 'plate';
  push?: number;
  credit?: string;
  total: number;
}> = ({clip, fit, push = 0.06, credit, total}) => {
  const frame = useCurrentFrame();
  const weave = useWeave(fit === 'bleed' ? 2.4 : 1.6);
  const fade = holdFade(frame, total, 18, 20);
  // A slow push over the whole shot. A static archival frame held for twelve
  // seconds reads as a still; the same frame with a 6% push reads as picture.
  const zoom = 1 + push * (frame / Math.max(1, total));

  const video = (
    <OffthreadVideo
      src={staticFile(`genghis/clips/${clip}.mp4`)}
      muted
      // Every clip is cut a little longer than any beat needs it, so a beat
      // that runs long holds real picture instead of freezing on black.
      style={{
        width: '100%',
        height: '100%',
        objectFit: 'cover',
        filter: GRADE,
        transform: `scale(${zoom}) translate(${weave.x}px, ${weave.y}px)`,
      }}
    />
  );

  if (fit === 'bleed') {
    return (
      <AbsoluteFill style={{background: C.ink0, opacity: fade}}>
        <AbsoluteFill>{video}</AbsoluteFill>
        <Vignette strength={0.82} />
        {credit ? <Credit>{credit}</Credit> : null}
      </AbsoluteFill>
    );
  }

  return (
    <AbsoluteFill style={{background: C.ink0, opacity: fade}}>
      <AbsoluteFill
        style={{
          background:
            'radial-gradient(58% 52% at 50% 46%, rgba(200,160,85,0.07) 0%, rgba(0,0,0,0) 70%)',
        }}
      />
      <AbsoluteFill style={{alignItems: 'center', justifyContent: 'center'}}>
        <div
          style={{
            width: 1160,
            height: 840,
            marginTop: -26,
            overflow: 'hidden',
            border: `1px solid ${C.line}`,
            boxShadow: `0 0 0 1px rgba(200,160,85,0.16), 0 40px 90px rgba(0,0,0,0.6)`,
            position: 'relative',
          }}
        >
          {video}
          <AbsoluteFill style={{boxShadow: 'inset 0 0 120px rgba(0,0,0,0.55)'}} />
        </div>
      </AbsoluteFill>
      {credit ? <Credit>{credit}</Credit> : null}
    </AbsoluteFill>
  );
};

const Credit: React.FC<{children: React.ReactNode}> = ({children}) => {
  const frame = useCurrentFrame();
  return (
    <div
      style={{
        position: 'absolute',
        left: 96,
        bottom: 58,
        opacity: 0.72 * ramp(frame, 20, 26),
      }}
    >
      <Eyebrow size={13} color={C.muted}>
        {children}
      </Eyebrow>
    </div>
  );
};

// ------------------------------------------------------------------- art plate

const ART: Record<string, string> = {
  genghis: 'genghis-portrait.jpg',
  beijing: 'siege-of-zhongdu.jpeg',
  indus: 'indus.jpg',
  caravan: 'caravan.jpg',
  camp: 'camp.jpg',
  coronation: 'coronation.jpg',
};

export const PlateScene: React.FC<{
  art: string;
  focus?: [number, number];
  push?: number;
  credit?: string;
  total: number;
}> = ({art, focus = [0.5, 0.5], push = 0.07, credit, total}) => {
  const frame = useCurrentFrame();
  const fade = holdFade(frame, total, 22, 22);
  const zoom = 1.02 + push * (frame / Math.max(1, total));
  return (
    <AbsoluteFill style={{background: C.ink0, opacity: fade}}>
      <AbsoluteFill
        style={{
          background:
            'radial-gradient(60% 55% at 50% 46%, rgba(200,160,85,0.09) 0%, rgba(0,0,0,0) 72%)',
        }}
      />
      <AbsoluteFill style={{alignItems: 'center', justifyContent: 'center'}}>
        <div
          style={{
            width: 1000,
            height: 830,
            marginTop: -24,
            overflow: 'hidden',
            border: `1px solid rgba(200,160,85,0.30)`,
            boxShadow: '0 40px 100px rgba(0,0,0,0.65)',
            position: 'relative',
            background: C.ink1,
          }}
        >
          <Img
            src={staticFile(`genghis/art/${ART[art] ?? art}`)}
            style={{
              width: '100%',
              height: '100%',
              objectFit: 'cover',
              objectPosition: `${focus[0] * 100}% ${focus[1] * 100}%`,
              transform: `scale(${zoom})`,
              filter: 'contrast(1.04) saturate(1.04) brightness(0.97)',
            }}
          />
          <AbsoluteFill style={{boxShadow: 'inset 0 0 130px rgba(0,0,0,0.5)'}} />
        </div>
      </AbsoluteFill>
      {credit ? <Credit>{credit}</Credit> : null}
    </AbsoluteFill>
  );
};

// ----------------------------------------------------------------------- title

export const TitleScene: React.FC<{total: number}> = ({total}) => {
  const frame = useCurrentFrame();
  const a = ramp(frame, 4, 40);
  const b = ramp(frame, 26, 46);
  const c = ramp(frame, 52, 40);
  const out = 1 - ramp(frame, total - 22, 22, Easing.linear);
  return (
    <AbsoluteFill
      style={{
        background: C.ink0,
        alignItems: 'center',
        justifyContent: 'center',
        opacity: out,
      }}
    >
      <AbsoluteFill
        style={{
          background:
            'radial-gradient(70% 60% at 50% 42%, rgba(200,160,85,0.10) 0%, rgba(0,0,0,0) 70%)',
          opacity: a,
        }}
      />
      <div style={{alignItems: 'center', display: 'flex', flexDirection: 'column'}}>
        <div style={{opacity: a, transform: `translateY(${(1 - a) * 14}px)`}}>
          <Glyph name="seal" size={46} color={C.gold} weight={1.1} />
        </div>
        <div style={{height: 34}} />
        <div style={{opacity: a, transform: `translateY(${(1 - a) * 22}px)`}}>
          <Display size={158} weight={300} style={{letterSpacing: 2, textAlign: 'center'}}>
            Genghis Khan
          </Display>
        </div>
        <div style={{height: 30, opacity: b}} />
        <div style={{opacity: b}}>
          <Rule progress={b} width={520} color={C.gold} />
        </div>
        <div style={{height: 28}} />
        <div style={{opacity: b}}>
          <Eyebrow size={26} color={C.paper}>
            and the Mongol Empire
          </Eyebrow>
        </div>
        <div style={{height: 46}} />
        <div style={{opacity: c}}>
          <Eyebrow size={16} color={C.gold}>
            1162 — 1227
          </Eyebrow>
        </div>
      </div>
    </AbsoluteFill>
  );
};

// ----------------------------------------------------------------------- quote

export const QuoteScene: React.FC<{line: string; sub?: string; total: number}> = ({
  line,
  sub,
  total,
}) => {
  const frame = useCurrentFrame();
  const a = ramp(frame, 6, 34);
  const fade = holdFade(frame, total, 16, 20);
  return (
    <AbsoluteFill
      style={{
        background: C.ink0,
        alignItems: 'center',
        justifyContent: 'center',
        opacity: fade,
      }}
    >
      <DiagramGround glow={0.20} />
      <div style={{alignItems: 'center', display: 'flex', flexDirection: 'column'}}>
        <Rule progress={a} width={300} color={C.gold} />
        <div style={{height: 42}} />
        <div style={{opacity: a, transform: `translateY(${(1 - a) * 12}px)`}}>
          <Display size={126} weight={300} style={{textAlign: 'center'}}>
            {line}
          </Display>
        </div>
        {sub ? (
          <>
            <div style={{height: 40}} />
            <div style={{opacity: ramp(frame, 30, 34)}}>
              <Eyebrow size={17} color={C.gold}>
                {sub}
              </Eyebrow>
            </div>
          </>
        ) : null}
        <div style={{height: 42}} />
        <Rule progress={ramp(frame, 16, 34)} width={300} color={C.gold} />
      </div>
    </AbsoluteFill>
  );
};


/**
 * The ground under the four drawn scenes.
 *
 * Measured on the first render: the diagram scenes sat at 4-5% mean luminance
 * while the archival plates ran at 28-54%, so every cut into a diagram read as
 * the picture going out rather than as a change of register. blackdetect
 * flagged 27 consecutive seconds of the film as black. This lifts the ground
 * with a warm gradient, a glow behind the content and one very faint drawn arc
 * — enough structure that the frame is a surface instead of a void, without
 * putting anything on it that competes with the diagram.
 */
const DiagramGround: React.FC<{glow?: number}> = ({glow = 0.2}) => (
  <>
    <AbsoluteFill
      style={{
        background:
          'linear-gradient(176deg, #241B0D 0%, #17130C 44%, #0C0A08 100%)',
      }}
    />
    <AbsoluteFill
      style={{
        background: `radial-gradient(76% 64% at 50% 40%, rgba(214,172,92,${glow}) 0%, rgba(0,0,0,0) 74%)`,
      }}
    />
    <svg
      width={MAP_W}
      height={MAP_H}
      style={{position: 'absolute', inset: 0}}
      aria-hidden
    >
      <circle
        cx={960}
        cy={470}
        r={742}
        fill="none"
        stroke={C.gold}
        strokeOpacity={0.095}
        strokeWidth={1}
      />
      <circle
        cx={960}
        cy={470}
        r={498}
        fill="none"
        stroke={C.gold}
        strokeOpacity={0.06}
        strokeWidth={1}
      />
    </svg>
  </>
);

// ---------------------------------------------------------------------- tribes

/** Scattered clans pulled into one command, then the rule that did it. */
export const TribesScene: React.FC<{total: number}> = ({total}) => {
  const frame = useCurrentFrame();
  const fade = holdFade(frame, total, 18, 20);
  const N = 34;
  // Three movements: scattered, converging, one. The scatter holds through the
  // first third — that is the half of the line about dozens of clans.
  const pull = interpolate(frame, [total * 0.26, total * 0.66], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
    easing: Easing.bezier(0.4, 0, 0.2, 1),
  });
  const cx = 960;
  const cy = 400;
  const RING = 258;

  return (
    <AbsoluteFill style={{background: C.ink0, opacity: fade}}>
      <DiagramGround glow={0.20} />
      <svg width={MAP_W} height={MAP_H} style={{position: 'absolute', inset: 0}}>
        {new Array(N).fill(0).map((_, i) => {
          const sx = 250 + random(`x${i}`) * 1420;
          const sy = 150 + random(`y${i}`) * 560;
          const ang = (i / N) * Math.PI * 2;
          const tx = cx + Math.cos(ang) * RING;
          const ty = cy + Math.sin(ang) * RING;
          const x = sx + (tx - sx) * pull;
          const y = sy + (ty - sy) * pull;
          const r = 6.6 - 1.4 * pull;
          return (
            <g key={i}>
              {pull > 0.55 ? (
                <line
                  x1={x}
                  y1={y}
                  x2={cx}
                  y2={cy}
                  stroke={C.gold}
                  strokeOpacity={0.09 * (pull - 0.55) * 2.2}
                  strokeWidth={0.8}
                />
              ) : null}
              <circle cx={x} cy={y} r={r} fill={C.paper} fillOpacity={0.82} />
            </g>
          );
        })}
        <circle
          cx={cx}
          cy={cy}
          r={RING}
          fill="none"
          stroke={C.gold}
          strokeOpacity={0.45 * Math.max(0, (pull - 0.5) * 2)}
          strokeWidth={1.4}
        />
        <circle
          cx={cx}
          cy={cy}
          r={20 * Math.max(0, (pull - 0.72) / 0.28)}
          fill={C.goldBright}
          fillOpacity={0.95}
        />
      </svg>

      <div
        style={{
          position: 'absolute',
          left: 0,
          right: 0,
          top: 726,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
        }}
      >
        <div style={{opacity: 1 - ramp(frame, total * 0.34, 24)}}>
          <Eyebrow size={19} color={C.muted}>
            dozens of clans
          </Eyebrow>
        </div>
        <div
          style={{
            position: 'absolute',
            top: 0,
            opacity: ramp(frame, total * 0.42, 26),
            display: 'flex',
            gap: 54,
            alignItems: 'center',
          }}
        >
          <div
            style={{
              position: 'relative',
              opacity: 0.55,
            }}
          >
            <Eyebrow size={19} color={C.dim}>
              by birth
            </Eyebrow>
            <div
              style={{
                position: 'absolute',
                left: -6,
                right: -6,
                top: '52%',
                height: 1,
                background: C.dim,
                transform: `scaleX(${ramp(frame, total * 0.52, 20)})`,
                transformOrigin: 'left',
              }}
            />
          </div>
          <div style={{opacity: ramp(frame, total * 0.6, 26)}}>
            <Eyebrow size={19} color={C.goldBright}>
              by loyalty and competence
            </Eyebrow>
          </div>
        </div>
      </div>
    </AbsoluteFill>
  );
};

// --------------------------------------------------------------------- decimal

/**
 * The army in tens.
 *
 * Each row's field is ten tiles, and each tile is a miniature of the whole row
 * above it — which is the actual structure being described, so the picture is
 * the argument rather than an illustration of it. Rows three and four use a
 * CSS dot texture instead of ten thousand circles; the recursion still reads
 * and the frame renders in milliseconds.
 */
const UNITS = [
  {mn: 'Arban', gloss: 'ten men', n: 10},
  {mn: 'Zuun', gloss: 'ten arbans', n: 100},
  {mn: 'Myangan', gloss: 'ten zuun', n: 1000},
  {mn: 'Tümen', gloss: 'ten myangan', n: 10000},
];

const dotTexture = (pitch: number, r: number, alpha: number) => ({
  backgroundImage: `radial-gradient(circle at ${pitch / 2}px ${pitch / 2}px, rgba(237,227,212,${alpha}) ${r}px, rgba(0,0,0,0) ${r + 0.7}px)`,
  backgroundSize: `${pitch}px ${pitch}px`,
});

const Tile: React.FC<{level: number}> = ({level}) => {
  if (level === 0) {
    return (
      <div
        style={{
          width: 58,
          height: 58,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <div style={{width: 11, height: 11, borderRadius: 11, background: C.paper}} />
      </div>
    );
  }
  if (level === 1) {
    return (
      <div
        style={{
          width: 58,
          height: 58,
          display: 'grid',
          gridTemplateColumns: 'repeat(4, 1fr)',
          alignContent: 'center',
          justifyItems: 'center',
          gap: 3,
          padding: 6,
          border: `1px solid ${C.lineSoft}`,
        }}
      >
        {new Array(10).fill(0).map((_, i) => (
          <div
            key={i}
            style={{width: 6, height: 6, borderRadius: 6, background: C.paper, opacity: 0.92}}
          />
        ))}
      </div>
    );
  }
  return (
    <div
      style={{
        width: 58,
        height: 58,
        border: `1px solid ${C.lineSoft}`,
        ...dotTexture(level === 2 ? 7 : 4.4, level === 2 ? 1.5 : 1.0, 0.88),
      }}
    />
  );
};

export const DecimalScene: React.FC<{total: number}> = ({total}) => {
  const frame = useCurrentFrame();
  const fade = holdFade(frame, total, 18, 20);
  const each = total * 0.19;
  return (
    <AbsoluteFill style={{background: C.ink0, opacity: fade}}>
      <DiagramGround glow={0.18} />
      <AbsoluteFill style={{padding: '84px 110px'}}>
      <div style={{opacity: ramp(frame, 4, 26)}}>
        <Eyebrow size={17} color={C.gold}>
          the army was built in tens
        </Eyebrow>
      </div>
      <div style={{height: 46}} />
      {UNITS.map((u, i) => {
        const on = ramp(frame, 18 + i * each, 30);
        if (on <= 0.001) return <div key={u.mn} style={{height: 176}} />;
        const shown = Math.round(interpolate(on, [0, 1], [0, 10]));
        return (
          <div
            key={u.mn}
            style={{
              height: 176,
              display: 'flex',
              alignItems: 'center',
              gap: 40,
              opacity: 0.35 + 0.65 * on,
              borderTop: `1px solid ${C.lineSoft}`,
            }}
          >
            <div style={{width: 300, transform: `translateX(${(1 - on) * -18}px)`}}>
              <Display size={54} weight={400} color={C.paper}>
                {u.mn}
              </Display>
              <div style={{height: 8}} />
              <Body size={20} color={C.dim}>
                {u.gloss}
              </Body>
            </div>
            <div
              style={{
                width: 260,
                textAlign: 'right',
                fontFamily: NUM,
                fontSize: 76,
                fontWeight: 300,
                color: C.goldBright,
                fontVariantNumeric: 'tabular-nums lining-nums',
              }}
            >
              {Math.round(u.n * on).toLocaleString('en-US')}
            </div>
            <div style={{display: 'flex', gap: 12, marginLeft: 30}}>
              {new Array(10).fill(0).map((_, k) => (
                <div
                  key={k}
                  style={{
                    opacity: k < shown ? 1 : 0,
                    transform: `scale(${k < shown ? 1 : 0.7})`,
                  }}
                >
                  <Tile level={i} />
                </div>
              ))}
            </div>
          </div>
        );
      })}
      <div
        style={{
          marginTop: 22,
          opacity: ramp(frame, 18 + 3.4 * each, 30),
        }}
      >
        <Eyebrow size={16} color={C.muted}>
          ten thousand riders on one instruction
        </Eyebrow>
      </div>
      </AbsoluteFill>
    </AbsoluteFill>
  );
};

// ------------------------------------------------------------------------- kit

export const KitScene: React.FC<{total: number}> = ({total}) => {
  const frame = useCurrentFrame();
  const fade = holdFade(frame, total, 18, 20);
  const a = ramp(frame, 12, 34);
  const b = ramp(frame, total * 0.34, 34);
  const c = ramp(frame, total * 0.66, 34);
  const shot = ramp(frame, total * 0.2, 30);
  return (
    <AbsoluteFill style={{background: C.ink0, opacity: fade}}>
      <DiagramGround glow={0.21} />
      <div
        style={{
          position: 'absolute',
          inset: 0,
          display: 'flex',
          alignItems: 'center',
        }}
      >
        {/* the bow */}
        <div style={{flex: 1, paddingLeft: 130, opacity: a}}>
          <div style={{display: 'flex', alignItems: 'center', gap: 26}}>
            <div style={{transform: `translateY(${(1 - a) * 10}px)`}}>
              <Glyph name="bow" size={188} color={C.goldBright} weight={0.9} />
            </div>
            <svg width={300} height={188}>
              {/* the arrow's flight, drawn once as it is described */}
              <line
                x1={0}
                y1={94}
                x2={280 * shot}
                y2={94 - 26 * shot * (1 - shot) * 4}
                stroke={C.gold}
                strokeOpacity={0.55}
                strokeWidth={1.2}
                strokeDasharray="7 7"
              />
              <circle
                cx={280 * shot}
                cy={94 - 26 * shot * (1 - shot) * 4}
                r={3.6}
                fill={C.goldBright}
                opacity={shot > 0.02 ? 1 : 0}
              />
            </svg>
          </div>
          <div style={{height: 34}} />
          <Eyebrow size={18} color={C.paper}>
            composite bow
          </Eyebrow>
          <div style={{height: 14}} />
          <Body size={24} color={C.dim}>
            horn, wood and sinew, glued in layers
          </Body>
        </div>

        <div style={{width: 1, height: 420, background: C.lineSoft, opacity: b}} />

        {/* the horses */}
        <div style={{flex: 1, paddingLeft: 100, opacity: b}}>
          <div style={{display: 'flex', alignItems: 'flex-end', gap: 6}}>
            <Glyph name="rider" size={104} color={C.paper} weight={1.2} />
            {new Array(4).fill(0).map((_, i) => (
              <div
                key={i}
                style={{
                  opacity: ramp(frame, total * 0.4 + i * 8, 22),
                  transform: `translateY(${(1 - ramp(frame, total * 0.4 + i * 8, 22)) * 8}px)`,
                }}
              >
                <Glyph name="horse" size={86} color={C.gold} weight={1.15} />
              </div>
            ))}
          </div>
          <div style={{height: 34}} />
          <Eyebrow size={18} color={C.paper}>
            one rider, several horses
          </Eyebrow>
          <div style={{height: 14}} />
          <Body size={24} color={C.dim}>
            remounts, milk and meat, all of it walking
          </Body>
        </div>
      </div>

      <div
        style={{
          position: 'absolute',
          left: 0,
          right: 0,
          bottom: 108,
          display: 'flex',
          justifyContent: 'center',
          opacity: c,
        }}
      >
        <Eyebrow size={19} color={C.goldBright}>
          an army that carries its own supply does not need supply lines
        </Eyebrow>
      </div>
    </AbsoluteFill>
  );
};

// ------------------------------------------------------------------------- yam

export const YamScene: React.FC<{total: number}> = ({total}) => {
  const frame = useCurrentFrame();
  const fade = holdFade(frame, total, 18, 20);
  const STATIONS = 9;
  const x0 = 190;
  const x1 = 1730;
  const y = 452;
  const line = ramp(frame, 14, 40);
  // Two riders in sequence, each handing the message on: the pulse restarts
  // from a station rather than from the start, which is the point of a relay.
  const cycle = interpolate(frame, [total * 0.2, total * 0.94], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
    easing: Easing.linear,
  });
  const px = x0 + (x1 - x0) * cycle;
  const reached = Math.floor(cycle * (STATIONS - 1) + 0.0001);

  return (
    <AbsoluteFill style={{background: C.ink0, opacity: fade}}>
      <DiagramGround glow={0.20} />
      <div style={{position: 'absolute', left: 110, top: 96, opacity: ramp(frame, 4, 26)}}>
        <Eyebrow size={17} color={C.gold}>
          the yam · relay stations
        </Eyebrow>
      </div>
      <svg width={MAP_W} height={MAP_H} style={{position: 'absolute', inset: 0}}>
        <line
          x1={x0}
          y1={y}
          x2={x0 + (x1 - x0) * line}
          y2={y}
          stroke={C.line}
          strokeWidth={1.2}
        />
        {new Array(STATIONS).fill(0).map((_, i) => {
          const sx = x0 + ((x1 - x0) / (STATIONS - 1)) * i;
          const lit = i <= reached ? 1 : 0;
          const on = ramp(frame, 14 + i * 5, 24);
          return (
            <g key={i} opacity={on}>
              <line x1={sx} y1={y - 16} x2={sx} y2={y + 16} stroke={C.line} strokeWidth={1} />
              <circle
                cx={sx}
                cy={y}
                r={5.2}
                fill={lit ? C.goldBright : C.ink0}
                stroke={lit ? C.goldBright : C.dim}
                strokeWidth={1.2}
              />
              {lit ? (
                <circle cx={sx} cy={y} r={13} fill="none" stroke={C.gold} strokeOpacity={0.25} strokeWidth={1} />
              ) : null}
              <g
                transform={`translate(${sx - 29}, ${y + 30})`}
                opacity={lit ? 0.95 : 0.4}
              >
                <Glyph name="horse" size={58} color={lit ? C.gold : C.dim} weight={1.15} />
              </g>
            </g>
          );
        })}
        {/* the message itself */}
        {cycle > 0.001 ? (
          <>
            <line
              x1={x0}
              y1={y - 44}
              x2={px}
              y2={y - 44}
              stroke={C.goldBright}
              strokeOpacity={0.35}
              strokeWidth={1.4}
            />
            <circle cx={px} cy={y - 44} r={6} fill={C.goldBright} />
            <circle cx={px} cy={y - 44} r={16} fill="none" stroke={C.goldBright} strokeOpacity={0.3} strokeWidth={1} />
          </>
        ) : null}
      </svg>
      <div
        style={{
          position: 'absolute',
          left: 0,
          right: 0,
          bottom: 292,
          display: 'flex',
          justifyContent: 'center',
          flexDirection: 'column',
          alignItems: 'center',
          gap: 16,
          opacity: ramp(frame, total * 0.42, 30),
        }}
      >
        <Eyebrow size={18} color={C.paper}>
          hand off, and keep going
        </Eyebrow>
        <Body size={22} color={C.dim}>
          fresh horses standing ready at every station, day and night
        </Body>
      </div>
    </AbsoluteFill>
  );
};

// ------------------------------------------------------------------------ stat

export const StatScene: React.FC<{
  big: string;
  unit?: string;
  note: string;
  tone?: 'gold' | 'blood';
  total: number;
}> = ({big, unit, note, tone = 'gold', total}) => {
  const frame = useCurrentFrame();
  const fade = holdFade(frame, total, 18, 20);
  const a = ramp(frame, 8, 34);
  const accent = tone === 'blood' ? C.bloodSoft : C.goldBright;
  return (
    <AbsoluteFill
      style={{
        background: C.ink0,
        alignItems: 'center',
        justifyContent: 'center',
        opacity: fade,
      }}
    >
      <AbsoluteFill
        style={{
          background:
            tone === 'blood'
              ? 'radial-gradient(62% 56% at 50% 48%, rgba(156,55,34,0.16) 0%, rgba(0,0,0,0) 72%)'
              : 'radial-gradient(62% 56% at 50% 48%, rgba(200,160,85,0.09) 0%, rgba(0,0,0,0) 72%)',
        }}
      />
      <div style={{alignItems: 'center', display: 'flex', flexDirection: 'column'}}>
        <div
          style={{
            display: 'flex',
            alignItems: 'baseline',
            gap: 22,
            opacity: a,
            transform: `translateY(${(1 - a) * 14}px)`,
          }}
        >
          <Display size={210} weight={300} color={accent}>
            {big}
          </Display>
          {unit ? (
            <Display size={64} weight={300} color={C.muted}>
              {unit}
            </Display>
          ) : null}
        </div>
        <div style={{height: 40}} />
        <Rule progress={ramp(frame, 26, 34)} width={420} color={accent} />
        <div style={{height: 34}} />
        <div style={{opacity: ramp(frame, 34, 34), maxWidth: 1100, textAlign: 'center'}}>
          <Body size={30} color={C.muted}>
            {note}
          </Body>
        </div>
      </div>
    </AbsoluteFill>
  );
};

// ---------------------------------------------------------------------- routes

/** The Pax Mongolica: the same map, with the road open instead of the army. */
export const RoutesScene: React.FC<{total: number}> = ({total}) => {
  const frame = useCurrentFrame();
  const fade = holdFade(frame, total, 20, 20);
  const draw = interpolate(frame, [total * 0.12, total * 0.86], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
    easing: Easing.bezier(0.36, 0.02, 0.28, 1),
  });
  const full = routeD(ROUTE.length);
  // The polyline is drawn on with a dash offset, so the road opens westward at
  // a readable pace instead of appearing whole.
  const LEN = 4200;
  return (
    <AbsoluteFill style={{background: C.ink0, opacity: fade}}>
      <svg width={MAP_W} height={MAP_H} viewBox={`0 0 ${MAP_W} ${MAP_H}`} style={{position: 'absolute', inset: 0}}>
        <path d={SPHERE_D} fill="#080A0D" />
        <path d={LAND_D} fill="#2A241B" stroke="#584B39" strokeWidth={1.3} />
        <path d={LAND_D} fill="none" stroke="#6E5F49" strokeOpacity={0.5} strokeWidth={0.6} />
        <path
          d={full}
          fill="none"
          stroke={C.goldBright}
          strokeWidth={2.2}
          strokeOpacity={0.9}
          strokeDasharray={LEN}
          strokeDashoffset={LEN * (1 - draw)}
          strokeLinecap="round"
        />
        <path
          d={full}
          fill="none"
          stroke={C.gold}
          strokeWidth={9}
          strokeOpacity={0.12}
          strokeDasharray={LEN}
          strokeDashoffset={LEN * (1 - draw)}
        />
        {ROUTE.map((p, i) => {
          const xy = projection(p);
          if (!xy) return null;
          const on = ramp(frame, total * 0.12 + (i / ROUTE.length) * total * 0.74, 20);
          return (
            <g key={i} opacity={on}>
              <circle cx={xy[0]} cy={xy[1]} r={4} fill={C.paper} />
              <circle cx={xy[0]} cy={xy[1]} r={11} fill="none" stroke={C.gold} strokeOpacity={0.3} strokeWidth={1} />
            </g>
          );
        })}
      </svg>
      <div style={{position: 'absolute', left: 96, bottom: 80}}>
        <div
          style={{
            fontFamily: NUM,
            fontSize: 104,
            fontWeight: 300,
            color: C.paper,
            fontVariantNumeric: 'tabular-nums lining-nums',
            lineHeight: 1,
          }}
        >
          {Math.round(ROUTE_TOTAL_KM * draw).toLocaleString('en-US')}
          <span style={{fontSize: 46, color: C.muted, marginLeft: 16}}>km</span>
        </div>
        <div style={{height: 18}} />
        <Eyebrow size={16} color={C.gold}>
          one authority, from Zhongdu to the Black Sea
        </Eyebrow>
      </div>
    </AbsoluteFill>
  );
};

// ------------------------------------------------------------------------- end

/**
 * The end card carries the source credits.
 *
 * Every frame of footage and every painting in this film is public domain or
 * CC0, and naming them is both the courteous thing and the thing that lets a
 * viewer go and find the originals.
 */
export const EndScene: React.FC<{total: number}> = ({total}) => {
  const frame = useCurrentFrame();
  const a = ramp(frame, 10, 40);
  const b = ramp(frame, 56, 44);
  const out = 1 - ramp(frame, total - 46, 44, Easing.linear);
  return (
    <AbsoluteFill
      style={{
        background: C.ink0,
        alignItems: 'center',
        justifyContent: 'center',
        opacity: out,
      }}
    >
      <div style={{alignItems: 'center', display: 'flex', flexDirection: 'column'}}>
        <div style={{opacity: a}}>
          <Display size={78} weight={300} style={{textAlign: 'center'}}>
            Genghis Khan
          </Display>
        </div>
        <div style={{height: 20, opacity: a}} />
        <div style={{opacity: a}}>
          <Eyebrow size={16} color={C.gold}>
            and the Mongol Empire
          </Eyebrow>
        </div>
        <div style={{height: 64}} />
        <div style={{opacity: b}}>
          <Rule progress={b} width={420} color={C.line} />
        </div>
        <div style={{height: 44}} />
        <div style={{opacity: b, textAlign: 'center', lineHeight: 2.05}}>
          <Eyebrow size={13} color={C.dim}>
            archive footage
          </Eyebrow>
          <div style={{height: 12}} />
          <Body size={20} color={C.muted}>
            Storm Over Asia, V. I. Pudovkin, 1928 — public domain
          </Body>
          <Body size={20} color={C.muted}>
            Morden–Clark Asiatic Expedition, 1926 — US National Archives, CC0
          </Body>
          <div style={{height: 30}} />
          <Eyebrow size={13} color={C.dim}>
            paintings
          </Eyebrow>
          <div style={{height: 12}} />
          <Body size={20} color={C.muted}>
            Rashid al-Din, Jami’ al-tawārīkh, 14th–15th c. · Yuan dynasty album
          </Body>
          <Body size={20} color={C.muted}>
            Catalan Atlas, 1375 — via Wikimedia Commons, public domain
          </Body>
        </div>
      </div>
    </AbsoluteFill>
  );
};
