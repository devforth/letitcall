# Theme Integration Examples

This guide shows how to use the dark/light theme in your components.

## Adding Theme Toggle to AppShell

Update your AppShell header to include the theme toggle button:

```svelte
<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { Snippet } from 'svelte';
	import { api, appPath } from '$lib/api';
	import type { SessionUser } from '$lib/types';
	import Button from '$lib/components/ui/Button.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';

	let {
		user,
		children
	}: {
		user: SessionUser;
		children: Snippet;
	} = $props();

	let loggingOut = $state(false);

	async function logout() {
		loggingOut = true;
		try {
			await api('/api/auth/logout', { method: 'POST' });
		} finally {
			await goto(appPath('/auth/login'), { replaceState: true });
		}
	}
</script>

<div class="min-h-screen bg-background text-foreground">
	<header class="border-b border-border">
		<div class="mx-auto flex max-w-7xl flex-wrap items-center justify-between gap-4 px-4 py-4 sm:px-6 lg:px-8">
			<a class="text-lg font-bold tracking-tight" href={appPath('/')}>Let It Call</a>
			<div class="flex items-center gap-4">
				<span class="hidden text-sm sm:inline">{user.email}</span>
				<ThemeToggle />
				<Button variant="secondary" disabled={loggingOut} onclick={logout}>
					{loggingOut ? 'Signing out…' : 'Sign out'}
				</Button>
			</div>
		</div>
	</header>

	<div class="mx-auto grid max-w-7xl md:grid-cols-[14rem_1fr]">
		<nav class="border-b border-border p-4 md:min-h-[calc(100vh-77px)] md:border-r md:border-b-0 sm:p-6">
			<ul class="flex gap-2 md:grid">
				<li>
					<a
						class={`block border border-border px-4 py-3 text-sm font-medium ${
							page.url.pathname === appPath('/') 
								? 'bg-foreground text-background' 
								: 'bg-background text-foreground'
						}`}
						href={appPath('/')}>Dashboard</a
					>
				</li>
				<li>
					<a
						class={`block border border-border px-4 py-3 text-sm font-medium ${
							page.url.pathname.startsWith(appPath('/users'))
								? 'bg-foreground text-background'
								: 'bg-background text-foreground'
						}`}
						href={appPath('/users')}>Users</a
					>
				</li>
			</ul>
		</nav>

		<main class="min-w-0 p-4 sm:p-6 lg:p-10">
			{@render children()}
		</main>
	</div>
</div>
```

## Conditional Styling Based on Theme

Sometimes you need different styles for light and dark modes:

```svelte
<script lang="ts">
	import { theme } from '$lib/stores/theme';
</script>

<div class={$theme === 'dark' ? 'shadow-lg shadow-black' : 'shadow-lg shadow-gray-300'}>
	Content
</div>
```

Or using CSS classes with Tailwind dark mode:

```html
<!-- Applies different shadows in dark mode -->
<div class="shadow-lg shadow-gray-300 dark:shadow-black">
	Content
</div>
```

## Using CSS Variables in Component Styles

```svelte
<script lang="ts">
	export let accentColor = 'blue';
</script>

<div style="border-color: rgb(var(--color-border)); background: rgb(var(--color-background));">
	Content
</div>

<style>
	div {
		padding: 1rem;
		border: 1px solid rgb(var(--color-border));
		background: rgb(var(--color-background));
		color: rgb(var(--color-foreground));
	}
</style>
```

## Creating Semantic Color Utilities

Create helper functions for theme-aware colors:

```typescript
// src/lib/utils/colors.ts
import { get } from 'svelte/store';
import { theme } from '$lib/stores/theme';

export function getThemeColor(semanticColor: 'foreground' | 'background' | 'border') {
	const currentTheme = get(theme);
	const colorMap = {
		light: {
			foreground: '#000000',
			background: '#FFFFFF',
			border: '#DEDEDE'
		},
		dark: {
			foreground: '#FFFFFF',
			background: '#141414',
			border: '#333333'
		}
	};
	return colorMap[currentTheme][semanticColor];
}
```

## Accessible Color Contrast

All theme colors are designed to meet WCAG AA standards (4.5:1 contrast ratio):

**Light Theme:**
- Black text (#000) on white background (#FFF) = 21:1 ✓
- Gray text (#666) on white background (#FFF) = 7.5:1 ✓

**Dark Theme:**
- White text (#FFF) on dark background (#141414) = 17.5:1 ✓
- Light gray text (#999) on dark background (#141414) = 8:1 ✓

## Customizing Colors for Your Brand

Edit `src/routes/layout.css` to match your brand colors:

```css
:root {
	/* Your light theme colors */
	--color-foreground: 10 20% 5%;      /* Dark blue-ish black */
	--color-background: 240 100% 98%;   /* Light blue tint */
	--color-border: 220 10% 85%;        /* Blue-gray border */
	--color-muted-foreground: 220 10% 40%;
	--color-muted-background: 240 100% 95%;
}

html.dark {
	/* Your dark theme colors */
	--color-foreground: 240 100% 98%;
	--color-background: 220 15% 12%;
	--color-border: 220 15% 25%;
	--color-muted-foreground: 220 10% 60%;
	--color-muted-background: 220 15% 22%;
}
```

HSL values make it easy to maintain color relationships between light and dark modes.

## Testing Theme Changes

```svelte
<script lang="ts">
	import { theme } from '$lib/stores/theme';
</script>

<button onclick={() => theme.setTheme('light')}>Light</button>
<button onclick={() => theme.setTheme('dark')}>Dark</button>

<p>Current theme: {$theme}</p>
```
