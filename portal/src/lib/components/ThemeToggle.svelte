<script lang="ts">
	import { theme } from '$lib/stores/theme';

	function handleToggle() {
		const newTheme = $theme === 'light' ? 'dark' : 'light';
		theme.setTheme(newTheme);
	}
</script>

<button
	onclick={handleToggle}
	aria-label="Toggle theme"
	class="toggle-switch"
	class:dark={$theme === 'dark'}
>
	<svg class="sun-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
		<circle cx="12" cy="12" r="5" />
		<line x1="12" y1="1" x2="12" y2="3" />
		<line x1="12" y1="21" x2="12" y2="23" />
		<line x1="4.22" y1="4.22" x2="5.64" y2="5.64" />
		<line x1="18.36" y1="18.36" x2="19.78" y2="19.78" />
		<line x1="1" y1="12" x2="3" y2="12" />
		<line x1="21" y1="12" x2="23" y2="12" />
		<line x1="4.22" y1="19.78" x2="5.64" y2="18.36" />
		<line x1="18.36" y1="5.64" x2="19.78" y2="4.22" />
	</svg>
	<svg class="moon-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
		<path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
	</svg>
</button>

<style>
	.toggle-switch {
		position: relative !important;
		width: 68px !important;
		height: 36px !important;
		border: 2px solid rgb(var(--color-border)) !important;
		border-radius: 20px !important;
		background: rgb(var(--color-foreground)) !important;
		color: rgb(var(--color-text)) !important;
		cursor: pointer !important;
		padding: 0 !important;
		display: flex !important;
		align-items: center !important;
		justify-content: flex-start !important;
		transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) !important;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1) !important;
	}

	.toggle-switch:hover {
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15) !important;
	}

	.toggle-switch.dark {
		justify-content: flex-end !important;
	}

	.toggle-switch.dark:hover {
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3) !important;
	}

	.sun-icon {
		position: absolute !important;
		left: 6px !important;
		width: 20px !important;
		height: 20px !important;
		transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) !important;
		flex-shrink: 0 !important;
	}

	.toggle-switch.dark .sun-icon {
		opacity: 0 !important;
		transform: scale(0.5) translateX(-10px) !important;
	}

	.moon-icon {
		position: absolute !important;
		right: 6px !important;
		width: 20px !important;
		height: 20px !important;
		opacity: 0 !important;
		transform: scale(0.5) translateX(10px) !important;
		transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) !important;
		flex-shrink: 0 !important;
	}

	.toggle-switch.dark .moon-icon {
		opacity: 1 !important;
		transform: scale(1) translateX(0) !important;
	}
</style>
