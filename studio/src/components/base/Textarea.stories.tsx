import type {Meta, StoryObj} from '@storybook/react';
import {Textarea} from './Textarea';

const meta = {
  title: 'Base/Textarea',
  component: Textarea,
  tags: ['autodocs'],
  args: {placeholder: 'Write your lesson notes…', rows: 4},
} satisfies Meta<typeof Textarea>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Disabled: Story = {args: {disabled: true, value: 'Locked content'}};
