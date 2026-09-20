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

<Dialog {open} label={title} bare oncancel={cancel}>
	<div class="confirm-card">
		<div class="glyph" aria-hidden="true">
			<svg
				width="52"
				height="52"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2.2"
				stroke-linecap="round"
				stroke-linejoin="round"
			>
				<path d="M12 9v4" />
				<path d="M12 17h.01" />
				<path d="M10.3 3.9 1.8 18a2 2 0 0 0 1.7 3h17a2 2 0 0 0 1.7-3L13.7 3.9a2 2 0 0 0-3.4 0z" />
			</svg>
		</div>
		<h2 class="confirm-title">{title}</h2>
		<p class="confirm-desc">{description}</p>
		<div class="confirm-actions">
			<Button rounded variant="primary-outline" class="modal-action-button" disabled={confirming} onclick={cancel}>
				{cancelLabel}
			</Button>
			<Button rounded class="modal-action-button" disabled={confirming} onclick={onconfirm}>
				{confirming ? confirmingLabel : confirmLabel}
			</Button>
		</div>
	</div>
</Dialog>

<style>
	.confirm-card {
		background: rgb(var(--color-background));
		border: 1px solid color-mix(in srgb, var(--color-border) 60%, transparent);
		border-radius: 16px;
		box-shadow: var(--shadow);
		padding: 1.75rem 1.5rem;
		text-align: center;
	}

	.glyph {
		display: flex;
		justify-content: center;
		margin-bottom: 1rem;
		color: rgb(var(--error));
	}

	.confirm-title {
		margin: 0;
		font-size: 1.25rem;
		font-weight: 700;
		letter-spacing: -0.01em;
		color: #1a1a1a;
	}

	:global(html.dark) .confirm-title {
		color: #f5f5f5;
	}

	.confirm-desc {
		margin: 0.5rem 0 0;
		font-size: 0.9rem;
		line-height: 1.55;
		color: rgb(var(--color-text));
	}

	.confirm-actions {
		margin-top: 1.5rem;
		display: flex;
		gap: 1rem;
	}

	:global(.confirm-actions .modal-action-button) {
		min-width: 0;
		flex: 1;
		padding-inline: 0.75rem !important;
	}
</style>
