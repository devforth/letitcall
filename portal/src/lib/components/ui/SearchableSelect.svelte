<script lang="ts">
	import Icon, { type IconifyIcon } from '@iconify/svelte';
	import chevronDownIcon from '@iconify-icons/tabler/chevron-down';
	import xIcon from '@iconify-icons/tabler/x';

	type Option = { value: string; label: string; disabled?: boolean };

	let {
		id,
		label,
		options,
		value = $bindable(''),
		placeholder = 'Search…',
		required = false,
		disabled = false,
		icon,
		emptyText = 'No matches',
		placement = 'bottom',
		onchange
	}: {
		id: string;
		label: string;
		/** Plain strings when the value is the text you show, pairs when it isn't. */
		options: (string | Option)[];
		value?: string;
		placeholder?: string;
		required?: boolean;
		disabled?: boolean;
		icon?: IconifyIcon;
		/** Shown in place of the list when the query matches nothing. */
		emptyText?: string;
		/**
		 * Which side of the field the list opens on. Use 'top' where the field sits at the
		 * bottom of a scrolling panel, since a list opening downward is clipped there.
		 */
		placement?: 'bottom' | 'top';
		/**
		 * The chosen value, for callers that need more than an assignment. Fires on
		 * selection and on clear, never while typing, so it is safe to run side effects.
		 */
		onchange?: (value: string) => void;
	} = $props();

	let open = $state(false);
	// null while the field is showing its selection, a string once the user types. That
	// distinction is what lets focus show the current label against the *whole* list
	// instead of filtering the list down to the thing already chosen.
	let query = $state<string | null>(null);

	const normalized = $derived(
		options.map((option) => (typeof option === 'string' ? { value: option, label: option } : option))
	);
	const selectedLabel = $derived(normalized.find((option) => option.value === value)?.label ?? '');
	// The input is display-only: it renders the query while searching and the selection
	// otherwise, so `value` only ever holds a real option — never half-typed text.
	const text = $derived(query ?? selectedLabel);
	const matchingOptions = $derived.by(() => {
		if (!query) return normalized;
		const needle = query.toLowerCase();
		return normalized.filter((option) => option.label.toLowerCase().includes(needle));
	});

	function openOptions() {
		open = true;
		query = null;
	}

	function filterOptions(event: Event) {
		open = true;
		query = (event.currentTarget as HTMLInputElement).value;
	}

	function selectOption(option: Option) {
		value = option.value;
		open = false;
		query = null;
		onchange?.(option.value);
	}

	function clearValue() {
		value = '';
		open = true;
		query = null;
		onchange?.('');
	}

	function closeOptions(event: FocusEvent) {
		const field = event.currentTarget as HTMLDivElement;
		if (field.contains(event.relatedTarget as Node | null)) return;
		open = false;
		// Abandoned search text goes back to showing the selection it never replaced.
		query = null;
	}
</script>

<div class="field" onfocusout={closeOptions}>
	<div class="input-group" class:filled={!!text} class:has-icon={!!icon}>
		<input
			{id}
			type="search"
			value={text}
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
				if (event.key !== 'Escape') return;
				open = false;
				query = null;
			}}
			class="input"
		/>
		<label class="float-label" for={id}>{label}</label>
		{#if icon}
			<span class="lead-icon"><Icon {icon} width="18" height="18" /></span>
		{/if}
		<div class="actions">
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
					query = null;
				}}
			>
				<Icon icon={chevronDownIcon} width="18" height="18" class={open ? 'open' : ''} />
			</button>
		</div>
		{#if open}
			<div
				id={`${id}-options`}
				class="options"
				class:above={placement === 'top'}
				role="listbox"
				aria-label={`${label} options`}
			>
				{#each matchingOptions as option (option.value)}
					<button
						type="button"
						class:selected={option.value === value}
						class="option"
						role="option"
						aria-selected={option.value === value}
						disabled={option.disabled}
						onclick={() => selectOption(option)}
					>
						{option.label}
					</button>
				{:else}
					<p class="empty-options">{emptyText}</p>
				{/each}
			</div>
		{/if}
	</div>
</div>

<style>
	.field {
		display: grid;
		align-self: start;
		font-size: 0.875rem;
	}

	.input-group {
		position: relative;
		align-self: start;
		margin: 3px;
	}

	.input {
		width: 100%;
		font: inherit;
		font-size: 0.9rem;
		color: rgb(var(--color-text));
		background: rgb(var(--color-foreground));
		border: 0;
		border-radius: 10px;
		padding: 10px 5rem 10px 12px;
		min-height: 44px;
		outline: none;
		box-shadow: 0 0 0 1px rgb(var(--color-border));
		transition: box-shadow 0.18s;
	}

	.has-icon .input {
		padding-left: 40px;
	}

	.lead-icon {
		position: absolute;
		left: 12px;
		top: 50%;
		transform: translateY(-50%);
		display: grid;
		place-items: center;
		width: 18px;
		height: 18px;
		color: rgb(var(--color-text));
		pointer-events: none;
		transition: color 0.18s;
	}

	.input:focus ~ .lead-icon {
		color: rgb(var(--color-primary));
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

	.has-icon .float-label {
		left: 40px;
		max-width: calc(100% - 40px - 4.75rem);
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
		box-shadow:
			0 0 0 1px rgb(var(--color-primary)),
			0 0 0 3px rgb(var(--color-primary) / 0.25);
	}

	.input:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.input::-webkit-search-cancel-button {
		appearance: none;
		-webkit-appearance: none;
	}

	.actions {
		display: flex;
		position: absolute;
		inset-block: 0;
		right: 0;
		align-items: center;
	}

	.clear-button {
		display: grid;
		width: 2rem;
		height: 2rem;
		place-items: center;
		border: 0;
		border-radius: 999px;
		background: transparent;
		color: rgb(var(--color-text) / 0.65);
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
		width: 2.75rem;
		height: 2rem;
		place-items: center;
		border: 0;
		background: transparent;
		color: rgb(var(--color-text) / 0.65);
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
		border: 0;
		border-radius: 10px;
		background: rgb(var(--color-foreground));
		box-shadow: 0 0 0 1px rgb(var(--color-border)), var(--shadow-small);
		scrollbar-color: rgb(var(--color-border)) transparent;
		scrollbar-width: thin;
	}

	/* Opens over the field instead of under it. Two classes, so it wins on specificity
	   rather than on source order. */
	.options.above {
		top: auto;
		bottom: calc(100% + 0.5rem);
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

	.option:hover:not(:disabled),
	.option.selected {
		background: rgb(var(--color-primary) / 0.12);
		color: rgb(var(--color-primary));
	}

	.option:focus-visible {
		outline: 2px solid rgb(var(--color-primary));
		outline-offset: -2px;
	}

	/* Listed but unpickable — the dropdown's equivalent of a locked slot cell. */
	.option:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.empty-options {
		margin: 0;
		padding: 0.625rem 0.75rem;
		font-size: 0.8125rem;
		color: rgb(var(--color-text) / 0.65);
	}
</style>
