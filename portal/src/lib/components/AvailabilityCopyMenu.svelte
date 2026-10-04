<script lang="ts">
	import Icon from '@iconify/svelte';
	import copyIcon from '@iconify-icons/tabler/copy';
	import arrowRightIcon from '@iconify-icons/tabler/arrow-right';
	import IconButton from '$lib/components/ui/IconButton.svelte';
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
	let menu: HTMLDetailsElement;
	let panel: HTMLDivElement;
	let above = $state(false);

	$effect(() => {
		if (open) above = window.innerHeight - menu.getBoundingClientRect().bottom < panel.offsetHeight + 8;
	});
	let selected = $state<Record<string, boolean>>(emptySelection());

	const hasSelection = $derived(days.some(({ day }) => selected[day]));

	function copyRanges() {
		oncopy(days.filter(({ day }) => selected[day]).map(({ day }) => day));
		selected = emptySelection();
		open = false;
	}

	function emptySelection() {
		return Object.fromEntries(days.map(({ day }) => [day, false]));
	}
</script>

<svelte:window onpointerdown={(event) => open && !menu.contains(event.target as Node) && (open = false)} />

<details bind:this={menu} class="relative" bind:open>
	<summary
		class="copy-trigger grid size-10 cursor-pointer list-none place-items-center rounded-[10px] transition focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2"
		aria-label={`Copy ${sourceDay} ranges to other days`}
		title="Copy ranges to other days"
	>
		<Icon icon={copyIcon} width="22" height="22" />
	</summary>
	<div bind:this={panel} class={`absolute right-0 z-10 ${above ? 'bottom-full mb-2' : 'mt-2'} w-64 rounded-lg border-2 p-4 shadow-[var(--shadow-small)]`} style="background: rgb(var(--color-background)); border-color: var(--color-border);">
		<p class="text-sm leading-4" style="color: rgb(var(--color-text) / 0.65);">Copy ranges to:</p>
		<div class="mt-2 flex items-end justify-between gap-3">
			<div class="copy-days grid">
				{#each days as target (target.day)}
					<Checkbox
						id={`copy-${sourceDay}-${target.day}`}
						label={target.label}
						bind:checked={selected[target.day]}
					/>
				{/each}
			</div>
			{#if hasSelection}
				<IconButton filled tone="primary" label="Copy ranges" onclick={copyRanges}>
					<Icon icon={arrowRightIcon} width="22" height="22" />
				</IconButton>
			{/if}
		</div>
	</div>
</details>

<style>
	.copy-days :global(.checkbox) {
		min-height: 2rem;
	}

	.copy-trigger {
		background: rgb(var(--color-primary) / 0.12);
		color: rgb(var(--color-primary));
		--tw-ring-color: rgb(var(--color-primary));
	}

	.copy-trigger:hover,
	details[open] .copy-trigger {
		background: rgb(var(--color-primary));
		color: rgb(var(--color-background));
	}

</style>
