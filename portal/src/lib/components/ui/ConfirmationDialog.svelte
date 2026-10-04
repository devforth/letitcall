<script lang="ts">
	import Button from '$lib/components/ui/Button.svelte';
	import Dialog from '$lib/components/ui/Dialog.svelte';

	let {
		open,
		title,
		description,
		confirmLabel,
		cancelLabel = 'Cancel',
		confirmingLabel = 'Confirming…',
		confirming = false,
		onconfirm,
		oncancel
	}: {
		open: boolean;
		title: string;
		description: string;
		confirmLabel: string;
		cancelLabel?: string;
		confirmingLabel?: string;
		confirming?: boolean;
		onconfirm: () => void;
		oncancel: () => void;
	} = $props();

	function cancel() {
		if (!confirming) oncancel();
	}
</script>

<Dialog {open} label={title} bare bottom oncancel={cancel}>
	<div class="confirm-bar">
		<div class="confirm-content">
			<div class="confirm-text">
				<h2 class="confirm-title">{title}</h2>
				<p class="confirm-desc">{description}</p>
			</div>
			<div class="confirm-actions">
				<Button rounded variant="primary-outline" class="modal-action-button" disabled={confirming} onclick={cancel}>
					{cancelLabel}
				</Button>
				<Button rounded class="modal-action-button" disabled={confirming} onclick={onconfirm}>
					{confirming ? confirmingLabel : confirmLabel}
				</Button>
			</div>
		</div>
	</div>
</Dialog>

<style>
	.confirm-bar {
		border-top: 3px solid rgb(var(--error));
		background: rgb(var(--color-background));
		box-shadow: var(--shadow);
		padding: 1rem 1.5rem calc(1rem + env(safe-area-inset-bottom, 0px));
	}

	.confirm-content {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.75rem 1.5rem;
		max-width: 72rem;
		margin: 0 auto;
	}

	.confirm-text {
		display: grid;
		flex: 1 1 16rem;
		gap: 0.125rem;
		min-width: 0;
	}

	.confirm-title {
		margin: 0;
		color: rgb(var(--color-text));
		font-size: 1.125rem;
		font-weight: 600;
	}

	.confirm-desc {
		margin: 0;
		color: rgb(var(--color-text) / 0.65);
		font-size: 0.875rem;
		line-height: 1.45;
	}

	.confirm-actions {
		display: flex;
		flex: none;
		gap: 0.5rem;
	}
</style>
