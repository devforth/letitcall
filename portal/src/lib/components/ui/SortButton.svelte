<script lang="ts">
	import Icon from '@iconify/svelte';
	import arrowDownIcon from '@iconify-icons/tabler/arrow-down';
	import arrowUpIcon from '@iconify-icons/tabler/arrow-up';

	let {
		label,
		active,
		direction,
		onclick
	}: { label: string; active: boolean; direction: 'ascending' | 'descending'; onclick: () => void } = $props();
</script>

<button type="button" class="sort-button" {onclick}>
	<span>{label}</span>
	<span class="sort-arrow" class:inactive={!active}>
		<Icon icon={direction === 'ascending' ? arrowUpIcon : arrowDownIcon} width="16" height="16" aria-hidden="true" />
	</span>
</button>

<style>
	.sort-button {
		display: inline-flex;
		align-items: center;
		gap: 0.0625rem;
		border: 0;
		padding: 0;
		background: transparent;
		color: rgb(var(--color-primary));
		font: inherit;
		letter-spacing: inherit;
		text-transform: inherit;
		cursor: pointer;
	}

	.sort-button:focus-visible {
		outline: 2px solid rgb(var(--color-primary));
		outline-offset: 3px;
	}

	.sort-arrow {
		display: inline-flex;
	}

	.sort-arrow.inactive {
		visibility: hidden;
	}
</style>
