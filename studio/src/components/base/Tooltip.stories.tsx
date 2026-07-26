import type {Meta, StoryObj} from '@storybook/react';
import {Tooltip, TooltipContent, TooltipProvider, TooltipTrigger} from './Tooltip';
import {Button} from './Button';

const meta = {
  title: 'Base/Tooltip',
  component: Tooltip,
  tags: ['autodocs'],
} satisfies Meta<typeof Tooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button variant="secondary">Hover me</Button>
        </TooltipTrigger>
        <TooltipContent>Re-renders the current lesson</TooltipContent>
      </Tooltip>
    </TooltipProvider>
  ),
};
