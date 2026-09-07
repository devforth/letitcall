<script lang="ts">
	import Icon, { type IconifyIcon } from '@iconify/svelte';

	let {
		id,
		label,
		value = $bindable(''),
		min = 1,
		max,
		placeholder = '',
		disabled = false,
		required = false,
		icon
	}: {
		id: string;
		label: string;
		value?: string;
		min?: number;
		max?: number;
		placeholder?: string;
		disabled?: boolean;
		required?: boolean;
		icon?: IconifyIcon;
	} = $props();
</script>

<div class="field">
	<div class="input-group" class:filled={value !== ''} class:has-icon={!!icon}>
		<input
			{id}
			type="number"
			bind:value
			{min}
			{max}
			{placeholder}
			{disabled}
			{required}
			step="1"
			class="input"
		/>
		<label class="float-label" for={id}>{label}</label>
		{#if icon}
			<span class="lead-icon"><Icon {icon} width="18" height="18" /></span>
		{/if}
	</div>
</div>

<style>
	.field {
		display: grid;
		font-size: 0.875rem;
	}

	.input-group {
		position: relative;
		margin: 3px;
	}

	.input {
		width: 100%;
		min-height: 44px;
		border: 0;
		border-radius: 10px;
		padding: 10px 12px;
		outline: none;
		background: rgb(var(--color-foreground));
		box-shadow: 0 0 0 1px rgb(var(--color-border));
		color: rgb(var(--color-text));
		font: inherit;
		font-size: 0.9rem;
		transition: box-shadow 0.18s;
	}

	.has-icon .input {
		padding-left: 40px;
	}

	.input::placeholder {
		color: transparent;
	}

	.input:focus::placeholder {
		color: rgb(var(--color-text) / 0.4);
	}

	.float-label {
		position: absolute;
		left: 12px;
		top: 50%;
		transform: translateY(-50%);
		max-width: calc(100% - 52px);
		overflow: hidden;
		padding: 0 4px;
		background: rgb(var(--color-foreground));
		color: rgb(var(--color-text));
		font-size: 0.9rem;
		font-weight: 400;
		opacity: 0.4;
		pointer-events: none;
		text-overflow: ellipsis;
		white-space: nowrap;
		transition: top 0.16s, left 0.16s, font-size 0.16s, color 0.16s, opacity 0.16s;
	}

	.has-icon .float-label {
		left: 40px;
	}

	.input:focus ~ .float-label,
	.filled .float-label {
		top: 0;
		left: 10px;
		color: rgb(var(--color-primary));
		font-size: 0.72rem;
		opacity: 1;
	}

	.lead-icon {
		position: absolute;
		left: 12px;
		top: 50%;
		display: grid;
		transform: translateY(-50%);
		place-items: center;
		color: rgb(var(--color-text));
		pointer-events: none;
		transition: color 0.18s;
	}

	.input:focus ~ .lead-icon {
		color: rgb(var(--color-primary));
	}

	.input:focus {
		box-shadow:
			0 0 0 1px rgb(var(--color-primary)),
			0 0 0 3px rgb(var(--color-primary) / 0.25);
	}

	.input:disabled {
		cursor: not-allowed;
		opacity: 0.4;
	}
</style>
