<script lang="ts">
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
			<button
				type="button"
				class="btn-delete"
				disabled={confirming}
				onclick={onconfirm}
			>
				{confirming ? confirmingLabel : confirmLabel}
			</button>
			<button type="button" class="btn-cancel" disabled={confirming} onclick={cancel}>Cancel</button>
		</div>
	</div>
</Dialog>

<style>
	.confirm-card {
		background: rgb(var(--color-foreground));
		border: 1px solid rgb(var(--color-border) / 0.6);
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
		flex-direction: column;
		gap: 0.6rem;
	}

	.confirm-actions button {
		width: 100%;
		padding: 0.7rem 1rem;
		border-radius: 10px;
		font: inherit;
		font-weight: 700;
		font-size: 0.95rem;
		cursor: pointer;
		transition: all 0.2s;
	}

	.confirm-actions button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.btn-delete {
		background: rgb(var(--error));
		color: #fff;
		border: 2px solid rgb(var(--error));
	}

	.btn-delete:hover:not(:disabled) {
		background: color-mix(in srgb, rgb(var(--error)), black 8%);
		border-color: color-mix(in srgb, rgb(var(--error)), black 8%);
	}

	.btn-cancel {
		background: transparent;
		color: rgb(var(--color-text));
		border: 2px solid rgb(var(--color-border));
	}

	.btn-cancel:hover:not(:disabled) {
		background: rgb(var(--color-text) / 0.06);
	}
</style>
