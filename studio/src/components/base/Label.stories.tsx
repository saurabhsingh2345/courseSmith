import type {Meta, StoryObj} from '@storybook/react';
import {Label} from './Label';

const meta = {
  title: 'Base/Label',
  component: Label,
  tags: ['autodocs'],
  args: {children: 'Field label'},
} satisfies Meta<typeof Label>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
