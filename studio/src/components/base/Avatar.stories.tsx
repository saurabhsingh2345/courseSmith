import type {Meta, StoryObj} from '@storybook/react';
import {Avatar, AvatarFallback, AvatarImage} from './Avatar';

const meta = {
  title: 'Base/Avatar',
  component: Avatar,
  tags: ['autodocs'],
} satisfies Meta<typeof Avatar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const WithImage: Story = {
  render: () => (
    <Avatar>
      <AvatarImage src="https://i.pravatar.cc/80?img=12" alt="Instructor" />
      <AvatarFallback>IN</AvatarFallback>
    </Avatar>
  ),
};

export const FallbackOnly: Story = {
  render: () => (
    <Avatar>
      <AvatarFallback>CS</AvatarFallback>
    </Avatar>
  ),
};
