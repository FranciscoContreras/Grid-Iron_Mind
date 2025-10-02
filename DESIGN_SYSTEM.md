# UI Design System Documentation

## System Name: **Frost** (Glassmorphic Design System)

**Version:** 2.0
**Last Updated:** October 2025
**CDN URL:** `https://nfl.wearemachina.com/ui-system.css`
**Documentation:** `https://nfl.wearemachina.com/ui-system.html`

---

## Table of Contents

1. [Philosophy & Design Language](#philosophy--design-language)
2. [Core Design Pattern](#core-design-pattern)
3. [Design Tokens](#design-tokens)
4. [Component Architecture](#component-architecture)
5. [Building New Components](#building-new-components)
6. [Expansion Guidelines](#expansion-guidelines)
7. [Best Practices](#best-practices)
8. [Common Patterns](#common-patterns)
9. [Accessibility](#accessibility)
10. [Browser Support](#browser-support)

---

## Philosophy & Design Language

### Visual Metaphor
The design system is built around the metaphor of **"UI elements cut from white glass on a pristine table"**. Components should appear:
- **Translucent** - Light passes through with subtle frosting
- **Minimal depth** - Just enough shadow/contrast to define edges
- **Soft tones** - No harsh colors, gentle gradients
- **Ethereal** - Elements float with subtle elevation

### Core Aesthetic Principles

1. **Glassmorphism First**
   - Every component uses frosted glass background with backdrop blur
   - Transparency creates visual hierarchy through layering
   - White gradient base maintains consistency

2. **Gradient Overlay System**
   - Purple → Blue → Orange signature gradient identifies interactive elements
   - Applied via `::before` pseudo-element with proper z-index stacking
   - Opacity varies by component type (0.08-0.3)

3. **Border Radius Standards**
   - **Pills (9999px):** Buttons, badges, form inputs, toggles, progress bars
   - **Rounded (24px / var(--radius-2xl)):** Cards, modals, alerts, tables, containers, textareas
   - Consistent rounding creates visual cohesion

4. **Soft Tones Palette**
   - White/off-white backgrounds (#fafafa)
   - Subtle borders (rgba(0,0,0,0.04-0.06))
   - Text: Dark gray (#2c2c2c) with lighter variants
   - No harsh black or pure white

---

## Core Design Pattern

Every glassmorphic component follows this exact structure:

```css
.component {
    /* Base Structure */
    position: relative;

    /* Glassmorphic Background */
    background: linear-gradient(135deg,
        rgba(255, 255, 255, 0.25) 0%,
        rgba(255, 255, 255, 0.15) 100%);
    backdrop-filter: blur(40px) saturate(200%);
    -webkit-backdrop-filter: blur(40px) saturate(200%);

    /* Border & Depth */
    border: 1px solid rgba(255, 255, 255, 0.3);
    border-radius: 9999px; /* or var(--radius-2xl) for containers */
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1),
                inset 0 1px 0 rgba(255, 255, 255, 0.4);

    /* Essential for Gradient Overlay */
    overflow: hidden;
    transition: all 0.3s ease;
}

/* Signature Gradient Overlay */
.component::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;

    /* Purple → Blue → Orange Gradient */
    background: linear-gradient(135deg,
        rgba(150, 130, 200, 0.12) 0%,    /* Purple */
        rgba(100, 150, 255, 0.12) 50%,   /* Blue */
        rgba(255, 200, 150, 0.08) 100%); /* Orange */

    opacity: 1;
    transition: opacity 0.3s ease;
    z-index: -1;
    border-radius: inherit;
    pointer-events: none;
}

/* Hover Enhancement */
.component:hover {
    background: linear-gradient(135deg,
        rgba(255, 255, 255, 0.3) 0%,
        rgba(255, 255, 255, 0.2) 100%);
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.15),
                inset 0 1px 0 rgba(255, 255, 255, 0.5);
    transform: translateY(-2px);
}

.component:hover::before {
    opacity: 1.3;
}

/* Content Protection (for containers) */
.component > * {
    position: relative;
    z-index: 1;
}
```

### Pattern Breakdown

**Why each line matters:**

1. **`position: relative`** - Required for absolute positioning of `::before` overlay
2. **`background: linear-gradient(...)`** - Creates white frosted glass base
3. **`backdrop-filter: blur(40px)`** - Blurs content behind component (glassmorphism)
4. **`-webkit-backdrop-filter`** - Safari/iOS compatibility
5. **`border: rgba(255,255,255,0.3)`** - Subtle white border defines edges
6. **`overflow: hidden`** - Clips gradient overlay to border-radius
7. **`::before` pseudo-element** - Overlay layer for purple/blue/orange gradient
8. **`z-index: -1` on ::before** - Places gradient behind content but above background
9. **`pointer-events: none` on ::before** - Prevents overlay from blocking interactions
10. **`border-radius: inherit`** - Matches parent's rounded corners
11. **`.component > * { z-index: 1 }`** - Ensures content sits above gradient overlay

---

## Design Tokens

### Color System

```css
:root {
    /* Base Colors - Soft Tones */
    --soft-white: #ffffff;
    --soft-bg: #fafafa;
    --soft-surface: rgba(255, 255, 255, 0.7);
    --soft-border: rgba(0, 0, 0, 0.06);
    --soft-border-subtle: rgba(0, 0, 0, 0.04);

    /* Text Hierarchy */
    --soft-text: #2c2c2c;           /* Primary text */
    --soft-text-light: #6b6b6b;     /* Secondary text */
    --soft-text-lighter: #9a9a9a;   /* Tertiary text */

    /* Signature Gradient Colors */
    --gradient-purple: rgba(150, 130, 200, x);
    --gradient-blue: rgba(100, 150, 255, x);
    --gradient-orange: rgba(255, 200, 150, x);

    /* Status Colors (Soft Versions) */
    --soft-success: rgba(100, 200, 100, x);
    --soft-warning: rgba(255, 200, 100, x);
    --soft-error: rgba(255, 100, 100, x);
}
```

**Opacity Scale Guide:**
- Background base: `0.15-0.25`
- Gradient overlays: `0.08-0.15` (default), `0.2-0.3` (emphasized)
- Borders: `0.18-0.4`
- Hover states: Increase all by ~0.05

### Spacing Scale

```css
--spacing-xs: 8px;
--spacing-sm: 12px;
--spacing-md: 16px;
--spacing-lg: 24px;
--spacing-xl: 32px;
--spacing-2xl: 48px;
```

### Border Radius Scale

```css
--radius-sm: 4px;    /* Rarely used */
--radius-md: 8px;    /* Rarely used */
--radius-lg: 12px;   /* Rarely used */
--radius-xl: 16px;   /* Rarely used */
--radius-2xl: 24px;  /* Standard for containers */
--radius-full: 9999px; /* Standard for interactive elements */
```

### Shadow System

```css
/* Subtle Depth - Resting State */
box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1),
            inset 0 1px 0 rgba(255, 255, 255, 0.4);

/* Enhanced Depth - Hover State */
box-shadow: 0 12px 40px rgba(0, 0, 0, 0.15),
            inset 0 1px 0 rgba(255, 255, 255, 0.5);

/* Maximum Depth - Modals/Overlays */
box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2),
            inset 0 1px 0 rgba(255, 255, 255, 0.6);
```

---

## Component Architecture

### Component Categories

1. **Interactive Elements** (pill-shaped, 9999px)
   - Buttons (`.btn`, `.btn-primary`, `.btn-glass`, `.btn-success`, `.btn-warning`, `.btn-error`)
   - Form inputs (`.form-input`, `.form-select`)
   - Badges (`.badge`)
   - Checkboxes/radios (`.form-checkbox`, `.form-radio`)
   - Toggle switches (`.toggle-slider`)
   - Progress bars (`.progress`, `.progress-bar`)

2. **Containers** (rounded, 24px)
   - Cards (`.card`, `.card-glass`, `.card-pricing`, `.card-stat`, `.card-tilt`)
   - Alerts (`.alert`, `.alert-success`, `.alert-warning`, `.alert-error`)
   - Modals (`.modal`, `.modal-overlay`)
   - Tables (`.table`)
   - Dropdowns (`.dropdown-menu`)
   - Accordions (`.accordion-header`)
   - Textareas (`.form-textarea`)
   - Navbar (`.navbar`)
   - Hero sections (`.hero`)

3. **Typography**
   - Gradient text (`.text-gradient`)
   - Headings (`.heading-xl`, `.heading-lg`, `.heading-md`, `.heading-sm`)

4. **Loading States**
   - Spinner (`.spinner`)
   - Skeleton (`.skeleton`, `.skeleton-pulse`)
   - Button loading (`.btn-loading`)

5. **Utility Components**
   - Tooltips (`.tooltip`, `.tooltip-text`)
   - Navigation tabs (`.nav-tabs`, `.nav-tab`)

---

## Building New Components

### Step-by-Step Guide

#### Step 1: Determine Component Type

**Ask yourself:**
- Is it interactive? → Use pill shape (9999px)
- Is it a container? → Use rounded (24px / var(--radius-2xl))
- Does it display data? → Likely a container
- Does it trigger actions? → Likely interactive

#### Step 2: Choose Base Structure

**For Interactive Elements:**
```css
.new-interactive-component {
    position: relative;
    display: inline-flex; /* or inline-block */
    align-items: center;
    justify-content: center;

    padding: 10px 24px; /* Adjust for size */
    font-size: 14px;
    font-weight: 400;

    border-radius: 9999px; /* Always pill-shaped */
    border: 1px solid rgba(255, 255, 255, 0.3);

    background: linear-gradient(135deg,
        rgba(255, 255, 255, 0.25) 0%,
        rgba(255, 255, 255, 0.15) 100%);
    backdrop-filter: blur(40px) saturate(200%);
    -webkit-backdrop-filter: blur(40px) saturate(200%);

    color: var(--soft-text);
    cursor: pointer;
    transition: all 0.3s ease;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1),
                inset 0 1px 0 rgba(255, 255, 255, 0.4);
    overflow: hidden;
}
```

**For Containers:**
```css
.new-container-component {
    position: relative;

    padding: var(--spacing-lg);

    border-radius: var(--radius-2xl); /* Always 24px */
    border: 1px solid rgba(255, 255, 255, 0.3);

    background: linear-gradient(135deg,
        rgba(255, 255, 255, 0.25) 0%,
        rgba(255, 255, 255, 0.15) 100%);
    backdrop-filter: blur(40px) saturate(200%);
    -webkit-backdrop-filter: blur(40px) saturate(200%);

    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1),
                inset 0 1px 0 rgba(255, 255, 255, 0.4);
    transition: all 0.3s ease;
    overflow: hidden;
}
```

#### Step 3: Add Gradient Overlay

**Always include the `::before` pseudo-element:**

```css
.new-component::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;

    background: linear-gradient(135deg,
        rgba(150, 130, 200, 0.12) 0%,    /* Purple */
        rgba(100, 150, 255, 0.12) 50%,   /* Blue */
        rgba(255, 200, 150, 0.08) 100%); /* Orange */

    opacity: 1;
    transition: opacity 0.3s ease;
    z-index: -1;
    border-radius: inherit;
    pointer-events: none;
}
```

**Adjust gradient opacity based on component type:**
- Default components: `0.12 / 0.12 / 0.08`
- Emphasized (primary): `0.2 / 0.2 / 0.15`
- Subtle (cards): `0.08 / 0.08 / 0.05`

#### Step 4: Add Hover State

```css
.new-component:hover {
    background: linear-gradient(135deg,
        rgba(255, 255, 255, 0.3) 0%,
        rgba(255, 255, 255, 0.2) 100%);
    border-color: rgba(255, 255, 255, 0.4);
    transform: translateY(-2px); /* Subtle lift */
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.15),
                inset 0 1px 0 rgba(255, 255, 255, 0.5);
}

.new-component:hover::before {
    opacity: 1.3; /* Intensify gradient */
}
```

#### Step 5: Protect Content (Containers Only)

**For components with text/images inside:**

```css
.new-component > * {
    position: relative;
    z-index: 1;
}
```

This ensures content sits above the gradient overlay.

#### Step 6: Add Variants (Optional)

**Status color variants:**

```css
.new-component-success::before {
    background: linear-gradient(135deg,
        rgba(100, 200, 100, 0.2) 0%,
        rgba(100, 200, 150, 0.2) 50%,
        rgba(100, 220, 120, 0.15) 100%);
}

.new-component-warning::before {
    background: linear-gradient(135deg,
        rgba(255, 200, 100, 0.2) 0%,
        rgba(255, 180, 100, 0.2) 50%,
        rgba(255, 190, 80, 0.15) 100%);
}

.new-component-error::before {
    background: linear-gradient(135deg,
        rgba(255, 100, 100, 0.2) 0%,
        rgba(255, 120, 120, 0.2) 50%,
        rgba(255, 110, 110, 0.15) 100%);
}
```

#### Step 7: Add States

**Disabled state:**
```css
.new-component:disabled,
.new-component[disabled] {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
    pointer-events: none;
}
```

**Active/pressed state:**
```css
.new-component:active {
    transform: translateY(0);
    background: linear-gradient(135deg,
        rgba(255, 255, 255, 0.2) 0%,
        rgba(255, 255, 255, 0.1) 100%);
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12),
                inset 0 1px 0 rgba(255, 255, 255, 0.3);
}
```

**Focus state (for accessibility):**
```css
.new-component:focus {
    outline: 2px solid rgba(100, 150, 255, 0.3);
    outline-offset: 2px;
}

.new-component:focus-visible {
    outline: 2px solid rgba(100, 150, 255, 0.5);
    outline-offset: 2px;
}
```

---

## Expansion Guidelines

### Adding New Component Families

When adding a completely new component type (e.g., breadcrumbs, stepper, tabs v2):

1. **Research existing patterns** - Check if similar components exist
2. **Determine category** - Interactive vs Container
3. **Follow core pattern** - Always use glassmorphic base + gradient overlay
4. **Maintain consistency** - Use design tokens, standard spacing, border-radius rules
5. **Add variants** - Consider size variants (.sm, .lg), status variants (.success, .warning, .error)
6. **Test accessibility** - Keyboard navigation, screen readers, focus states
7. **Document usage** - Add examples to ui-system.html

### Extending Existing Components

**To add a new button variant:**

```css
/* Follow the established pattern */
.btn-new-variant {
    /* Inherit base .btn styles */
    background: linear-gradient(135deg,
        rgba(255, 255, 255, 0.25) 0%,
        rgba(255, 255, 255, 0.15) 100%);
}

.btn-new-variant::before {
    /* Custom gradient for this variant */
    background: linear-gradient(135deg,
        rgba(R, G, B, 0.2) 0%,
        rgba(R, G, B, 0.2) 50%,
        rgba(R, G, B, 0.15) 100%);
}

.btn-new-variant:hover::before {
    opacity: 1.5;
}
```

**To add a new card variant:**

```css
.card-new-variant {
    /* Inherits .card base styles */
}

.card-new-variant::before {
    /* Adjust gradient if needed */
    background: linear-gradient(135deg,
        rgba(150, 130, 200, 0.1) 0%,
        rgba(100, 150, 255, 0.1) 50%,
        rgba(255, 200, 150, 0.06) 100%);
}

/* Add specific child elements */
.card-new-variant-header {
    /* Custom styling */
}
```

### Creating Component Modifiers

**Size modifiers:**

```css
.component-sm {
    padding: 6px 16px;
    font-size: 12px;
}

.component-lg {
    padding: 14px 32px;
    font-size: 16px;
}
```

**State modifiers:**

```css
.component.is-active {
    /* Enhanced gradient */
}

.component.is-loading {
    color: transparent;
    pointer-events: none;
}

.component.is-valid {
    border-color: rgba(100, 200, 100, 0.4);
}

.component.is-invalid {
    border-color: rgba(255, 100, 100, 0.4);
}
```

---

## Best Practices

### DO ✅

1. **Always use the core pattern** - Every component gets glassmorphic background + gradient overlay
2. **Respect border-radius rules** - Pills for interactive, 24px for containers
3. **Use design tokens** - Reference CSS custom properties for consistency
4. **Add `overflow: hidden`** - Required for gradient overlay clipping
5. **Include `-webkit-backdrop-filter`** - Safari compatibility
6. **Set `pointer-events: none` on ::before** - Prevents interaction blocking
7. **Use `z-index: 1` on content** - Ensures text/images sit above overlays
8. **Add hover states** - Slight lift (translateY), enhanced shadows, intensified gradients
9. **Include accessibility** - Focus states, ARIA attributes, keyboard support
10. **Test on Safari/iOS** - Glassmorphism can be buggy on Apple devices

### DON'T ❌

1. **Don't use solid backgrounds** - Breaks glassmorphic aesthetic
2. **Don't skip the gradient overlay** - Signature purple/blue/orange is brand identity
3. **Don't mix border-radius styles** - Interactive = 9999px, containers = 24px, never arbitrary
4. **Don't use harsh colors** - System is based on soft tones
5. **Don't forget `border-radius: inherit` on ::before** - Creates visual bugs
6. **Don't use excessive box-shadows** - Minimal depth is key
7. **Don't skip transitions** - All state changes should be smooth (0.3s ease)
8. **Don't add gradient overlay to non-glassmorphic elements** - It's a cohesive system
9. **Don't use `!important`** - Proper specificity and cascading
10. **Don't create arbitrary opacity values** - Use established scale (0.05 increments)

### Common Mistakes to Avoid

**❌ Incorrect z-index stacking:**
```css
/* WRONG - Content will appear behind gradient */
.component::before {
    z-index: 1;
}
```

```css
/* CORRECT - Gradient behind, content in front */
.component::before {
    z-index: -1;
}
.component > * {
    position: relative;
    z-index: 1;
}
```

**❌ Missing overflow:**
```css
/* WRONG - Gradient overlay bleeds outside rounded corners */
.component {
    border-radius: 24px;
    /* Missing: overflow: hidden; */
}
```

**❌ Forgetting Safari prefix:**
```css
/* WRONG - Won't work on Safari/iOS */
.component {
    backdrop-filter: blur(40px);
}
```

```css
/* CORRECT - Cross-browser support */
.component {
    backdrop-filter: blur(40px) saturate(200%);
    -webkit-backdrop-filter: blur(40px) saturate(200%);
}
```

---

## Common Patterns

### Pattern 1: Interactive Button

```css
.btn-custom {
    position: relative;
    display: inline-flex;
    align-items: center;
    padding: 10px 24px;
    font-size: 14px;
    border-radius: 9999px;
    border: 1px solid rgba(255, 255, 255, 0.3);
    background: linear-gradient(135deg, rgba(255, 255, 255, 0.25) 0%, rgba(255, 255, 255, 0.15) 100%);
    backdrop-filter: blur(40px) saturate(200%);
    -webkit-backdrop-filter: blur(40px) saturate(200%);
    cursor: pointer;
    transition: all 0.3s ease;
    overflow: hidden;
}

.btn-custom::before {
    content: '';
    position: absolute;
    inset: 0;
    background: linear-gradient(135deg, rgba(150, 130, 200, 0.12) 0%, rgba(100, 150, 255, 0.12) 50%, rgba(255, 200, 150, 0.08) 100%);
    z-index: -1;
    border-radius: inherit;
    pointer-events: none;
}

.btn-custom:hover {
    transform: translateY(-2px);
}

.btn-custom:hover::before {
    opacity: 1.3;
}
```

### Pattern 2: Container with Content

```css
.container-custom {
    position: relative;
    padding: var(--spacing-lg);
    border-radius: var(--radius-2xl);
    border: 1px solid rgba(255, 255, 255, 0.3);
    background: linear-gradient(135deg, rgba(255, 255, 255, 0.25) 0%, rgba(255, 255, 255, 0.15) 100%);
    backdrop-filter: blur(40px) saturate(200%);
    -webkit-backdrop-filter: blur(40px) saturate(200%);
    overflow: hidden;
}

.container-custom::before {
    content: '';
    position: absolute;
    inset: 0;
    background: linear-gradient(135deg, rgba(150, 130, 200, 0.08) 0%, rgba(100, 150, 255, 0.08) 50%, rgba(255, 200, 150, 0.05) 100%);
    z-index: 0;
    border-radius: inherit;
    pointer-events: none;
}

.container-custom > * {
    position: relative;
    z-index: 1;
}
```

### Pattern 3: Status Color Variant

```css
.status-component {
    /* Base glassmorphic styles */
}

.status-component-success::before {
    background: linear-gradient(135deg, rgba(100, 200, 100, 0.2) 0%, rgba(100, 200, 150, 0.2) 50%, rgba(100, 220, 120, 0.15) 100%);
}

.status-component-warning::before {
    background: linear-gradient(135deg, rgba(255, 200, 100, 0.2) 0%, rgba(255, 180, 100, 0.2) 50%, rgba(255, 190, 80, 0.15) 100%);
}

.status-component-error::before {
    background: linear-gradient(135deg, rgba(255, 100, 100, 0.2) 0%, rgba(255, 120, 120, 0.2) 50%, rgba(255, 110, 110, 0.15) 100%);
}
```

### Pattern 4: Animated Loader

```css
.loader {
    position: relative;
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background: linear-gradient(135deg, rgba(255, 255, 255, 0.25) 0%, rgba(255, 255, 255, 0.15) 100%);
    backdrop-filter: blur(40px) saturate(200%);
    -webkit-backdrop-filter: blur(40px) saturate(200%);
    overflow: hidden;
}

.loader::before {
    content: '';
    position: absolute;
    inset: 0;
    background: linear-gradient(135deg, rgba(150, 130, 200, 0.4) 0%, rgba(100, 150, 255, 0.4) 50%, rgba(255, 200, 150, 0.3) 100%);
    clip-path: polygon(50% 50%, 50% 0%, 100% 0%, 100% 100%, 0% 100%, 0% 0%, 50% 0%);
    animation: spin 0.8s linear infinite;
}

@keyframes spin {
    to { transform: rotate(360deg); }
}
```

---

## Accessibility

### Focus Management

Always include visible focus states for keyboard navigation:

```css
.interactive-element:focus {
    outline: 2px solid rgba(100, 150, 255, 0.3);
    outline-offset: 2px;
}

.interactive-element:focus-visible {
    outline: 2px solid rgba(100, 150, 255, 0.5);
    outline-offset: 2px;
}
```

### Disabled States

Clearly indicate disabled components:

```css
.component:disabled,
.component[disabled] {
    opacity: 0.5;
    cursor: not-allowed;
    pointer-events: none;
}
```

### ARIA Support

Add appropriate ARIA attributes in HTML:

```html
<button class="btn" aria-label="Submit form">Submit</button>
<div class="alert" role="alert">Important message</div>
<nav class="navbar" role="navigation" aria-label="Main navigation"></nav>
```

### Color Contrast

- Text color `#2c2c2c` on white backgrounds meets WCAG AA (4.5:1)
- Light text `#6b6b6b` meets WCAG AA for large text
- Status colors maintain sufficient contrast when used with appropriate text colors

---

## Browser Support

### Supported Browsers

- **Chrome/Edge:** 88+ (full support)
- **Firefox:** 103+ (full support)
- **Safari:** 15.4+ (full support with `-webkit-` prefix)
- **iOS Safari:** 15.4+
- **Opera:** 74+

### Fallbacks

For browsers without backdrop-filter support:

```css
@supports not (backdrop-filter: blur(40px)) {
    .component {
        background: rgba(255, 255, 255, 0.95);
        /* Remove blur, increase opacity for readability */
    }
}
```

### Testing Checklist

When adding new components, test on:
- [ ] Chrome (latest)
- [ ] Firefox (latest)
- [ ] Safari (macOS)
- [ ] Safari (iOS)
- [ ] Edge (latest)
- [ ] Mobile viewport (320px - 768px)
- [ ] Tablet viewport (768px - 1024px)
- [ ] Desktop viewport (1024px+)

---

## Usage Examples

### Import in HTML

```html
<link rel="stylesheet" href="https://nfl.wearemachina.com/ui-system.css">
```

### Import in CSS

```css
@import url('https://nfl.wearemachina.com/ui-system.css');
```

### Basic Button

```html
<button class="btn btn-primary">Click Me</button>
<button class="btn btn-success">Success</button>
<button class="btn btn-glass">Glassmorphic</button>
```

### Card with Content

```html
<div class="card">
    <div class="card-header">Card Title</div>
    <div class="card-body">
        Card content goes here with automatic gradient overlay.
    </div>
    <div class="card-footer">
        <button class="btn btn-primary">Action</button>
    </div>
</div>
```

### Form with Validation

```html
<div class="form-group">
    <label class="form-label">Email</label>
    <input type="email" class="form-input is-valid" placeholder="you@example.com">
</div>

<div class="form-group">
    <label class="form-label">Password</label>
    <input type="password" class="form-input is-invalid" placeholder="Enter password">
</div>
```

---

## Version History

### v2.0 (October 2025)
- Complete glassmorphic redesign
- Added signature purple/blue/orange gradient system
- Implemented Phase 4 advanced components
- Standardized border-radius (pills vs containers)
- Added loading states, tooltips, toggles, progress bars, dropdowns, accordions
- Enhanced form features (floating labels, validation states)
- Advanced card variants (pricing, stat, tilt)
- Layout components (navbar, hero)

### v1.0 (Initial Release)
- Basic component library with soft tones
- Minimal depth aesthetic
- Foundation design tokens

---

## Credits & License

**Design System:** Frost (Glassmorphic Design System)
**Created by:** Grid Iron Mind Team
**License:** Proprietary (for Grid Iron Mind project use)
**CDN Host:** nfl.wearemachina.com

For questions or contributions, refer to project documentation or contact the development team.

---

## Quick Reference Card

### When Building Components, Always Remember:

1. ✅ **Use the core pattern** (glassmorphic base + gradient overlay)
2. ✅ **Pills (9999px) for interactive**, **24px for containers**
3. ✅ **Purple → Blue → Orange gradient on ::before**
4. ✅ **`overflow: hidden`** required
5. ✅ **`-webkit-backdrop-filter`** for Safari
6. ✅ **`z-index: -1` on ::before**, **z-index: 1 on content**
7. ✅ **Hover = lift + intensify gradient**
8. ✅ **Focus states for accessibility**
9. ✅ **Test on Safari/iOS**
10. ✅ **Document in ui-system.html**
