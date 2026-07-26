import type {Meta, StoryObj} from '@storybook/react';
import {Slider} from './Slider';

const meta = {
  title: 'Base/Slider',
  component: Slider,
  tags: ['autodocs'],
} satisfies Meta<typeof Slider>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => <Slider defaultValue={[40]} max={100} step={1} className="w-72" />,
};

export const Range: Story = {
  render: () => <Slider defaultValue={[25, 75]} max={100} step={1} className="w-72" />,
};
