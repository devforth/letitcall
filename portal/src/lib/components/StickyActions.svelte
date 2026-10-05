<script lang="ts">
	import { onMount, type Snippet } from 'svelte';

	let { class: className = '', children }: { class?: string; children: Snippet } = $props();

	let bar: HTMLDivElement;
	let stuck = $state(false);
	let atPageBottom = $state(false);

	onMount(() => {
		// The bar sticks 1px below the viewport, so it is never fully visible while stuck.
		const observer = new IntersectionObserver(([entry]) => (stuck = entry.intersectionRatio < 1), { threshold: [1] });
		const updatePageBottom = () =>
			(atPageBottom = document.documentElement.scrollHeight - window.scrollY - window.innerHeight <= 1);
		observer.observe(bar);
		updatePageBottom();
		window.addEventListener('scroll', updatePageBottom, { passive: true });
		return () => {
			observer.disconnect();
			window.removeEventListener('scroll', updatePageBottom);
		};
	});
</script>

<div bind:this={bar} class={`sticky-actions ${className}`} class:stuck={stuck && !atPageBottom}>
	{@render children()}
</div>

<style>
	.sticky-actions {
		--action-area-width: 100vw;
		position: sticky;
		z-index: 10;
		bottom: -1px;
		display: flex;
		width: var(--action-area-width);
		flex-wrap: wrap;
		align-items: center;
		justify-content: flex-end;
		gap: 0.75rem;
		margin-left: calc((100% - var(--action-area-width)) / 2);
		padding: 1.25rem max(var(--content-padding, 0px), calc((var(--action-area-width) - 72rem) / 2)) calc(1.25rem + 1px);
		background: rgb(var(--color-background));
		transition: box-shadow 0.2s ease;
	}

	@media (min-width: 48rem) {
		.sticky-actions {
			--action-area-width: calc(100vw - var(--sidebar-w));
		}
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
