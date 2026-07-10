<script lang="ts">
	import Button from '$lib/components/ui/Button.svelte';

	let {
		open,
		title,
		description,
		confirmLabel,
		confirmingLabel = 'Confirming…',
		confirming = false,
		onconfirm,
		oncancel
	}: {
		open: boolean;
		title: string;
		description: string;
		confirmLabel: string;
		confirmingLabel?: string;
		confirming?: boolean;
		onconfirm: () => void;
		oncancel: () => void;
	} = $props();

	let dialog: HTMLDialogElement;

	$effect(() => {
		if (open && !dialog.open) dialog.showModal();
		if (!open && dialog.open) dialog.close();
	});

	function cancel(event?: Event) {
		event?.preventDefault();
		if (!confirming) oncancel();
	}
</script>

<dialog
	bind:this={dialog}
	class="m-auto w-[min(32rem,calc(100%-2rem))] border border-black bg-white p-0 text-black shadow-[6px_6px_0_0_#000] backdrop:bg-black/50"
	aria-label={title}
	oncancel={cancel}
>
	<div class="p-6">
		<h2 class="text-xl font-semibold tracking-tight">{title}</h2>
		<p class="mt-3 text-sm leading-6">{description}</p>
		<div class="mt-6 flex flex-wrap justify-end gap-2">
			<Button variant="secondary" disabled={confirming} onclick={() => cancel()}>Cancel</Button>
			<Button variant="danger" disabled={confirming} onclick={onconfirm}>
				{confirming ? confirmingLabel : confirmLabel}
			</Button>
		</div>
	</div>
</dialog>
