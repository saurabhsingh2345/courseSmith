import type {Meta, StoryObj} from '@storybook/react';
import {Tabs, TabsContent, TabsList, TabsTrigger} from './Tabs';

const meta = {
  title: 'Base/Tabs',
  component: Tabs,
  tags: ['autodocs'],
} satisfies Meta<typeof Tabs>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <Tabs defaultValue="script" className="w-96">
      <TabsList>
        <TabsTrigger value="script">Script</TabsTrigger>
        <TabsTrigger value="visuals">Visuals</TabsTrigger>
        <TabsTrigger value="quiz">Quiz</TabsTrigger>
      </TabsList>
      <TabsContent value="script" className="text-sm text-muted">
        The narrated lesson script lives here.
      </TabsContent>
      <TabsContent value="visuals" className="text-sm text-muted">
        Diagrams and animations for the lesson.
      </TabsContent>
      <TabsContent value="quiz" className="text-sm text-muted">
        Auto-generated comprehension questions.
      </TabsContent>
    </Tabs>
  ),
};
