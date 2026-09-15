// The cut, assembled.
//
// Beat lengths come from `timing.json`, which `build_voice.py` wrote after
// measuring Adam's actual delivery of each line. Nothing here guesses a
// duration, so a rewritten line reflows the film rather than desyncing it.

import React from 'react';
import {AbsoluteFill, Audio, Sequence, staticFile, useVideoConfig} from 'remotion';
import {C, Grain} from './kit';
import {BEATS, type Vis} from './edl';
import timing from './timing.json';
import {MapScene} from './MapScene';
import {
  FootageScene,
  PlateScene,
  TitleScene,
  QuoteScene,
  TribesScene,
  DecimalScene,
  KitScene,
  YamScene,
  StatScene,
  RoutesScene,
  EndScene,
} from './scenes';

/**
 * Cross-dissolve length.
 *
 * Every scene fades its own head and tail. Cut end-to-end that produces a dip
 * to black between all 27 beats, which at this length reads as 27 endings. Each
 * sequence instead runs XF frames past its slot so the outgoing tail and the
 * incoming head overlap, and the film dissolves instead.
 */
const XF = 14;

const renderVis = (vis: Vis, total: number) => {
  switch (vis.kind) {
    case 'footage':
      return (
        <FootageScene
          clip={vis.clip}
          fit={vis.fit}
          push={vis.push}
          credit={vis.credit}
          total={total}
        />
      );
    case 'plate':
      return (
        <PlateScene
          art={vis.art}
          focus={vis.focus}
          push={vis.push}
          credit={vis.credit}
          total={total}
        />
      );
    case 'title':
      return <TitleScene total={total} />;
    case 'quote':
      return <QuoteScene line={vis.line} sub={vis.sub} total={total} />;
    case 'tribes':
      return <TribesScene total={total} />;
    case 'decimal':
      return <DecimalScene total={total} />;
    case 'kit':
      return <KitScene total={total} />;
    case 'yam':
      return <YamScene total={total} />;
    case 'map':
      return (
        <MapScene
          from={vis.from}
          to={vis.to}
          cities={vis.cities}
          label={vis.label}
          total={total}
        />
      );
    case 'stat':
      return (
        <StatScene
          big={vis.big}
          unit={vis.unit}
          note={vis.note}
          tone={vis.tone}
          total={total}
        />
      );
    case 'routes':
      return <RoutesScene total={total} />;
    case 'end':
      return <EndScene total={total} />;
    default:
      return null;
  }
};

export const Film: React.FC = () => {
  const {durationInFrames} = useVideoConfig();
  const byId = new Map(timing.beats.map((b) => [b.id, b]));

  return (
    <AbsoluteFill style={{background: C.ink0}}>
      <Audio src={staticFile('genghis/bed.wav')} />

      {BEATS.map((beat) => {
        const t = byId.get(beat.id);
        if (!t) return null;
        const isLast = t.startFrame + t.durFrames >= durationInFrames - 1;
        // The last beat has nothing to dissolve into, so it keeps its slot.
        const total = t.durFrames + (isLast ? 0 : XF);
        return (
          <Sequence
            key={beat.id}
            from={t.startFrame}
            durationInFrames={total}
            name={beat.id}
          >
            {renderVis(beat.vis, total)}
          </Sequence>
        );
      })}

      {/* One emulsion over the whole film, so the drawn scenes and the 1926
          stock sit in the same grain rather than looking like two sources. */}
      <Grain amount={0.045} />
    </AbsoluteFill>
  );
};
