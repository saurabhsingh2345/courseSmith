import {ResolvedTheme, withAlpha} from '../theme/theme';
import {CourseListScene, CourseRow, LIST_W} from './CourseListScene';

// The task that closes a lesson, and the condition that settles it.
//
// The done-when is pinned under the steps in a bordered block rather than being
// a fourth list row, because it is a different KIND of thing: the steps are what
// you do, and this is how you find out you did it. A viewer alone at the end of
// a video with nobody to mark their work needs that one line to be the most
// findable thing in the frame.

export const CheckpointScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const list = (Array.isArray(props.list) ? props.list : []) as string[];
  const task = String(props.task ?? '');
  const done = String(props.done ?? '');
  const stuck = String(props.stuck ?? '');
  return (
    <CourseListScene
      theme={theme}
      sceneStartMs={sceneStartMs}
      props={props}
      count={list.length}
      lead={task}
      row={(i, weight, focused) => (
        <CourseRow
          theme={theme}
          weight={weight}
          focused={focused}
          accent={theme.accentQuantity}
          primary={`${i + 1}.  ${list[i] ?? ''}`}
        />
      )}
      tail={
        done ? (
          <div
            style={{
              border: `2px solid ${withAlpha(theme.accentQuantity, 0.55)}`,
              borderRadius: 14,
              padding: '22px 28px',
              width: LIST_W,
              display: 'flex',
              flexDirection: 'column',
              gap: 8,
            }}
          >
            <div
              style={{
                fontFamily: theme.fontMono,
                fontSize: 15,
                letterSpacing: 3,
                textTransform: 'uppercase',
                color: theme.accentQuantity,
              }}
            >
              done when
            </div>
            <div style={{fontFamily: theme.fontDisplay, fontSize: 34, fontWeight: 700, color: theme.text, lineHeight: 1.2}}>
              {done}
            </div>
            {stuck ? (
              <div style={{fontFamily: theme.fontBody, fontSize: 23, color: theme.textMuted, marginTop: 4}}>
                Stuck? {stuck}
              </div>
            ) : null}
          </div>
        ) : null
      }
    />
  );
};
