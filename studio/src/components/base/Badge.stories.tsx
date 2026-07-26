import type {Meta, StoryObj} from '@storybook/react';
import {Badge} from './Badge';

const meta = {
  title: 'Base/Badge',
  component: Badge,
  tags: ['autodocs'],
  argTypes: {
    variant: {
      options: ['default', 'secondary', 'success', 'error', 'warning', 'outline'],
      control: {type: 'select'},
    },
  },
  args: {children: 'Badge', variant: 'default'},
} satisfies Meta<typeof Badge>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const AllVariants: Story = {
  render: () => (
    <div className="flex flex-wrap gap-2">
      <Badge variant="default">Default</Badge>
      <Badge variant="secondary">Secondary</Badge>
      <Badge variant="success">Passed</Badge>
      <Badge variant="error">Failed</Badge>
      <Badge variant="warning">Review</Badge>
      <Badge variant="outline">Draft</Badge>
    </div>
  ),
};
