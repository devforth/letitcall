<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		open,
		label,
		wide = false,
		bare = false,
		bottom = false,
		sheet = false,
		children,
		oncancel
	}: {
		open: boolean;
		label: string;
		wide?: boolean;
		bare?: boolean;
		bottom?: boolean;
		/** On small screens, a bottom panel that can be swiped down to close. */
		sheet?: boolean;
		children: Snippet;
		oncancel: () => void;
	} = $props();

	let dialog: HTMLDialogElement;

	$effect(() => {
		if (open && !dialog.open) dialog.showModal();
		if (!open && dialog.open) dialog.close();
	});

	function cancel(event: Event) {
		event.preventDefault();
		oncancel();
	}

	const swipeCloseDistance = 80;
	let swipeStart = 0;
	let swipeOffset = $state(0);

	function startSwipe(event: PointerEvent) {
		swipeStart = event.clientY;
		(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId);
	}

	function moveSwipe(event: PointerEvent) {
		if (swipeStart) swipeOffset = Math.max(0, event.clientY - swipeStart);
	}

	function endSwipe() {
		if (swipeOffset > swipeCloseDistance) oncancel();
		swipeStart = 0;
		swipeOffset = 0;
	}
</script>

<dialog
	bind:this={dialog}
	class:wide
	class:bare
	class:bottom
	class:sheet
	class:swiping={swipeOffset > 0}
	class="dialog"
	style={swipeOffset ? `transform: translateY(${swipeOffset}px)` : undefined}
	aria-label={label}
	oncancel={cancel}
>
	{#if sheet}
		<div
			class="sheet-grip"
			aria-hidden="true"
			onpointerdown={startSwipe}
			onpointermove={moveSwipe}
			onpointerup={endSwipe}
			onpointercancel={endSwipe}
		></div>
	{/if}
	<div class={bare ? '' : 'p-6'}>{@render children()}</div>
</dialog>

<style>
	.dialog {
		margin: auto;
		width: min(32rem, calc(100% - 2rem));
		border: 2px solid #000;
		padding: 0;
		background: #fff;
		color: #000;
		box-shadow: 8px 8px 0 #000;
	}

	.dialog.wide {
		width: min(38rem, calc(100% - 2rem));
	}

	.dialog.bare {
		width: min(23rem, calc(100% - 2rem));
		border: none;
		background: transparent;
		box-shadow: none;
	}

	.dialog.wide.bare {
		width: min(38rem, calc(100% - 2rem));
	}

	.dialog.bottom {
		margin: auto 0 0;
		width: 100%;
		max-width: 100%;
	}

	.sheet-grip {
		display: none;
	}

	@media (max-width: 39.999rem) {
		/* Also outranks the wide bare width set above. */
		.dialog.sheet,
		.dialog.wide.bare.sheet {
			position: fixed;
			margin: auto 0 0;
			width: 100%;
			max-width: 100%;
			overflow: hidden;
			border-radius: 1rem 1rem 0 0;
			animation: sheet-in 0.25s ease-out;
			transition: transform 0.2s ease;
		}

		.dialog.sheet.swiping {
			transition: none;
		}

		.sheet-grip {
			position: absolute;
			z-index: 1;
			top: 0;
			left: 0;
			right: 0;
			display: grid;
			height: 1.5rem;
			place-items: center;
			touch-action: none;
			cursor: grab;
		}

		.sheet-grip::before {
			content: '';
			width: 2.5rem;
			height: 4px;
			border-radius: 999px;
			background: var(--color-border);
		}
	}

	@keyframes sheet-in {
		from {
			transform: translateY(100%);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.dialog.sheet {
			animation: none;
			transition: none;
		}
	}

	.dialog::backdrop {
		background: rgb(0 0 0 / 0.65);
	}

</style>
