<script lang="ts">
	let {
		id,
		label,
		value = $bindable(''),
		placeholder = '',
		required = false,
		disabled = false,
		maxlength,
		rows = 5,
		resizable = true
	}: {
		id: string;
		label: string;
		value?: string;
		placeholder?: string;
		required?: boolean;
		disabled?: boolean;
		maxlength?: number;
		rows?: number;
		resizable?: boolean;
	} = $props();
</script>

<div class="field">
	<div class="input-group" class:filled={!!value}>
		<textarea
			{id}
			bind:value
			{placeholder}
			{required}
			{disabled}
			{maxlength}
			{rows}
			class="input"
			class:not-resizable={!resizable}
		></textarea>
		<label class="float-label" for={id}>{label}</label>
		<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><line x1="4" y1="8" x2="20" y2="8" /><line x1="4" y1="12" x2="14" y2="12" /><line x1="4" y1="16" x2="18" y2="16" /></svg>
	</div>
</div>

<style>
	/* Mirrors Input.svelte: leading icon, floating label that doubles as the
	   placeholder, same border, radius and focus ring. */
	.field {
		display: grid;
		gap: 6px;
		font-size: 0.875rem;
	}

	.input-group {
		position: relative;
		margin: 3px;
	}

	/* Sits on the first text line rather than the middle of the box. */
	.input-group > svg {
		position: absolute;
		left: 12px;
		top: 22px;
		transform: translateY(-50%);
		width: 18px;
		height: 18px;
		color: rgb(var(--color-text));
		pointer-events: none;
		transition: color 0.18s;
	}

	.input {
		display: block;
		width: 100%;
		font: inherit;
		font-size: 0.9rem;
		color: rgb(var(--color-text));
		background: rgb(var(--color-foreground));
		border: 0;
		border-radius: 10px;
		padding: 10px 12px 10px 40px;
		resize: vertical;
		outline: none;
		box-shadow: 0 0 0 1px rgb(var(--color-border));
		transition: box-shadow 0.18s;
	}

	/* label doubles as placeholder — hide the native placeholder until focused */
	.input::placeholder {
		color: transparent;
	}

	.input:focus::placeholder {
		color: rgb(var(--color-text) / 0.4);
	}

	.float-label {
		position: absolute;
		left: 40px;
		top: 22px;
		transform: translateY(-50%);
		max-width: calc(100% - 52px);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		font-size: 0.9rem;
		font-weight: 400;
		color: rgb(var(--color-text));
		opacity: 0.4;
		background: rgb(var(--color-foreground));
		padding: 0 4px;
		pointer-events: none;
		transition: top 0.16s, left 0.16s, font-size 0.16s, color 0.16s, opacity 0.16s;
	}

	.input:focus ~ .float-label,
	.filled .float-label {
		top: 0;
		left: 10px;
		font-size: 0.72rem;
		color: rgb(var(--color-primary));
		opacity: 1;
	}

	.input:focus ~ svg {
		color: rgb(var(--color-primary));
	}

	.input:focus {
		box-shadow:
			0 0 0 1px rgb(var(--color-primary)),
			0 0 0 3px rgb(var(--color-primary) / 0.25);
	}

	.input:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.input.not-resizable {
		resize: none;
	}
</style>
