<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		label,
		children,
		disabled = false,
		tone = 'neutral',
		onclick
	}: {
		label: string;
		children: Snippet;
		disabled?: boolean;
		tone?: 'neutral' | 'primary' | 'danger';
		onclick?: (event: MouseEvent) => void;
	} = $props();
</script>

<button
	type="button"
	aria-label={label}
	title={label}
	{disabled}
	{onclick}
	class="icon-button tone-{tone} grid size-10 shrink-0 cursor-pointer place-items-center rounded-[10px] transition focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent"
>
	{@render children()}
</button>

<style>
	.icon-button {
		color: rgb(var(--color-text));
		--ring-color: rgb(var(--color-primary));
	}

	.icon-button:focus-visible {
		--tw-ring-color: var(--ring-color);
	}

	.tone-neutral:hover:not(:disabled) {
		background: rgb(var(--color-text) / 0.1);
		color: rgb(var(--color-text));
	}

	.tone-primary:hover:not(:disabled) {
		background: rgb(var(--color-primary) / 0.14);
		color: rgb(var(--color-primary));
	}

	.tone-danger {
		--ring-color: rgb(var(--error));
	}

	.tone-danger:hover:not(:disabled) {
		background: rgb(var(--error) / 0.14);
		color: rgb(var(--error));
	}
</style>
