import {ResolvedTheme, withAlpha} from '../theme/theme';
import {CourseListScene, CourseRow, LIST_W} from './CourseListScene';

// What earlier lessons established, each tagged with where it came from.
//
// The thread sits UNDER the list rather than above it, and that is deliberate:
// it is the payoff, not the premise. Read top-down the clip is a set of things
// you already knew separately, and then one line says what they were all
// building — which is the difference between a recap and an index.

type Claim = {claim: string; from: string};

export const RecapScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const claims = (Array.isArray(props.claims) ? props.claims : []) as Claim[];
  const thread = String(props.thread ?? '');
  return (
    <CourseListScene
      theme={theme}
      sceneStartMs={sceneStartMs}
      props={props}
      count={claims.length}
      row={(i, weight, focused) => (
        <CourseRow
          theme={theme}
          weight={weight}
          focused={focused}
          accent={theme.accentRival}
          primary={claims[i]?.claim ?? ''}
          secondary={claims[i]?.from}
          tag="from"
        />
      )}
      tail={
        thread ? (
          <div
            style={{
              borderTop: `2px solid ${withAlpha(theme.accentText, 0.5)}`,
              paddingTop: 22,
              width: LIST_W,
              fontFamily: theme.fontDisplay,
              fontSize: 34,
              fontWeight: 600,
              lineHeight: 1.25,
              color: theme.text,
            }}
          >
            {thread}
          </div>
        ) : null
      }
    />
  );
};
