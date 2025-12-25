You are Gemini 3 Pro acting as a senior mobile UI/UX designer.

Your task is to DESIGN HIGH-FIDELITY MOBILE UI SCREENS
for a Progressive Web App (PWA) called:

"Oreo Tennis Scoring (OTS)"

This is a real, buildable product. Do NOT invent features or flows.
Strictly follow the provided requirements.

────────────────────────────────
PRODUCT CONTEXT
────────────────────────────────
OTS is a lightweight tennis scoring app for a closed group of players.

Key characteristics:
- Mobile-first
- Installable PWA (Chrome “Add to Home Screen”)
- Used during live tennis matches
- One-handed usage
- Outdoor usage (sunlight, sweat, distraction)
- Offline-first during live matches

There are two roles:
1. Admin (rarely used, simple screens)
2. Players (no login, primary users)

────────────────────────────────
DESIGN GOALS
────────────────────────────────
- Extremely clear and readable UI
- Large tap targets (thumb-friendly)
- Minimal cognitive load
- Dark mode by default
- Elegant but restrained (no flashy gradients, no gamification)

Think:
"Professional sports tool" NOT "social fitness app"

────────────────────────────────
GLOBAL DESIGN RULES
────────────────────────────────
- Mobile frame: 390 × 844
- Dark theme by default
- No scrolling on live match screen
- No text input during live match
- Max 1 primary action per screen
- Buttons must be usable with sweaty fingers

────────────────────────────────
DESIGN TOKENS (MANDATORY)
────────────────────────────────

Colors:
- Background primary: #0F172A
- Background secondary: #111827
- Surface: #1F2933
- Text primary: #E5E7EB
- Text secondary: #9CA3AF
- Accent (tennis green): #22C55E
- Danger (double fault): #EF4444

Typography:
- Font: Inter (or clean system sans-serif)
- Title: 22px, semibold
- Section header: 16px, semibold
- Body: 14px, regular
- Button text: 16px, semibold

Spacing system:
- 4px / 8px / 16px / 24px / 32px
Use spacing consistently.

Corner radius:
- Buttons: 12px
- Cards: 16px

────────────────────────────────
SCREENS TO DESIGN
────────────────────────────────

Design ALL of the following screens as separate mobile layouts.

1️⃣ HOME SCREEN
Purpose: Immediate entry

Elements:
- App title: "OTS – Oreo Tennis Scoring"
- Primary button: "Start New Match"
- Secondary card (only if match exists): "Resume Match"

Layout:
- Centered primary CTA
- Minimal text
- Calm, confident look

────────────────────────────────

2️⃣ MATCH SETUP – VENUE & TYPE

Elements:
- Header: "New Match"
- Venue selector (dropdown-style)
- Match type selector:
  - Singles (selected)
  - Doubles
- Primary button: "Continue"

Rules:
- No typing
- Simple selection UI
- Large touch targets

────────────────────────────────

3️⃣ PLAYER SELECTION

Singles:
- Team A: 1 player selector
- Team B: 1 player selector

Doubles:
- Team A: 2 player selectors
- Team B: 2 player selectors

Rules:
- Prevent duplicate player selection
- Use dropdown/list modal UI
- Clear team separation

Primary CTA:
- "Start Match"

────────────────────────────────

4️⃣ LIVE MATCH SCORING (MOST IMPORTANT)

This screen must feel UNBREAKABLE.

Elements:
- Venue name (top)
- Score display (Team A vs Team B)
- Server indicator (with subtle switch option)
- FIVE LARGE ACTION BUTTONS:
  - First Serve – Won
  - First Serve – Lost
  - Second Serve – Won
  - Second Serve – Lost
  - Double Fault (danger color)

Rules:
- No scrolling
- Buttons fill width
- One tap = one point
- Clear visual hierarchy
- No charts, no menus, no distractions

This screen is used mid-rally. Design accordingly.

────────────────────────────────

5️⃣ MATCH SUMMARY

Purpose: Post-match insight

Elements:
- Match info (venue, match type)
- Serve statistics per player:
  - First serve in %
  - First serve points won %
  - Second serve points won %
  - Double faults
- Simple bar or percentage visuals
- Button: "Finish"

Rules:
- Charts are simple
- No animations
- Clarity over beauty

────────────────────────────────

6️⃣ ADMIN LOGIN

Elements:
- Title: "Admin Login"
- Username field
- Password field
- Login button

Rules:
- No branding flair
- Neutral, utilitarian look
- Clear error state (but generic messaging)

────────────────────────────────

7️⃣ ADMIN DASHBOARD

Elements:
- Simple navigation cards:
  - Players
  - Venues
  - Matches
- Logout button

Rules:
- Functional, not decorative
- Minimal layout

────────────────────────────────

8️⃣ ADMIN – PLAYERS MANAGEMENT

Elements:
- List of players
- Add Player button
- Enable / Disable toggle per player

Rules:
- No complex forms
- Clean list-based UI

────────────────────────────────

9️⃣ ADMIN – VENUES MANAGEMENT

Elements:
- List of venues
- Venue name + surface type
- Add Venue button

────────────────────────────────
IMPORTANT CONSTRAINTS
────────────────────────────────
- DO NOT invent new features
- DO NOT add social elements
- DO NOT add animations or illustrations
- DO NOT add onboarding flows
- DO NOT change screen flow
- DO NOT add extra buttons

Focus on clarity, spacing, hierarchy, and calm confidence.

────────────────────────────────
OUTPUT FORMAT
────────────────────────────────
Provide:
- High-fidelity mobile UI designs
- Clear visual hierarchy
- Consistent component usage
- Design that can be directly translated to Svelte components

You are designing for builders, not marketers.
