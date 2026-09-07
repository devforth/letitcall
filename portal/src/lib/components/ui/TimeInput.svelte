<script lang="ts">
	import Icon from '@iconify/svelte';
	import clockIcon from '@iconify-icons/tabler/clock';

	let {
		id,
		label,
		value = $bindable(''),
		disabled = false,
		onchange
	}: {
		id: string;
		label: string;
		value?: string;
		disabled?: boolean;
		onchange?: (value: string) => void;
	} = $props();
</script>

<div class="field">
	<div class="input-group">
		<input
			{id}
			type="time"
			bind:value
			oninput={(event) => onchange?.(event.currentTarget.value)}
			{disabled}
			step="900"
			class="input"
		/>
		<label class="float-label" for={id}>{label}</label>
		<span class="lead-icon" aria-hidden="true">
			<Icon icon={clockIcon} width="18" height="18" />
		</span>
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
		min-height: 38px;
		border: 0;
		border-radius: 10px;
		padding: 9px 12px 9px 40px;
		outline: none;
		background: rgb(var(--color-foreground));
		box-shadow: 0 0 0 1px rgb(var(--color-border));
		color: rgb(var(--color-text));
		font: inherit;
		font-size: 0.9rem;
		transition: box-shadow 0.18s;
	}

	.input::-webkit-calendar-picker-indicator {
		display: none;
	}

	.float-label {
		position: absolute;
		left: 10px;
		top: 0;
		transform: translateY(-50%);
		padding: 0 4px;
		background: rgb(var(--color-foreground));
		color: rgb(var(--color-primary));
		font-size: 0.72rem;
		pointer-events: none;
		transition: top 0.16s, left 0.16s, font-size 0.16s, color 0.16s, opacity 0.16s;
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
