<script lang="ts">
	import type { HTMLInputAttributes } from 'svelte/elements';

	let {
		id,
		label,
		type = 'text',
		value = $bindable(''),
		placeholder = '',
		required = false,
		autocomplete,
		disabled = false,
		readonly = false,
		minlength,
		error = '',
		hint = '',
		icon
	}: {
		id: string;
		label: string;
		type?: 'text' | 'email' | 'password' | 'search';
		value?: string;
		placeholder?: string;
		required?: boolean;
		autocomplete?: HTMLInputAttributes['autocomplete'];
		disabled?: boolean;
		readonly?: boolean;
		minlength?: number;
		error?: string;
		hint?: string;
		icon?: 'text' | 'email' | 'password' | 'search' | 'user';
	} = $props();

	const activeIcon = $derived(icon ?? type);

	let revealed = $state(false);
	let inputEl = $state<HTMLInputElement>();

	// `type` can't be a dynamic attribute alongside bind:value, so set it imperatively
	$effect(() => {
		if (inputEl) inputEl.type = type === 'password' && revealed ? 'text' : type;
	});
</script>

<div class="field">
	<div class="input-group" class:filled={!!value} class:has-error={!!error}>
		<input
			{id}
			bind:this={inputEl}
			bind:value
			{placeholder}
			{required}
			{autocomplete}
			{disabled}
			{readonly}
			{minlength}
			aria-invalid={error ? 'true' : undefined}
			aria-describedby={error ? `${id}-error` : undefined}
			class="input"
			class:has-trailing={type === 'password'}
		/>
		<label class="float-label" for={id}>{label}</label>
		{#if activeIcon === 'email'}
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="3" y="5" width="18" height="14" rx="2" /><path d="m3 7 9 6 9-6" /></svg>
		{:else if activeIcon === 'password'}
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="4" y="10" width="16" height="11" rx="2" /><path d="M8 10V7a4 4 0 0 1 8 0v3" /></svg>
		{:else if activeIcon === 'search'}
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="11" cy="11" r="7" /><path d="m21 21-4.3-4.3" /></svg>
		{:else if activeIcon === 'user'}
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="12" cy="8" r="4" /><path d="M4 21a8 8 0 0 1 16 0" /></svg>
		{:else}
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><line x1="4" y1="8" x2="20" y2="8" /><line x1="4" y1="12" x2="14" y2="12" /><line x1="4" y1="16" x2="18" y2="16" /></svg>
		{/if}
		{#if type === 'password'}
			<button
				type="button"
				class="reveal"
				onclick={() => (revealed = !revealed)}
				aria-label={revealed ? 'Hide password' : 'Show password'}
				aria-pressed={revealed}
				tabindex="-1"
			>
				{#if revealed}
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M17.94 17.94A10.1 10.1 0 0 1 12 20c-7 0-10-8-10-8a18.5 18.5 0 0 1 5.06-5.94M9.9 4.24A9.1 9.1 0 0 1 12 4c7 0 10 8 10 8a18.5 18.5 0 0 1-2.16 3.19M9.9 9.9a3 3 0 0 0 4.2 4.2" /><line x1="1" y1="1" x2="23" y2="23" /></svg>
				{:else}
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7-10-7-10-7z" /><circle cx="12" cy="12" r="3" /></svg>
				{/if}
			</button>
		{/if}
	</div>
	{#if error}
		<span id={`${id}-error`} class="field-error" role="alert">{error}</span>
	{:else if hint}
		<span class="field-hint">{hint}</span>
	{/if}
</div>

<style>
	.field {
		display: grid;
		gap: 6px;
		font-size: 0.875rem;
	}

	.input-group {
		position: relative;
	}

	.input-group > svg {
		position: absolute;
		left: 12px;
		top: 50%;
		transform: translateY(-50%);
		width: 18px;
		height: 18px;
		color: rgb(var(--color-text));
		pointer-events: none;
		transition: color 0.18s;
	}

	.reveal {
		position: absolute;
		right: 8px;
		top: 50%;
		transform: translateY(-50%);
		display: grid;
		place-items: center;
		width: 30px;
		height: 30px;
		padding: 0;
		border: 0;
		border-radius: 8px;
		background: transparent;
		color: rgb(var(--color-muted-foreground));
		cursor: pointer;
		transition: color 0.18s, background 0.18s;
	}

	.reveal:hover {
		color: rgb(var(--color-primary));
		background: rgb(var(--color-primary) / 0.1);
	}

	.reveal svg {
		width: 18px;
		height: 18px;
	}

	.input {
		width: 100%;
		font: inherit;
		font-size: 0.9rem;
		color: rgb(var(--color-text));
		background: rgb(var(--color-foreground));
		border: 2px solid rgb(var(--color-border));
		border-radius: 10px;
		padding: 10px 12px 10px 40px;
		min-height: 44px;
		outline: none;
		transition: border-color 0.18s, box-shadow 0.18s;
	}

	.input.has-trailing {
		padding-right: 44px;
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
		top: 50%;
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
	.input:-webkit-autofill ~ .float-label,
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
		border-color: rgb(var(--color-primary));
		box-shadow: 0 0 0 3px rgb(var(--color-primary) / 0.25);
	}

	.has-error .input {
		border-color: rgb(var(--error));
		box-shadow: 0 0 0 3px rgb(var(--error) / 0.15);
	}

	.has-error > svg,
	.has-error .float-label {
		color: rgb(var(--error));
	}

	/* neutralize the browser autofill background/text so it matches the field */
	.input:-webkit-autofill,
	.input:-webkit-autofill:hover {
		-webkit-box-shadow: 0 0 0 1000px rgb(var(--color-foreground)) inset !important;
		-webkit-text-fill-color: rgb(var(--color-text)) !important;
		caret-color: rgb(var(--color-text));
		border-color: rgb(var(--color-border));
		transition: background-color 9999s ease-in-out 0s;
	}

	.input:-webkit-autofill:focus {
		-webkit-box-shadow:
			0 0 0 1000px rgb(var(--color-foreground)) inset,
			0 0 0 3px rgb(var(--color-primary) / 0.25) !important;
		border-color: rgb(var(--color-primary));
	}

	.has-error .input:-webkit-autofill {
		-webkit-box-shadow: 0 0 0 1000px rgb(var(--color-foreground)) inset !important;
		border-color: rgb(var(--error));
	}

	.input:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.field-error {
		font-size: 0.75rem;
		font-weight: 600;
		color: rgb(var(--error));
	}

	.field-hint {
		font-size: 0.75rem;
		color: var(--color-text);
	}
</style>
