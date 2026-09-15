// The cartography for the conquest sequence.
//
// Two deliberate choices:
//
// 1. LAND, NOT COUNTRIES. `world-atlas` ships countries-*.json and land-*.json.
//    Drawing modern national borders under a 13th-century empire is a factual
//    error the audience can see, so this uses the land mesh: one coastline, no
//    borders. The empire outline is then the only political line on the map,
//    which is also the right emphasis.
//
// 2. ORTHOGRAPHIC. Eurasia from the Danube to the Pacific is 110 degrees of
//    longitude. On any flat projection that is a wide, distorted band; on a
//    globe centred over Central Asia it reads as one continuous landmass with
//    curvature, which is closer to the thing being described — a single empire
//    spanning a hemisphere.
//
// Everything is projected once at module load. The geometry never changes
// across the sequence; only which parts are revealed does.

import {geoOrthographic, geoPath, geoGraticule10, geoDistance} from 'd3-geo';
import {feature} from 'topojson-client';
import land50 from 'world-atlas/land-50m.json';

export const MAP_W = 1920;
export const MAP_H = 1080;

/** Centre of the projection: over the western steppe, so the empire is centred. */
const CENTRE: [number, number] = [78, 40];

/** The homeland the conquest radiates out from — the Onon/Kherlen heartland. */
export const HOMELAND: [number, number] = [106, 47.5];

export const projection = geoOrthographic()
  .rotate([-CENTRE[0], -CENTRE[1]])
  .clipAngle(90);

// Scale and translate are fitted to the empire itself rather than hand-tuned.
// The thing that has to be legible is the territory at its greatest extent —
// Anatolia to Korea — so that shape is what the frame is fitted to, with a
// margin left at the bottom for the year read-out and the phase label.
const FIT_BOX: [[number, number], [number, number]] = [
  [170, 150],
  [1750, 880],
];

const path = geoPath(projection);

const collection = feature(
  land50 as never,
  (land50 as never as {objects: {land: unknown}}).objects.land as never,
) as unknown as {type: string; features?: unknown[]; geometry?: unknown};

const LAND_GEO = (collection.features
  ? {type: 'FeatureCollection', features: collection.features}
  : collection) as never;

// ------------------------------------------------------------------- the empire
//
// Approximate boundaries, traced as lat/lon rings. These are a documentary's
// map, not a survey: the Mongol frontier was a moving band of tributary
// steppe rather than a line, and no two historical atlases agree on it. Each
// phase is drawn generously overlapping its predecessor so the union has no
// seams.

type Ring = [number, number][];

/**
 * Normalise a ring's winding for d3-geo.
 *
 * d3-geo reads a polygon on the SPHERE by its winding, and a ring wound the
 * wrong way means "the entire sphere except this shape". These rings were
 * traced by hand and ended up mixed — some clockwise, some not — which rendered
 * three of the six phases as the whole ocean and made `fitExtent` fit the globe
 * instead of the empire.
 *
 * Rather than reverse them all (which just moves the bug to the other three),
 * each ring is normalised by its planar signed area. The convention was
 * measured against a ring known to render correctly: d3 wants NEGATIVE signed
 * area for the interior of a region traced in lon/lat.
 */
const signedArea = (r: Ring) => {
  let s = 0;
  for (let i = 0; i < r.length - 1; i++) {
    s += r[i][0] * r[i + 1][1] - r[i + 1][0] * r[i][1];
  }
  return s / 2;
};

const wind = (r: Ring): Ring => (signedArea(r) > 0 ? r.slice().reverse() : r);
export type Phase = {
  /** First year this territory is shown. */
  year: number;
  label: string;
  rings: Ring[];
};

export const PHASES: Phase[] = [
  {
    year: 1206,
    label: 'the Mongol homeland',
    rings: [
      [
        [87, 50], [92, 52.5], [99, 52], [107, 51.5], [113, 50], [118, 49.5],
        [121, 47], [119, 44], [114, 42.5], [108, 41.5], [102, 41.5],
        [96, 42.5], [90, 45.5], [87, 50],
      ],
    ],
  },
  {
    year: 1211,
    label: 'Western Xia and the Jin',
    rings: [
      [
        [96, 43], [102, 41.5], [108, 41.5], [114, 42.5], [119, 44], [124, 44.5],
        [126.5, 41], [124, 38.5], [120, 36], [116, 34.5], [110, 34],
        [104, 34.5], [99, 36.5], [95.5, 39.5], [96, 43],
      ],
    ],
  },
  {
    year: 1219,
    label: 'Transoxiana and Khorasan',
    rings: [
      [
        [88, 50], [80, 50.5], [72, 50], [64, 48], [57, 45], [51, 42], [48, 39],
        [50, 36], [55, 33.5], [61, 32], [67, 32.5], [72, 34.5], [77, 37],
        [81, 40], [85, 44], [89, 46], [88, 50],
      ],
    ],
  },
  {
    year: 1227,
    label: 'at his death',
    rings: [
      // The Persian lobe and the Caspian steppe held at 1227.
      [
        [51, 42], [46, 44], [42, 41], [41, 37], [44, 34], [49, 31], [56, 30],
        [63, 30.5], [68, 32], [64, 32.5], [57, 33.5], [51, 36], [48, 39],
        [51, 42],
      ],
      // Tibet and the Tangut west.
      [
        [79, 36], [85, 35], [92, 33.5], [99, 33], [102, 31], [97, 29],
        [90, 29], [84, 31], [79, 33.5], [79, 36],
      ],
    ],
  },
  {
    year: 1240,
    label: 'the Kipchak steppe and the Rus’',
    rings: [
      [
        [57, 46], [51, 42], [46, 45], [40, 48], [34, 50], [28, 50], [24, 52],
        [27, 56], [35, 58.5], [45, 58], [52, 55], [57, 51], [57, 46],
      ],
    ],
  },
  {
    year: 1258,
    label: 'Iraq, Anatolia and Korea',
    rings: [
      [
        [42, 41], [37, 39], [32, 38], [30, 40], [34, 42], [40, 43], [45, 44],
        [46, 42], [42, 41],
      ],
      [
        [44, 34], [43, 31], [46, 29], [49, 30], [49, 31], [44, 34],
      ],
      [
        [124, 38.5], [126.5, 41], [129.5, 42.5], [130.5, 38.5], [128, 35],
        [125.5, 36.5], [124, 38.5],
      ],
    ],
  },
];

const ringToD = (rings: Ring[]) =>
  path({
    type: 'MultiPolygon',
    coordinates: rings.map((r) => [wind(r)]),
  } as never) ?? '';

projection.fitExtent(FIT_BOX, {
  type: 'MultiPolygon',
  coordinates: PHASES.flatMap((p) => p.rings.map((r) => [wind(r)])),
} as never);

export const PHASE_PATHS = PHASES.map((p) => ({
  ...p,
  d: ringToD(p.rings),
}));

// Drawn after the fit, or every path would carry the pre-fit scale.
export const LAND_D = path(LAND_GEO) ?? '';
export const GRATICULE_D = path(geoGraticule10() as never) ?? '';
/** The globe's own disc, which is the sea fill. */
export const SPHERE_D = path({type: 'Sphere'} as never) ?? '';

/**
 * The reveal radius, in screen px from the projected homeland, that a phase
 * needs in order to be fully uncovered.
 *
 * The conquest is animated by one growing circle centred on Mongolia rather
 * than by fading each region in place, because that is what actually happened:
 * the empire is a distance from one valley, not a set of provinces switching
 * colour. Measuring the radius off the real projected vertices means the
 * circle always just clears the territory it is meant to.
 */
const home = projection(HOMELAND) as [number, number];
export const HOME_XY = home;

export const PHASE_RADIUS = PHASES.map((p) => {
  let max = 0;
  for (const ring of p.rings) {
    for (const v of ring) {
      const xy = projection(v);
      if (!xy) continue;
      const d = Math.hypot(xy[0] - home[0], xy[1] - home[1]);
      if (d > max) max = d;
    }
  }
  return max * 1.06;
});

/** Year → reveal radius, interpolated across the phase control points. */
export const radiusForYear = (year: number) => {
  const ys = PHASES.map((p) => p.year);
  if (year <= ys[0]) return PHASE_RADIUS[0];
  for (let i = 1; i < ys.length; i++) {
    if (year <= ys[i]) {
      const t = (year - ys[i - 1]) / (ys[i] - ys[i - 1]);
      return PHASE_RADIUS[i - 1] + t * (PHASE_RADIUS[i] - PHASE_RADIUS[i - 1]);
    }
  }
  return PHASE_RADIUS[PHASE_RADIUS.length - 1];
};

// --------------------------------------------------------------------- cities

export type City = {
  id: string;
  name: string;
  when: string;
  lonlat: [number, number];
  /** Which side of the dot the label sits on. */
  side?: 'l' | 'r';
};

export const CITIES: City[] = [
  {id: 'karakorum', name: 'Karakorum', when: 'the capital', lonlat: [102.8, 47.2], side: 'r'},
  {id: 'zhongdu', name: 'Zhongdu', when: '1215', lonlat: [116.4, 39.9], side: 'r'},
  {id: 'otrar', name: 'Otrar', when: '1218', lonlat: [68.3, 42.85], side: 'r'},
  {id: 'bukhara', name: 'Bukhara', when: '1220', lonlat: [64.4, 39.77], side: 'l'},
  {id: 'samarkand', name: 'Samarkand', when: '1220', lonlat: [66.97, 39.65], side: 'r'},
  {id: 'urgench', name: 'Urgench', when: '1221', lonlat: [60.63, 41.55], side: 'l'},
  {id: 'nishapur', name: 'Nishapur', when: '1221', lonlat: [58.8, 36.2], side: 'l'},
  {id: 'baghdad', name: 'Baghdad', when: '1258', lonlat: [44.36, 33.31], side: 'l'},
  {id: 'kyiv', name: 'Kyiv', when: '1240', lonlat: [30.5, 50.45], side: 'l'},
];

export const cityXY = (c: City) => projection(c.lonlat) as [number, number];

/** Great-circle km between two lon/lat points, for the trade-route captions. */
export const km = (a: [number, number], b: [number, number]) =>
  Math.round(geoDistance(a, b) * 6371);

// ------------------------------------------------------------- trade routes
//
// The Pax Mongolica beat. Drawn as great-circle arcs between the nodes a
// 13th-century merchant would actually have travelled between.

export const ROUTE: [number, number][] = [
  [116.4, 39.9],   // Zhongdu
  [102.8, 47.2],   // Karakorum
  [87.6, 43.8],    // Beshbalik / the Tarim road
  [66.97, 39.65],  // Samarkand
  [52.5, 37.8],    // the Caspian shore
  [46.3, 38.1],    // Tabriz
  [35.4, 45.2],    // Caffa, on the Black Sea
];

export const routeD = (n: number) => {
  const pts = ROUTE.slice(0, Math.max(2, n));
  return (
    path({type: 'LineString', coordinates: pts} as never) ?? ''
  );
};

export const ROUTE_TOTAL_KM = ROUTE.slice(1).reduce(
  (acc, p, i) => acc + km(ROUTE[i], p),
  0,
);
