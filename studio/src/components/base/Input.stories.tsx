import type {Meta, StoryObj} from '@storybook/react';
import {Input} from './Input';
import {Label} from './Label';

const meta = {
  title: 'Base/Input',
  component: Input,
  tags: ['autodocs'],
  args: {placeholder: 'you@example.com'},
} satisfies Meta<typeof Input>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Disabled: Story = {args: {disabled: true, value: 'Read only'}};

export const WithLabel: Story = {
  render: (args) => (
    <div className="grid w-72 gap-2">
      <Label htmlFor="email">Email</Label>
      <Input id="email" {...args} />
    </div>
  ),
};
