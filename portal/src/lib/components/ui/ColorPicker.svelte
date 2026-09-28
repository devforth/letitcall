<script lang="ts">
	let {
		id,
		label,
		description,
		value = $bindable(),
		size = 'compact',
		onchange
	}: {
		id: string;
		label: string;
		description: string;
		value: string;
		size?: 'brand' | 'compact';
		onchange?: (value: string) => void;
	} = $props();

	let draft = $state(value);

	$effect(() => {
		draft = value;
	});

	function pickColor(event: Event) {
		value = (event.currentTarget as HTMLInputElement).value.toUpperCase();
		onchange?.(value);
	}

	function enterColor(event: Event) {
		draft = (event.currentTarget as HTMLInputElement).value.toUpperCase();
		if (/^#[0-9A-F]{6}$/.test(draft)) {
			value = draft;
			onchange?.(value);
		}
	}
</script>

<div class:brand={size === 'brand'} class="picker">
	<div class="color-copy">
		<label for={id}>{label}</label>
		<p>{description}</p>
	</div>
	<div class="swatch" style={`background: ${value};`}>
		<input
			id={`${id}-swatch`}
			class="color-input"
			type="color"
			value={value}
			aria-label={`Choose ${label}`}
			oninput={pickColor}
		/>
	</div>
	<input
		{id}
		class="hex"
		type="text"
		bind:value={draft}
		maxlength="7"
		inputmode="text"
		spellcheck="false"
		aria-label={`${label} hex value`}
		oninput={enterColor}
		onblur={() => (draft = value)}
	/>
</div>

<style>
	.picker {
		display: grid;
		grid-template-columns: 3.25rem minmax(0, 1fr) 7rem;
		align-items: center;
		gap: 0.75rem;
		width: 100%;
	}

	.picker.brand {
		grid-template-columns: minmax(0, 1fr) 7rem;
	}

	.color-copy {
		min-width: 0;
		grid-column: 2;
	}

	.brand .color-copy {
		grid-column: 1;
	}

	.color-copy label,
	.color-copy p {
		display: block;
		margin: 0;
		line-height: 1.25;
	}

	.color-copy label {
		font-size: 1rem;
		font-weight: 600;
	}

	.color-copy p {
		color: rgb(var(--color-text) / 0.65);
		font-size: 0.875rem;
	}

	.swatch {
		position: relative;
		grid-column: 1;
		grid-row: 1;
		width: 3.25rem;
		height: 3.25rem;
		border-radius: 0.75rem;
		box-shadow: inset 0 0 0 1px var(--color-border);
		overflow: hidden;
	}

	.brand .swatch {
		grid-column: 1 / -1;
		grid-row: 2;
		width: 100%;
		height: 4.5rem;
	}

	.color-input {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		opacity: 0;
		cursor: pointer;
	}

	.swatch:focus-within {
		box-shadow:
			inset 0 0 0 1px rgb(var(--color-primary)),
			0 0 0 3px rgb(var(--color-primary) / 0.25);
	}

	.hex {
		grid-column: 3;
		width: 100%;
		height: 2.75rem;
		padding: 0.5rem 0.75rem;
		border: 1px solid var(--color-border);
		border-radius: 0.75rem;
		outline: none;
		background: rgb(var(--color-background));
		color: rgb(var(--color-text));
		font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
		font-size: 1rem;
		font-weight: 400;
		text-transform: uppercase;
	}

	.brand .hex {
		grid-column: 2;
		grid-row: 1;
	}

	.hex:focus {
		border-color: rgb(var(--color-primary));
		box-shadow: 0 0 0 3px rgb(var(--color-primary) / 0.25);
	}

	@media (max-width: 420px) {
		.picker {
			grid-template-columns: 3.25rem minmax(0, 1fr) 6.5rem;
			gap: 0.625rem;
		}
	}
</style>
