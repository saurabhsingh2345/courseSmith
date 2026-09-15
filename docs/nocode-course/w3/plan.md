# CONTROL TOWER

## What it is

A live operations console for a delivery fleet. One screen, always moving, that
a dispatcher would keep open all day.

## Why it is built this way

- Everything on the screen moves, so the state of the operation is readable at a
  glance rather than by reading rows.
- All data comes from a simulator, so it runs offline, costs nothing, and holds
  no real customer information.
- The parts are the ordinary parts of any live console — a feed, a list, a map,
  a heat map, a chart and a chat box. Nothing here is exotic.

## The screen

```
┌───────────────┬────────────────────────────────┬──────────────────┐
│ FLEET         │ MAP                            │ REGION HEAT      │
│ 60 vehicles   │ vans moving along routes       │ box = value      │
│ grouped by    │ live, ~1s tick                 │ colour = risk    │
│ lane          │                                ├──────────────────┤
│ status dot    ├────────────────────────────────┤ ON-TIME TODAY    │
│ ETA counting  │ EXCEPTIONS TICKER              │ line chart       │
│ down          │ breakdown · weather · customs  │                  │
├───────────────┴────────────────────────────────┴──────────────────┤
│ DISPATCHER  ›  reroute VAN-14 around the closure                  │
└───────────────────────────────────────────────────────────────────┘
```

## Must have

### 1. Simulator
- 60 vehicles, ids `VAN-01` … `VAN-60`.
- Six regions: North, Midlands, South East, South West, Wales, Scotland.
- Each vehicle: current position, route, speed, load value, ETA, status of
  `rolling` · `delayed` · `stopped` · `delivered`.
- Ticks **once a second**. Positions advance, ETAs recompute, status changes.
- Randomly injects exceptions: breakdown, weather hold, customs, traffic.
- **Deterministic seed option** so a demo can be replayed. This matters for
  filming — a retake must show the same world.

### 2. Fleet list (left)
- All 60, grouped by lane, scrollable.
- Coloured status dot, vehicle id, destination, **ETA counting down live**.
- Late vehicles sort to the top.
- Clicking one selects it everywhere else on the screen.

### 3. Map (centre)
- The six regions drawn simply — **no map tiles, no external images**, just
  regions as shapes. It must work with the network unplugged.
- Vehicles as dots moving along their routes.
- Selected vehicle highlighted; its route drawn.
- Delayed vehicles pulse.

### 4. Region heat map (right)
- One box per region. **Box size = total load value in transit. Colour = delay
  risk**, green through amber to red.
- Updates as the simulator ticks.
- Hovering shows the numbers behind the box.

### 5. Exceptions ticker (under the map)
- Newest first, timestamped, colour-coded by kind.
- Caps at 50 rows so it never grows without bound.

### 6. On-time chart (right, under the heat map)
- On-time percentage across the working day, line chart, live.

### 7. Dispatcher chat (bottom)
- Free text in, and it can actually **do** things, not just answer:
  - `reroute VAN-14 around the closure`
  - `which lanes are at risk this afternoon?`
  - `hold the Leeds shipment and tell me why`
  - `what is our on-time rate right now?`
- An action changes the simulator and the change shows on screen immediately.
- Answers ground themselves in current state — never invented numbers.

### 8. Light and dark
- Dark by default. A toggle in the top bar. **Both must look deliberate**, not
  an inverted dark theme.

### 9. Packaging
- Runs in Docker, `docker compose up`.
- `scripts/start.sh` and `scripts/stop.sh`.
- Serves on **:8080**.
- `.env.example` committed, **`.env` never committed**.

### 10. Tests
- Unit tests for the simulator, the risk calculation and the chat actions.
- One end-to-end test that starts the app, buys a change through the chat and
  asks the API for the result.
- **Tests must survive a reimplementation and break on a logic error.** Do not
  chase a coverage number and do not mock everything — a suite of mocks that
  proves nothing is worse than fewer honest tests.

## Explicitly out of scope for v1

User accounts and login · real map tiles · a real telematics feed ·
mobile layout · anything that needs a paid API key.

## Non-negotiable

- **No personal data anywhere** — no real names, emails, companies or addresses.
  Destinations are cities. Drivers are `DRV-07`, not people.
- **No emojis in the UI or the code.**
- Long files get split. Over-defensive code gets rejected.
- Every dependency must be one an offline machine already has, or must be
  installable in one command.

## The one-line brief

> Build a live delivery-fleet operations console: 60 simulated vans on a map,
> a fleet list with live ETAs, a region heat map sized by value and coloured by
> delay risk, an exceptions ticker, an on-time chart, and a chat box that can
> reroute or hold a vehicle and answer questions about current state. Dark and
> light. Docker. Tests. Everything simulated — no external APIs.
