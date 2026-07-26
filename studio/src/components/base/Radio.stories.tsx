import type {Meta, StoryObj} from '@storybook/react';
import {RadioGroup, RadioGroupItem} from './Radio';
import {Label} from './Label';

const meta = {
  title: 'Base/RadioGroup',
  component: RadioGroup,
  tags: ['autodocs'],
} satisfies Meta<typeof RadioGroup>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <RadioGroup defaultValue="smooth">
      {['minimal', 'smooth', 'playful'].map((v) => (
        <div key={v} className="flex items-center gap-2">
          <RadioGroupItem value={v} id={v} />
          <Label htmlFor={v} className="capitalize">
            {v}
          </Label>
        </div>
      ))}
    </RadioGroup>
  ),
};
