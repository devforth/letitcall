<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		label,
		children,
		disabled = false,
		filled = false,
		tone = 'neutral',
		class: className = '',
		onclick
	}: {
		label: string;
		children: Snippet;
		disabled?: boolean;
		filled?: boolean;
		tone?: 'neutral' | 'primary' | 'danger';
		class?: string;
		onclick?: (event: MouseEvent) => void;
	} = $props();
</script>

<button
	type="button"
	aria-label={label}
	title={label}
	{disabled}
	{onclick}
	class="icon-button tone-{tone} grid size-10 shrink-0 cursor-pointer place-items-center rounded-[10px] transition focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent {filled ? 'filled' : ''} {className}"
>
	{@render children()}
</button>

<style>
	.icon-button {
		background: rgb(var(--color-primary) / 0.12);
		color: rgb(var(--color-primary));
		--ring-color: rgb(var(--color-primary));
	}

	.icon-button:focus-visible {
		--tw-ring-color: var(--ring-color);
	}

	.icon-button:hover:not(:disabled) {
		background: rgb(var(--color-primary));
		color: rgb(var(--color-background));
	}

	.icon-button:active:not(:disabled) {
		background: color-mix(in srgb, rgb(var(--color-primary)), black 8%);
		color: rgb(var(--color-background));
	}
</style>
