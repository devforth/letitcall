<script lang="ts">
	import Icon from '@iconify/svelte';
	import type { IconifyIcon } from '@iconify/svelte';
	import type { Snippet } from 'svelte';

	let {
		id,
		title,
		description,
		icon,
		children
	}: {
		id: string;
		title: string;
		description: string;
		icon: IconifyIcon;
		children?: Snippet;
	} = $props();
</script>

<header class="page-header">
	<div class="page-heading">
		<div class="page-heading-icon">
			<Icon {icon} width="46" height="46" />
		</div>
		<div class="min-w-0">
			<h1 {id}>{title}</h1>
			<p>{description}</p>
		</div>
	</div>
	{#if children}
		<div class="page-heading-actions">{@render children()}</div>
	{/if}
</header>

<style>
	.page-header,
	.page-heading {
		display: flex;
		align-items: center;
	}

	.page-header {
		flex-wrap: wrap;
		justify-content: space-between;
		gap: 1.25rem;
	}

	.page-heading {
		min-width: 0;
		gap: 1rem;
	}

	.page-heading-icon {
		flex: none;
		color: rgb(var(--color-text));
	}

	.page-heading-icon :global(svg) {
		margin: -4px -8px -8px;
		transform: scale(0.95);
		transform-origin: center;
	}

	.page-heading-icon :global([stroke]) {
		stroke-width: 1.5;
	}

	h1 {
		margin: 0;
		color: rgb(var(--color-text));
		font-size: 1.5rem;
		font-weight: 600;
		letter-spacing: -0.025em;
		line-height: 1.25;
	}

	p {
		margin: 0;
		color: rgb(var(--color-text) / 0.65);
		font-size: 0.875rem;
		line-height: 1.25;
	}

	.page-heading-actions {
		flex: none;
	}

</style>
