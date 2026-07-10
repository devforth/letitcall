<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		label,
		children,
		disabled = false,
		onclick
	}: {
		label: string;
		children: Snippet;
		disabled?: boolean;
		onclick?: (event: MouseEvent) => void;
	} = $props();
</script>

<button
	type="button"
	aria-label={label}
	title={label}
	{disabled}
	{onclick}
	class="grid size-11 shrink-0 place-items-center border border-black bg-white text-black transition hover:bg-black hover:text-white focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-40"
>
	{@render children()}
</button>
