import type {Meta, StoryObj} from '@storybook/react';
import {Button} from './Button';

const meta = {
  title: 'Base/Button',
  component: Button,
  tags: ['autodocs'],
  argTypes: {
    variant: {
      options: ['primary', 'secondary', 'ghost', 'danger', 'outline'],
      control: {type: 'select'},
    },
    size: {options: ['sm', 'md', 'lg', 'icon'], control: {type: 'select'}},
  },
  args: {children: 'Button', variant: 'primary', size: 'md'},
} satisfies Meta<typeof Button>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Primary: Story = {args: {variant: 'primary'}};
export const Secondary: Story = {args: {variant: 'secondary'}};
export const Ghost: Story = {args: {variant: 'ghost'}};
export const Danger: Story = {args: {variant: 'danger', children: 'Delete'}};
export const Outline: Story = {args: {variant: 'outline'}};
export const Disabled: Story = {args: {disabled: true}};

export const AllVariants: Story = {
  render: () => (
    <div className="flex flex-wrap items-center gap-3">
      <Button variant="primary">Primary</Button>
      <Button variant="secondary">Secondary</Button>
      <Button variant="ghost">Ghost</Button>
      <Button variant="danger">Danger</Button>
      <Button variant="outline">Outline</Button>
    </div>
  ),
};
