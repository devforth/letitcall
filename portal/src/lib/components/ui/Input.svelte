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
		icon?: 'text' | 'email' | 'password' | 'search' | 'user';
	} = $props();

	const activeIcon = $derived(icon ?? type);
</script>

<label class="field" for={id}>
	<span class="field-label">{label}</span>
	<div class="input-group">
		<input
			{id}
			{type}
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
			class:has-error={!!error}
		/>
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
	</div>
	{#if error}
		<span id={`${id}-error`} class="field-error" role="alert">{error}</span>
	{/if}
</label>

<style>
	.field {
		display: grid;
		gap: 8px;
		font-size: 0.875rem;
	}

	.field-label {
		font-weight: 600;
		color: rgb(var(--color-text));
	}

	.input-group {
		position: relative;
	}

	.input-group svg {
		position: absolute;
		left: 12px;
		top: 50%;
		transform: translateY(-50%);
		width: 18px;
		height: 18px;
		color: rgb(var(--color-muted-foreground));
		pointer-events: none;
		transition: color 0.18s;
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

	.input::placeholder {
		color: rgb(var(--color-text) / 0.45);
	}

	.input:focus {
		border-color: rgb(var(--color-primary));
		box-shadow: 0 0 0 3px rgb(var(--color-primary) / 0.25);
	}

	.input:focus ~ svg {
		color: rgb(var(--color-primary));
	}

	.input.has-error {
		border-color: rgb(var(--error));
		box-shadow: 0 0 0 3px rgb(var(--error) / 0.15);
	}

	.input.has-error ~ svg {
		color: rgb(var(--error));
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
</style>
