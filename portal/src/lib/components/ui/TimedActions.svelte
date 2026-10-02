<script lang="ts">
	import { browser } from '$app/environment';
	import { onMount, type Snippet } from 'svelte';
	import Icon from '@iconify/svelte';
	import dotsVerticalIcon from '@iconify-icons/tabler/dots-vertical';

	let {
		label,
		controlsId,
		size = '2.5rem',
		actionsVisibleOnSmallScreens = false,
		class: className = '',
		children
	}: {
		label: string;
		controlsId: string;
		size?: string;
		actionsVisibleOnSmallScreens?: boolean;
		class?: string;
		children: Snippet;
	} = $props();

	const actionTimeoutMs = 3000;
	let root: HTMLDivElement;
	let desktopActions = $state(browser && window.matchMedia('(min-width: 40rem)').matches);
	let hovered = $state(false);
	let open = $state(false);
	let actionTimeout: ReturnType<typeof setTimeout> | undefined;
	let actionsVisible = $derived(open || hovered || (actionsVisibleOnSmallScreens && !desktopActions));

	function scheduleClose(): void {
		clearTimeout(actionTimeout);
		actionTimeout = setTimeout(() => {
			if (root.contains(document.activeElement)) {
				scheduleClose();
				return;
			}
			open = false;
		}, actionTimeoutMs);
	}

	function showActions(event: MouseEvent): void {
		event.preventDefault();
		event.stopPropagation();
		open = true;
		scheduleClose();

		const trigger = event.currentTarget as HTMLButtonElement;
		if (event.detail === 0) {
			requestAnimationFrame(() =>
				root.querySelector<HTMLElement>('a[href], button:not(:disabled)')?.focus()
			);
		} else {
			trigger.blur();
		}
	}

	onMount(() => {
		const row = root.closest<HTMLElement>('[data-timed-actions-row]');
		const media = window.matchMedia('(min-width: 40rem)');
		const updateDesktopActions = () => (desktopActions = media.matches);
		const handlePointerEnter = (event: PointerEvent) => {
			if (event.pointerType === 'mouse') hovered = true;
		};
		const handlePointerLeave = (event: PointerEvent) => {
			if (event.pointerType === 'mouse') hovered = false;
		};

		updateDesktopActions();
		media.addEventListener('change', updateDesktopActions);
		row?.addEventListener('pointerenter', handlePointerEnter);
		row?.addEventListener('pointerleave', handlePointerLeave);

		return () => {
			clearTimeout(actionTimeout);
			media.removeEventListener('change', updateDesktopActions);
			row?.removeEventListener('pointerenter', handlePointerEnter);
			row?.removeEventListener('pointerleave', handlePointerLeave);
		};
	});
</script>

<div
	bind:this={root}
	class={`timed-actions ${className}`}
	class:actions-visible={actionsVisible}
	style={`--timed-actions-size: ${size}`}
>
	<button
		type="button"
		class="timed-actions-trigger"
		aria-label={label}
		aria-expanded={actionsVisible}
		aria-controls={controlsId}
		onclick={showActions}
	>
		<Icon icon={dotsVerticalIcon} width="22" height="22" />
	</button>
	<div id={controlsId} class="timed-actions-content" aria-hidden={!actionsVisible} inert={!actionsVisible}>
		{@render children()}
	</div>
</div>

<style>
	.timed-actions {
		position: relative;
		display: flex;
		min-height: var(--timed-actions-size);
		align-items: center;
		justify-content: flex-end;
	}

	.timed-actions-trigger {
		position: absolute;
		right: 0;
		display: grid;
		width: var(--timed-actions-size);
		height: var(--timed-actions-size);
		place-items: center;
		border: 0;
		border-radius: 10px;
		padding: 0;
		background: transparent;
		color: rgb(var(--color-text) / 0.6);
		cursor: pointer;
		transition:
			opacity 0.18s ease,
			transform 0.18s ease;
	}

	.timed-actions-content {
		opacity: 0;
		pointer-events: none;
		transform: translateX(0.5rem);
		transition:
			opacity 0.18s ease,
			transform 0.18s ease;
	}

	.actions-visible .timed-actions-content {
		opacity: 1;
		pointer-events: auto;
		transform: translateX(0);
	}

	.actions-visible .timed-actions-trigger {
		opacity: 0;
		pointer-events: none;
		transform: translateX(-0.5rem) scale(0.85);
	}

	.timed-actions-trigger:focus-visible {
		outline: 2px solid rgb(var(--color-text) / 0.65);
		outline-offset: 2px;
	}

	@media (prefers-reduced-motion: reduce) {
		.timed-actions-trigger,
		.timed-actions-content {
			transition: none;
		}
	}
</style>
