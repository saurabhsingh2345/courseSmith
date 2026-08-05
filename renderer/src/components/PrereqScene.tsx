import {ResolvedTheme} from '../theme/theme';
import {CourseListScene, CourseRow} from './CourseListScene';

// The floor a lesson stands on, with the skippable parts visibly skippable.
//
// A skippable assumption is struck through and painted in the neutral role, so
// the difference between "you need this" and "nice to have" is legible in one
// glance rather than read out of a sentence. That distinction is the reason the
// template exists: a viewer who cannot see it treats all of it as the floor.

type Assumption = {
  item: string;
  source: string;
  where: string;
  skippable?: boolean;
  breaks?: string;
};

export const PrereqScene: React.FC<{
  theme: ResolvedTheme;
  sceneStartMs: number;
  props: Record<string, unknown>;
}> = ({theme, sceneStartMs, props}) => {
  const items = (Array.isArray(props.assumptions) ? props.assumptions : []) as Assumption[];
  return (
    <CourseListScene
      theme={theme}
      sceneStartMs={sceneStartMs}
      props={props}
      count={items.length}
      row={(i, weight, focused) => {
        const a = items[i];
        if (!a) return null;
        return (
          <CourseRow
            theme={theme}
            weight={weight}
            focused={focused}
            // Skippable rows drop to the neutral colour: they are not the floor,
            // and colouring them like it would undo the whole distinction.
            accent={a.skippable ? theme.textMuted : theme.accentLimit}
            primary={a.item}
            struck={a.skippable}
            secondary={a.skippable && a.breaks ? a.breaks : a.where}
            tag={a.skippable ? 'skip, but' : a.source === 'taught' ? 'from' : 'bring'}
          />
        );
      }}
    />
  );
};
