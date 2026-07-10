# Dark/Light Theme Configuration

This project supports automatic dark/light theme switching with user customization and system preference detection.

## How It Works

- **Theme Store** (`src/lib/stores/theme.ts`): Centralized Svelte store for managing theme state
- **CSS Variables** (`src/routes/layout.css`): Dynamic color values that change based on theme
- **Tailwind Dark Mode** (`tailwind.config.ts`): Class-based dark mode with Tailwind CSS
- **Hooks** (`src/hooks.client.ts`): Prevents theme flash on page load by applying theme before rendering
- **localStorage**: Persists user's theme preference across sessions
- **System Preference**: Falls back to OS theme preference if no saved preference

## Theme State Priority

1. **localStorage** - User's saved preference (highest priority)
2. **System preference** - OS dark/light mode setting
3. **Light** - Default fallback

## Using the Theme Store

### Access Current Theme

```svelte
<script lang="ts">
	import { theme } from '$lib/stores/theme';
</script>

{$theme} <!-- "light" or "dark" -->
```

### Change Theme

```svelte
<script lang="ts">
	import { theme } from '$lib/stores/theme';

	// Set specific theme
	theme.setTheme('dark');
	theme.setTheme('light');

	// Toggle between themes
	const newTheme = $theme === 'light' ? 'dark' : 'light';
	theme.setTheme(newTheme);
</script>
```

## Using Theme Toggle Component

A pre-built component is available at `src/lib/components/ThemeToggle.svelte`:

```svelte
<script lang="ts">
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
</script>

<ThemeToggle />
```

The component displays a sun icon in light mode and a moon icon in dark mode.

## Styling with CSS Variables

Use the theme-aware CSS variables in your styles:

```css
/* Colors automatically adapt to theme */
color: rgb(var(--color-foreground));
background: rgb(var(--color-background));
border-color: rgb(var(--color-border));
```

## Available CSS Variables

| Variable | Purpose |
|----------|---------|
| `--color-foreground` | Text and foreground elements |
| `--color-background` | Page and element backgrounds |
| `--color-border` | Borders and dividers |
| `--color-muted-foreground` | Secondary text |
| `--color-muted-background` | Subtle backgrounds |

## Styling with Tailwind

Use the custom color classes in Tailwind (these use CSS variables):

```html
<div class="bg-background text-foreground border border-border">
	<span class="text-muted-foreground">Muted text</span>
</div>
```

## Color Values

### Light Theme
- **Foreground**: #000000 (black)
- **Background**: #FFFFFF (white)
- **Border**: #DEDEDE (light gray)
- **Muted Foreground**: #666666 (gray)
- **Muted Background**: #F5F5F5 (very light gray)

### Dark Theme
- **Foreground**: #FFFFFF (white)
- **Background**: #141414 (very dark gray)
- **Border**: #333333 (dark gray)
- **Muted Foreground**: #999999 (light gray)
- **Muted Background**: #262626 (dark gray)

## Customizing Theme Colors

Edit the CSS variables in `src/routes/layout.css`:

```css
:root {
	--color-foreground: 0 0% 0%;
	--color-background: 0 0% 100%;
	/* ... more variables */
}

html.dark {
	--color-foreground: 0 0% 100%;
	--color-background: 0 0% 8%;
	/* ... more variables */
}
```

Colors are defined in **HSL format** (Hue Saturation Lightness) for easier customization.

## Environment Configuration

To set a default theme (instead of using system preference), you can:

1. **Browser localStorage**: Automatically set when user toggles theme
2. **Client-side hook**: `src/hooks.client.ts` handles initialization

Currently, the theme respects:
- User's saved preference in localStorage
- System dark/light mode preference
- Default to light theme if neither is available

## Best Practices

1. **Always use CSS variables** for colors instead of hardcoded values
2. **Test both themes** during development using `ThemeToggle`
3. **Use semantic color names** (foreground, background, border) instead of color names (red, blue)
4. **Avoid overriding** `color-scheme` CSS property
5. **Keep contrast ratios** above 4.5:1 for WCAG AA compliance

## Verifying Theme Implementation

### Manual Testing
1. Click the `ThemeToggle` component in your UI
2. Refresh the page - theme should persist
3. Clear localStorage - theme should follow system preference
4. Change OS theme - app should adapt if no saved preference

### Code Testing
```svelte
<script lang="ts">
	import { theme } from '$lib/stores/theme';

	// Subscribe to theme changes
	theme.subscribe((t) => console.log('Theme:', t));
</script>
```

## Accessibility

- Theme preference respects `prefers-color-scheme` media query
- Uses `color-scheme` CSS property for form elements
- Color contrast meets WCAG AA standards in both themes
- Icons use `aria-hidden="true"` where appropriate
