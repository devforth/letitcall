<script lang="ts">
	let {
		id,
		label,
		options,
		value = $bindable(''),
		placeholder = 'Search…',
		required = false,
		disabled = false
	}: {
		id: string;
		label: string;
		options: string[];
		value?: string;
		placeholder?: string;
		required?: boolean;
		disabled?: boolean;
	} = $props();
</script>

<div class="field">
	<div class="input-group" class:filled={!!value}>
		<input
			{id}
			list={`${id}-options`}
			type="search"
			bind:value
			{placeholder}
			{required}
			{disabled}
			autocomplete="off"
			class="input"
		/>
		<label class="float-label" for={id}>{label}</label>
	</div>
	<datalist id={`${id}-options`}>
		{#each options as option (option)}
			<option value={option}></option>
		{/each}
	</datalist>
</div>

<style>
	.field {
		display: grid;
		font-size: 0.875rem;
	}

	.input-group {
		position: relative;
	}

	.input {
		width: 100%;
		font: inherit;
		font-size: 0.9rem;
		color: rgb(var(--color-text));
		background: rgb(var(--color-foreground));
		border: 2px solid rgb(var(--color-border));
		border-radius: 10px;
		padding: 10px 12px;
		min-height: 44px;
		outline: none;
		transition: border-color 0.18s, box-shadow 0.18s;
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
		max-width: calc(100% - 24px);
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

	.input:focus {
		border-color: rgb(var(--color-primary));
		box-shadow: 0 0 0 3px rgb(var(--color-primary) / 0.25);
	}

	.input:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}
</style>
