// The conquest map.
//
// The empire is revealed by ONE circle growing out of the Mongol homeland,
// not by fading each province in place. That is the honest animation: the
// empire is a distance travelled from one valley, and a viewer reads the
// spreading disc as exactly that. Territory is additionally gated on its own
// year, so a radius that has reached Persia in the west does not also light up
// Korea in the east before 1258.

import React from 'react';
import {AbsoluteFill, interpolate, useCurrentFrame, Easing} from 'remotion';
import {
  LAND_D,
  GRATICULE_D,
  SPHERE_D,
  PHASE_PATHS,
  HOME_XY,
  radiusForYear,
  CITIES,
  cityXY,
  MAP_W,
  MAP_H,
} from './geo';
import {C, NUM, BODY, Eyebrow, ramp, holdFade} from './kit';

export const MapScene: React.FC<{
  from: number;
  to: number;
  cities?: string[];
  label?: string;
  total: number;
}> = ({from, to, cities = [], label, total}) => {
  const frame = useCurrentFrame();

  // The year advances across the middle of the shot, so the map is readable
  // before it starts moving and settles before the cut.
  const t = interpolate(frame, [total * 0.1, total * 0.86], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
    easing: Easing.bezier(0.34, 0.02, 0.26, 1),
  });
  const year = from + (to - from) * t;
  const radius = radiusForYear(year);
  const fade = holdFade(frame, total, 20, 18);

  const shown = new Set(cities);

  return (
    <AbsoluteFill style={{background: C.ink0, opacity: fade}}>
      <svg
        width={MAP_W}
        height={MAP_H}
        viewBox={`0 0 ${MAP_W} ${MAP_H}`}
        style={{position: 'absolute', inset: 0}}
      >
        <defs>
          <clipPath id="reveal">
            <circle cx={HOME_XY[0]} cy={HOME_XY[1]} r={radius} />
          </clipPath>
        </defs>

        {/* Sea, land, graticule — the ground the empire is drawn onto. */}
        <path d={SPHERE_D} fill="#080A0D" />
        <path d={GRATICULE_D} fill="none" stroke={C.gold} strokeOpacity={0.05} strokeWidth={0.8} />
        <path d={LAND_D} fill="#2A241B" stroke="#584B39" strokeWidth={1.3} />
        {/* A second, brighter coastline just inside the first: the coast is the
            only thing telling the viewer this is Asia, so it gets the contrast
            rather than the empire's edge. */}
        <path d={LAND_D} fill="none" stroke="#6E5F49" strokeOpacity={0.5} strokeWidth={0.6} />

        {/* The empire.

            The fills are OPAQUE inside one translucent group. Drawn as six
            semi-transparent paths instead, every overlap between phases
            compounded into a different shade and the empire read as a
            patchwork of blobs rather than as one territory. */}
        <g clipPath="url(#reveal)">
          <g opacity={0.34}>
            {PHASE_PATHS.map((p) => {
              // A phase appears as the year crosses its start, over three
              // years, so consecutive phases overlap rather than snapping on.
              const on = interpolate(year, [p.year - 0.5, p.year + 3], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              });
              if (on <= 0) return null;
              return <path key={p.year} d={p.d} fill={C.gold} opacity={on} />;
            })}
          </g>
          {/* The stage outlines sit above the fill, so the viewer can still see
              which campaign added which ground. */}
          {PHASE_PATHS.map((p) => {
            const on = interpolate(year, [p.year - 0.5, p.year + 3], [0, 1], {
              extrapolateLeft: 'clamp',
              extrapolateRight: 'clamp',
            });
            if (on <= 0) return null;
            return (
              <path
                key={`e${p.year}`}
                d={p.d}
                fill="none"
                stroke={C.goldBright}
                strokeOpacity={0.55 * on}
                strokeWidth={1.5}
              />
            );
          })}
        </g>

        {/* The homeland itself, marked once and never moved. */}
        <circle cx={HOME_XY[0]} cy={HOME_XY[1]} r={4} fill={C.goldBright} />
        <circle
          cx={HOME_XY[0]}
          cy={HOME_XY[1]}
          r={10 + 4 * Math.sin(frame / 14)}
          fill="none"
          stroke={C.goldBright}
          strokeOpacity={0.35}
          strokeWidth={1}
        />

        {CITIES.filter((c) => shown.has(c.id)).map((c) => {
          const [x, y] = cityXY(c);
          const when = parseInt(c.when, 10);
          // A city that has a date appears when the year reaches it; the
          // capital has no date and is present throughout.
          const on = Number.isNaN(when)
            ? ramp(frame, total * 0.12, 22)
            : interpolate(year, [when - 0.2, when + 1.4], [0, 1], {
                extrapolateLeft: 'clamp',
                extrapolateRight: 'clamp',
              });
          if (on <= 0.01) return null;
          const left = c.side === 'l';
          return (
            <g key={c.id} opacity={on}>
              <circle cx={x} cy={y} r={3.4} fill={C.paper} />
              <circle
                cx={x}
                cy={y}
                r={3.4 + 12 * (1 - on)}
                fill="none"
                stroke={C.paper}
                strokeOpacity={0.6 * (1 - on)}
                strokeWidth={1}
              />
              <line
                x1={left ? x - 8 : x + 8}
                y1={y}
                x2={left ? x - 8 - 26 * on : x + 8 + 26 * on}
                y2={y}
                stroke={C.paper}
                strokeOpacity={0.4}
                strokeWidth={0.9}
              />
              <text
                x={left ? x - 40 : x + 40}
                y={y - 3}
                textAnchor={left ? 'end' : 'start'}
                fill={C.paper}
                style={{fontFamily: BODY, fontSize: 21, fontWeight: 400}}
              >
                {c.name}
              </text>
              <text
                x={left ? x - 40 : x + 40}
                y={y + 18}
                textAnchor={left ? 'end' : 'start'}
                fill={C.gold}
                style={{
                  fontFamily: BODY,
                  fontSize: 15,
                  letterSpacing: 3,
                  fontWeight: 400,
                }}
              >
                {c.when.toUpperCase()}
              </text>
            </g>
          );
        })}
      </svg>

      {/* The read-out. Bottom-left, out of the way of the territory. */}
      <div style={{position: 'absolute', left: 96, bottom: 76}}>
        <div
          style={{
            fontFamily: NUM,
            fontSize: 118,
            fontWeight: 300,
            color: C.paper,
            lineHeight: 0.92,
            letterSpacing: 1,
            // The counter must not reflow as digits change.
            fontVariantNumeric: 'tabular-nums lining-nums',
          }}
        >
          {Math.round(year)}
        </div>
        {label ? (
          <div style={{marginTop: 14}}>
            <Eyebrow size={17} color={C.gold}>
              {label}
            </Eyebrow>
          </div>
        ) : null}
      </div>
    </AbsoluteFill>
  );
};
