<script lang="ts">
	import Icon from '@iconify/svelte';
	import chevronDownIcon from '@iconify-icons/tabler/chevron-down';
	import xIcon from '@iconify-icons/tabler/x';

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

	let open = $state(false);
	let query = $state('');

	const matchingOptions = $derived(
		query ? options.filter((option) => option.toLowerCase().includes(query.toLowerCase())) : options
	);

	function openOptions() {
		open = true;
		query = '';
	}

	function filterOptions(event: Event) {
		open = true;
		query = (event.currentTarget as HTMLInputElement).value;
	}

	function selectOption(option: string) {
		value = option;
		open = false;
		query = '';
	}

	function clearValue() {
		value = '';
		open = true;
		query = '';
	}

	function closeOptions(event: FocusEvent) {
		const field = event.currentTarget as HTMLDivElement;
		if (!field.contains(event.relatedTarget as Node | null)) open = false;
	}
</script>

<div class="field" onfocusout={closeOptions}>
	<div class="input-group" class:filled={!!value}>
		<input
			{id}
			type="search"
			bind:value
			{placeholder}
			{required}
			{disabled}
			autocomplete="off"
			role="combobox"
			aria-autocomplete="list"
			aria-controls={`${id}-options`}
			aria-expanded={open}
			onfocus={openOptions}
			oninput={filterOptions}
			onkeydown={(event) => {
				if (event.key === 'Escape') open = false;
			}}
			class="input"
		/>
		<label class="float-label" for={id}>{label}</label>
		{#if value}
			<button
				type="button"
				class="clear-button"
				aria-label={`Clear ${label.toLowerCase()}`}
				{disabled}
				onclick={clearValue}
			>
				<Icon icon={xIcon} width="16" height="16" class="clear-icon" />
			</button>
		{/if}
		<button
			type="button"
			class="select-toggle"
			aria-label={`Show ${label.toLowerCase()} options`}
			aria-expanded={open}
			disabled={disabled}
			onclick={() => {
				open = !open;
				query = '';
			}}
		>
			<Icon icon={chevronDownIcon} width="18" height="18" class={open ? 'open' : ''} />
		</button>
		{#if open}
			<div id={`${id}-options`} class="options" role="listbox" aria-label={`${label} options`}>
				{#each matchingOptions as option (option)}
					<button
						type="button"
						class:selected={option === value}
						class="option"
						role="option"
						aria-selected={option === value}
						onclick={() => selectOption(option)}
					>
						{option}
					</button>
				{:else}
					<p class="empty-options">No matching timezones</p>
				{/each}
			</div>
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
	}

	.input {
		width: 100%;
		font: inherit;
		font-size: 0.9rem;
		color: rgb(var(--color-text));
		background: rgb(var(--color-foreground));
		border: 2px solid rgb(var(--color-border));
		border-radius: 10px;
		padding: 10px 5rem 10px 12px;
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
		max-width: calc(100% - 5.25rem);
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

	.input::-webkit-search-cancel-button {
		appearance: none;
		-webkit-appearance: none;
	}

	.clear-button {
		display: grid;
		position: absolute;
		top: 50%;
		right: 2.5rem;
		width: 2rem;
		height: 2rem;
		place-items: center;
		transform: translateY(-50%);
		border: 0;
		border-radius: 999px;
		background: transparent;
		color: rgb(var(--color-muted-foreground));
		cursor: pointer;
		transition: background 0.18s, color 0.18s;
	}

	.clear-button:hover,
	.clear-button:focus-visible {
		color: rgb(var(--color-primary));
		outline: none;
	}

	.clear-button:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.clear-button :global(.clear-icon path) {
		stroke-width: 2.5;
	}

	.select-toggle {
		display: grid;
		position: absolute;
		top: 0;
		right: 0;
		width: 2.75rem;
		height: 100%;
		place-items: center;
		border: 0;
		background: transparent;
		color: rgb(var(--color-muted-foreground));
		cursor: pointer;
	}

	.select-toggle:hover,
	.select-toggle:focus-visible {
		color: rgb(var(--color-primary));
		outline: none;
	}

	.select-toggle:disabled {
		cursor: not-allowed;
	}

	.select-toggle :global(svg) {
		transition: transform 0.18s ease;
	}

	.select-toggle :global(svg.open) {
		transform: rotate(180deg);
	}

	.options {
		display: grid;
		position: absolute;
		z-index: 10;
		top: calc(100% + 0.5rem);
		right: 0;
		left: 0;
		max-height: 16rem;
		overflow-y: auto;
		padding: 0.375rem;
		border: 2px solid rgb(var(--color-border));
		border-radius: 10px;
		background: rgb(var(--color-foreground));
		box-shadow: var(--shadow-small);
		scrollbar-color: rgb(var(--color-border)) transparent;
		scrollbar-width: thin;
	}

	.options::-webkit-scrollbar {
		width: 0.75rem;
	}

	.options::-webkit-scrollbar-button {
		display: none;
		width: 0;
		height: 0;
	}

	.options::-webkit-scrollbar-track {
		background: transparent;
	}

	.options::-webkit-scrollbar-thumb {
		border: 3px solid rgb(var(--color-foreground));
		border-radius: 999px;
		background: rgb(var(--color-border));
	}

	.option {
		width: 100%;
		border: 0;
		border-radius: 6px;
		padding: 0.625rem 0.75rem;
		background: transparent;
		color: rgb(var(--color-text));
		font: inherit;
		font-size: 0.8125rem;
		text-align: left;
		cursor: pointer;
	}

	.option:hover,
	.option.selected {
		background: rgb(var(--color-primary) / 0.12);
		color: rgb(var(--color-primary));
	}

	.empty-options {
		margin: 0;
		padding: 0.625rem 0.75rem;
		font-size: 0.8125rem;
		color: rgb(var(--color-muted-foreground));
	}
</style>
