<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		open,
		label,
		wide = false,
		bare = false,
		children,
		oncancel
	}: {
		open: boolean;
		label: string;
		wide?: boolean;
		bare?: boolean;
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
</script>

<dialog bind:this={dialog} class:wide class:bare class="dialog" aria-label={label} oncancel={cancel}>
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

	.dialog::backdrop {
		background: rgb(0 0 0 / 0.65);
	}

</style>
