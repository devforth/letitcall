<script lang="ts">
	import Button from '$lib/components/ui/Button.svelte';
	import Dialog from '$lib/components/ui/Dialog.svelte';

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

	function cancel() {
		if (!confirming) oncancel();
	}
</script>

<Dialog {open} label={title} oncancel={cancel}>
	<h2 class="text-xl font-semibold tracking-tight">{title}</h2>
	<p class="mt-3 text-sm leading-6">{description}</p>
	<div class="mt-6 flex flex-wrap justify-end gap-2">
		<Button variant="secondary" class="dialog-cancel" disabled={confirming} onclick={cancel}>Cancel</Button>
		<Button variant="danger" class="dialog-confirm" disabled={confirming} onclick={onconfirm}>
			{confirming ? confirmingLabel : confirmLabel}
		</Button>
	</div>
</Dialog>
