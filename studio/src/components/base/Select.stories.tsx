import type {Meta, StoryObj} from '@storybook/react';
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from './Select';

const meta = {
  title: 'Base/Select',
  component: Select,
  tags: ['autodocs'],
} satisfies Meta<typeof Select>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <Select defaultValue="smooth">
      <SelectTrigger className="w-56">
        <SelectValue placeholder="Animation style" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="minimal">Minimal</SelectItem>
        <SelectItem value="smooth">Smooth</SelectItem>
        <SelectItem value="playful">Playful</SelectItem>
      </SelectContent>
    </Select>
  ),
};
