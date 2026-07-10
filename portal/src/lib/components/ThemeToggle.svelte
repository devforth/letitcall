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
	class="rounded-md p-2 hover:bg-muted-background transition-colors"
>
	{#if $theme === 'light'}
		<svg
			class="w-5 h-5"
			fill="none"
			stroke="currentColor"
			viewBox="0 0 24 24"
			aria-hidden="true"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
			/>
		</svg>
	{:else}
		<svg
			class="w-5 h-5"
			fill="none"
			stroke="currentColor"
			viewBox="0 0 24 24"
			aria-hidden="true"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M12 3v1m0 16v1m9-9h-1m-16 0H1m15.364 1.636l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
			/>
		</svg>
	{/if}
</button>

<style>
	button {
		border: none;
		background: none;
		cursor: pointer;
		color: rgb(var(--color-foreground));
	}
</style>
