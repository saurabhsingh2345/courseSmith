// The world map: projection, and country paths keyed by name.
//
// The atlas is `world-atlas`'s countries-110m TopoJSON — Natural Earth data,
// public domain, 108KB. The 50m and 10m builds ship in the same package and
// both are the wrong choice here: at 1080p, over a shape that is on screen for
// thirty seconds and often half-dimmed, the extra coastline detail is invisible
// and costs a megabyte of bundle plus a slower parse on every frame worker.
//
// Everything is computed once at module load. Projecting 176 countries per
// frame is pure waste — the geometry never changes, only the fill does.

import {geoNaturalEarth1, geoPath} from 'd3-geo';
import {feature} from 'topojson-client';
import world from 'world-atlas/countries-110m.json';

/**
 * Antarctica is dropped.
 *
 * It is in the atlas, it is never the subject of an explainer, and at Natural
 * Earth's extent it is a bar across the bottom of the frame. It also drags the
 * projection's fit: including it costs about a fifth of the box's height to
 * something nobody is looking at, and leaves the inhabited world floating in
 * the upper two thirds.
 */
const OMIT = new Set(['Antarctica']);

/** Nominal width the map is projected into; the height follows from the fit. */
const BOX_W = 1000;

export type CountryPath = {name: string; d: string};

type Feat = {properties: {name: string}; geometry: unknown};

const collection = feature(
  world as never,
  (world as never as {objects: {countries: unknown}}).objects.countries as never,
) as unknown as {features: Feat[]};

const drawn = collection.features.filter((f) => !OMIT.has(f.properties.name));

// Fit to the width, then measure what height that actually produced, so the
// box is the map's own aspect rather than a guess. A box taller than the map
// would letterbox it inside its own SVG and read as a chart that failed to
// fill its space.
const projection = geoNaturalEarth1().fitWidth(BOX_W, {
  type: 'FeatureCollection',
  features: drawn,
} as never);
const path = geoPath(projection);
const [[, minY], [, maxY]] = path.bounds({
  type: 'FeatureCollection',
  features: drawn,
} as never);

export const MAP_BOX = {w: BOX_W, h: Math.ceil(maxY - minY)};

export const COUNTRIES: CountryPath[] = drawn
  .map((f) => ({name: f.properties.name, d: path(f as never) ?? ''}))
  // A few slivers project to nothing at this resolution. An empty `d` renders
  // as a console error in some engines and a stray dot in others, so they are
  // dropped rather than drawn.
  .filter((c) => c.d !== '');

/**
 * How far to translate the projected paths so they sit at the box's origin.
 *
 * fitWidth centres the map vertically in an unbounded space, so its top edge is
 * at minY rather than at zero. A caller draws into a MAP_BOX-sized viewBox and
 * applies this once to the whole group.
 */
export const MAP_OFFSET_Y = -minY;

/** Country paths by lower-cased name, for matching a plan's labels. */
export const COUNTRY_BY_NAME: Record<string, CountryPath> = Object.fromEntries(
  COUNTRIES.map((c) => [c.name.toLowerCase(), c]),
);
