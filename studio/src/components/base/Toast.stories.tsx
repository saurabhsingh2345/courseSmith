import type {Meta, StoryObj} from '@storybook/react';
import {Toaster, toast} from './Toast';
import {Button} from './Button';

const meta = {
  title: 'Base/Toast',
  component: Toaster,
  tags: ['autodocs'],
} satisfies Meta<typeof Toaster>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <div>
      <div className="flex gap-2">
        <Button onClick={() => toast({title: 'Render complete', description: 'Lesson 3 is ready.', variant: 'success'})}>
          Success toast
        </Button>
        <Button
          variant="danger"
          onClick={() => toast({title: 'Render failed', description: 'Check the logs.', variant: 'error'})}
        >
          Error toast
        </Button>
      </div>
      <Toaster />
    </div>
  ),
};
