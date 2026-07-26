import * as React from 'react';
import type {Decorator, Preview} from '@storybook/react';
import {applyTheme, type ThemeMode} from '../src/theme/applyTheme';
import '../src/index.css';

/** Applies the design-token CSS variables for the toolbar-selected theme, then
 *  frames every story on the themed page background so contrast is realistic. */
const withTheme: Decorator = (Story, context) => {
  const mode = (context.globals.theme as ThemeMode) ?? 'dark';
  React.useEffect(() => {
    applyTheme(mode);
  }, [mode]);
  return (
    <div style={{background: 'var(--color-bg)', color: 'var(--color-text)', minHeight: '100vh', padding: '2rem'}}>
      <Story />
    </div>
  );
};

const preview: Preview = {
  parameters: {
    controls: {matchers: {color: /(background|color)$/i, date: /Date$/i}},
    a11y: {test: 'error'},
  },
  globalTypes: {
    theme: {
      description: 'Design-token theme',
      defaultValue: 'dark',
      toolbar: {
        title: 'Theme',
        icon: 'circlehollow',
        items: [
          {value: 'light', title: 'Light', icon: 'sun'},
          {value: 'dark', title: 'Dark', icon: 'moon'},
        ],
        dynamicTitle: true,
      },
    },
  },
  decorators: [withTheme],
};

export default preview;
