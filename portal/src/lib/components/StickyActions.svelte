<script lang="ts">
	import { onMount, type Snippet } from 'svelte';

	let { class: className = '', children }: { class?: string; children: Snippet } = $props();

	let bar: HTMLDivElement;
	let stuck = $state(false);

	onMount(() => {
		// The bar sticks 1px below the viewport, so it is never fully visible while stuck.
		const observer = new IntersectionObserver(([entry]) => (stuck = entry.intersectionRatio < 1), { threshold: [1] });
		observer.observe(bar);
		return () => observer.disconnect();
	});
</script>

<div bind:this={bar} class={`sticky-actions ${className}`} class:stuck>
	{@render children()}
</div>

<style>
	.sticky-actions {
		position: sticky;
		z-index: 10;
		bottom: -1px;
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: flex-end;
		gap: 0.75rem;
		/* Spans the page card's padding so the pinned bar reaches its edges. */
		margin-inline: calc(-1 * var(--content-padding, 0px));
		padding: 1.25rem var(--content-padding, 0px) calc(1.25rem + 1px);
		background: rgb(var(--color-background));
		transition: box-shadow 0.2s ease;
	}

	.stuck {
		box-shadow:
			0 -1px 0 var(--color-border),
			0 -8px 24px rgb(0 0 0 / 0.06);
	}

	@media (prefers-reduced-motion: reduce) {
		.sticky-actions {
			transition: none;
		}
	}
</style>
