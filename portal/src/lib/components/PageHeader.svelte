<script lang="ts">
	import Icon from '@iconify/svelte';
	import BackLink from '$lib/components/BackLink.svelte';
	import type { IconifyIcon } from '@iconify/svelte';
	import type { Snippet } from 'svelte';

	let {
		id,
		title,
		description,
		icon,
		parent,
		count,
		children
	}: {
		id: string;
		title: string;
		description: string;
		icon: IconifyIcon;
		parent?: { href: string; label: string };
		count?: number;
		children?: Snippet;
	} = $props();
</script>

{#if parent}
	<div class="page-parent"><BackLink href={parent.href} label={parent.label} /></div>
{/if}
<header class="page-header">
	<div class="page-heading">
		<div class="page-heading-icon">
			<Icon {icon} width="46" height="46" />
		</div>
		<div class="min-w-0">
			<h1 {id}>{title}{#if count}<span class="page-count">·&nbsp;{count}</span>{/if}</h1>
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

	@media (max-width: 39.999rem) {
		.page-heading {
			align-items: flex-start;
		}
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

	.page-parent {
		margin-bottom: 0.75rem;
	}

	.page-count {
		margin-left: 0.375rem;
		color: rgb(var(--color-text) / 0.65);
		font-size: 1.125rem;
		font-weight: 400;
		font-variant-numeric: tabular-nums;
	}

	.page-heading-actions {
		flex: none;
	}

</style>
