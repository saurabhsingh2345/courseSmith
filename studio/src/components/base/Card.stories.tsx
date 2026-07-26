import type {Meta, StoryObj} from '@storybook/react';
import {Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle} from './Card';
import {Button} from './Button';

const meta = {
  title: 'Base/Card',
  component: Card,
  tags: ['autodocs'],
} satisfies Meta<typeof Card>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <Card className="w-80">
      <CardHeader>
        <CardTitle>Lesson 3 · Recursion</CardTitle>
        <CardDescription>Estimated 12 minutes · 4 exercises</CardDescription>
      </CardHeader>
      <CardContent className="text-sm text-muted">
        Break a problem into a base case and a smaller subproblem, then let the call stack do the rest.
      </CardContent>
      <CardFooter className="gap-2">
        <Button size="sm">Start</Button>
        <Button size="sm" variant="ghost">
          Preview
        </Button>
      </CardFooter>
    </Card>
  ),
};
