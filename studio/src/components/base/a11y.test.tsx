// @vitest-environment jsdom
// Axe smoke test over the base library. In jsdom, layout-dependent rules
// (colour-contrast) come back "incomplete" rather than failing — so this
// primarily guards the ARIA surface: accessible names, roles, and labels.
import type {ReactElement} from 'react';
import {afterEach, describe, expect, it} from 'vitest';
import {cleanup, render} from '@testing-library/react';
import {axe} from 'vitest-axe';
import * as axeMatchers from 'vitest-axe/matchers';

import {Alert, AlertDescription, AlertTitle} from './Alert';
import {Avatar, AvatarFallback} from './Avatar';
import {Badge} from './Badge';
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from './Breadcrumb';
import {Button} from './Button';
import {Card, CardContent, CardHeader, CardTitle} from './Card';
import {Checkbox} from './Checkbox';
import {Input} from './Input';
import {Label} from './Label';
import {Loader} from './Loader';
import {RadioGroup, RadioGroupItem} from './Radio';
import {Select, SelectTrigger, SelectValue} from './Select';
import {Skeleton} from './Skeleton';
import {Slider} from './Slider';
import {Tabs, TabsContent, TabsList, TabsTrigger} from './Tabs';
import {Textarea} from './Textarea';

expect.extend(axeMatchers);
afterEach(cleanup);

// Radix's Slider measures its track via ResizeObserver, which jsdom doesn't
// implement. A no-op stub is enough to let it mount for the ARIA audit.
if (!('ResizeObserver' in globalThis)) {
  globalThis.ResizeObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
  } as unknown as typeof ResizeObserver;
}

const cases: Array<[string, ReactElement]> = [
  ['Button', <Button>Save</Button>],
  ['Badge', <Badge>New</Badge>],
  [
    'Input+Label',
    <div>
      <Label htmlFor="a11y-email">Email</Label>
      <Input id="a11y-email" placeholder="you@example.com" />
    </div>,
  ],
  [
    'Textarea+Label',
    <div>
      <Label htmlFor="a11y-notes">Notes</Label>
      <Textarea id="a11y-notes" />
    </div>,
  ],
  [
    'Checkbox+Label',
    <div>
      <Checkbox id="a11y-cb" />
      <Label htmlFor="a11y-cb">Accept</Label>
    </div>,
  ],
  [
    'RadioGroup',
    <RadioGroup defaultValue="one" aria-label="Pick one">
      <RadioGroupItem value="one" aria-label="One" />
      <RadioGroupItem value="two" aria-label="Two" />
    </RadioGroup>,
  ],
  [
    'Alert',
    <Alert variant="info">
      <AlertTitle>Heads up</AlertTitle>
      <AlertDescription>Something happened.</AlertDescription>
    </Alert>,
  ],
  [
    'Card',
    <Card>
      <CardHeader>
        <CardTitle>Title</CardTitle>
      </CardHeader>
      <CardContent>Body</CardContent>
    </Card>,
  ],
  ['Loader', <Loader label="Loading course" />],
  ['Skeleton', <Skeleton className="h-4 w-32" aria-hidden="true" />],
  [
    'Avatar',
    <Avatar>
      <AvatarFallback>CS</AvatarFallback>
    </Avatar>,
  ],
  [
    'Breadcrumb',
    <Breadcrumb>
      <BreadcrumbList>
        <BreadcrumbItem>
          <BreadcrumbLink href="#">Courses</BreadcrumbLink>
        </BreadcrumbItem>
        <BreadcrumbSeparator />
        <BreadcrumbItem>
          <BreadcrumbPage>Recursion</BreadcrumbPage>
        </BreadcrumbItem>
      </BreadcrumbList>
    </Breadcrumb>,
  ],
  ['Slider', <Slider defaultValue={[40]} max={100} aria-label="Playback speed" />],
  [
    // Portal-backed components (Select/Dialog/Popover/Tooltip/Dropdown) render
    // their overlay only once opened; here we cover the always-present trigger,
    // which must expose an accessible name on its own.
    'Select trigger',
    <Select defaultValue="smooth">
      <SelectTrigger aria-label="Animation style" className="w-56">
        <SelectValue placeholder="Animation style" />
      </SelectTrigger>
    </Select>,
  ],
  [
    'Tabs',
    <Tabs defaultValue="a">
      <TabsList>
        <TabsTrigger value="a">A</TabsTrigger>
        <TabsTrigger value="b">B</TabsTrigger>
      </TabsList>
      <TabsContent value="a">Panel A</TabsContent>
      <TabsContent value="b">Panel B</TabsContent>
    </Tabs>,
  ],
];

describe('base components a11y', () => {
  it.each(cases)('%s has no axe violations', async (_name, ui) => {
    const {container} = render(ui);
    // colour-contrast needs real layout/canvas, which jsdom lacks — disable it
    // here; contrast is covered visually in Storybook via addon-a11y.
    const results = await axe(container, {rules: {'color-contrast': {enabled: false}}});
    expect(results).toHaveNoViolations();
  });
});
