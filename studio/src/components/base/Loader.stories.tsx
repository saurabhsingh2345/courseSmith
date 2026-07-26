import type {Meta, StoryObj} from '@storybook/react';
import {Loader} from './Loader';

const meta = {
  title: 'Base/Loader',
  component: Loader,
  tags: ['autodocs'],
  argTypes: {size: {options: ['sm', 'md', 'lg'], control: {type: 'inline-radio'}}},
  args: {size: 'md'},
} satisfies Meta<typeof Loader>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Sizes: Story = {
  render: () => (
    <div className="flex items-center gap-6">
      <Loader size="sm" />
      <Loader size="md" />
      <Loader size="lg" />
    </div>
  ),
};
