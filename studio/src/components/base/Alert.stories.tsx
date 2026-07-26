import type {Meta, StoryObj} from '@storybook/react';
import {Info} from 'lucide-react';
import {Alert, AlertDescription, AlertTitle} from './Alert';

const meta = {
  title: 'Base/Alert',
  component: Alert,
  tags: ['autodocs'],
  argTypes: {
    variant: {options: ['default', 'info', 'success', 'warning', 'error'], control: {type: 'select'}},
  },
} satisfies Meta<typeof Alert>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {variant: 'info'},
  render: (args) => (
    <Alert {...args} className="max-w-md">
      <Info />
      <AlertTitle>Render queued</AlertTitle>
      <AlertDescription>Your course video will be ready in a few minutes.</AlertDescription>
    </Alert>
  ),
};

export const AllVariants: Story = {
  render: () => (
    <div className="grid max-w-md gap-3">
      {(['default', 'info', 'success', 'warning', 'error'] as const).map((v) => (
        <Alert key={v} variant={v}>
          <AlertTitle className="capitalize">{v}</AlertTitle>
          <AlertDescription>This is a {v} alert.</AlertDescription>
        </Alert>
      ))}
    </div>
  ),
};
