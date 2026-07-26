import type {Meta, StoryObj} from '@storybook/react';
import {Checkbox} from './Checkbox';
import {Label} from './Label';

const meta = {
  title: 'Base/Checkbox',
  component: Checkbox,
  tags: ['autodocs'],
} satisfies Meta<typeof Checkbox>;

export default meta;
type Story = StoryObj<typeof meta>;

export const WithLabel: Story = {
  render: () => (
    <div className="flex items-center gap-2">
      <Checkbox id="captions" defaultChecked />
      <Label htmlFor="captions">Burn in captions</Label>
    </div>
  ),
};
