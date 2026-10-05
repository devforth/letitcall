<script lang="ts">
	import Icon from '@iconify/svelte';
	import arrowLeftIcon from '@iconify-icons/at-icons/arrow-left';

	let {
		href,
		label,
		reload = false,
		showIcon = true
	}: { href: string; label: string; reload?: boolean; showIcon?: boolean } = $props();
</script>

<a class="back-link" {href} data-sveltekit-reload={reload ? '' : undefined}>
	{#if showIcon}<Icon icon={arrowLeftIcon} width="12" height="12" style="margin-top: 0.15rem;" />{/if}
	{label}
</a>

<style>
	.back-link {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		color: rgb(var(--color-primary));
		font-size: 1.125rem;
		transition: opacity 0.2s ease, transform 0.1s ease;
	}

	.back-link:hover {
		opacity: 0.8;
	}

	.back-link:active {
		transform: scale(0.96);
	}

	@media (prefers-reduced-motion: reduce) {
		.back-link {
			transition: none;
		}
	}
</style>
