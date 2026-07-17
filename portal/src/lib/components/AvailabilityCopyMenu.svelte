<script lang="ts">
	import Icon from '@iconify/svelte';
	import copyIcon from '@iconify-icons/tabler/copy';
	import Button from '$lib/components/ui/Button.svelte';
	import Checkbox from '$lib/components/ui/Checkbox.svelte';

	let {
		sourceDay,
		days,
		oncopy
	}: {
		sourceDay: string;
		days: { day: string; label: string }[];
		oncopy: (days: string[]) => void;
	} = $props();

	let open = $state(false);
	let selected = $state<Record<string, boolean>>(emptySelection());

	const copyDisabled = $derived(!days.some(({ day }) => selected[day]));

	function copyRanges() {
		oncopy(days.filter(({ day }) => selected[day]).map(({ day }) => day));
		selected = emptySelection();
		open = false;
	}

	function emptySelection() {
		return Object.fromEntries(days.map(({ day }) => [day, false]));
	}
</script>

<details class="relative" bind:open>
	<summary
		class="copy-trigger grid size-10 cursor-pointer list-none place-items-center rounded-[10px] transition focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2"
		aria-label={`Copy ${sourceDay} ranges to other days`}
		title="Copy ranges to other days"
	>
		<Icon icon={copyIcon} width="22" height="22" />
	</summary>
	<div class="absolute right-0 z-10 mt-2 w-64 rounded-lg border-2 p-4 shadow-[var(--shadow-small)]" style="background: rgb(var(--color-foreground)); border-color: rgb(var(--color-border));">
		<p class="text-sm font-semibold" style="color: rgb(var(--color-text));">Copy ranges to</p>
		<div class="mt-2 grid">
			{#each days as target (target.day)}
				<Checkbox
					id={`copy-${sourceDay}-${target.day}`}
					label={target.label}
					bind:checked={selected[target.day]}
				/>
			{/each}
		</div>
		<div class="mt-3">
			<Button fullWidth disabled={copyDisabled} onclick={copyRanges}>Copy ranges</Button>
		</div>
	</div>
</details>

<style>
	.copy-trigger {
		color: rgb(var(--color-text));
		--tw-ring-color: rgb(var(--color-primary));
	}

	.copy-trigger:hover {
		background: rgb(var(--color-primary) / 0.14);
		color: rgb(var(--color-primary));
	}
</style>
