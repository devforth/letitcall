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
		class="grid size-11 cursor-pointer list-none place-items-center border border-black bg-white text-black transition hover:bg-black hover:text-white focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2"
		aria-label={`Copy ${sourceDay} ranges to other days`}
		title="Copy ranges to other days"
	>
		<Icon icon={copyIcon} width="22" height="22" />
	</summary>
	<div class="absolute right-0 z-10 mt-2 w-64 border border-black bg-white p-4 shadow-[4px_4px_0_0_#000]">
		<p class="text-sm font-semibold">Copy ranges to</p>
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
