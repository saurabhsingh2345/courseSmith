import type {Meta, StoryObj} from '@storybook/react';
import {Popover, PopoverContent, PopoverTrigger} from './Popover';
import {Button} from './Button';
import {Input} from './Input';
import {Label} from './Label';

const meta = {
  title: 'Base/Popover',
  component: Popover,
  tags: ['autodocs'],
} satisfies Meta<typeof Popover>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="secondary">Rename</Button>
      </PopoverTrigger>
      <PopoverContent>
        <div className="grid gap-2">
          <Label htmlFor="name">Course name</Label>
          <Input id="name" defaultValue="Intro to Python" />
        </div>
      </PopoverContent>
    </Popover>
  ),
};
