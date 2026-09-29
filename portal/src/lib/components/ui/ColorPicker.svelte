<script lang="ts">
	import Icon from '@iconify/svelte';
	import colorSwatchIcon from '@iconify-icons/tabler/color-swatch';
	import { accessibleTextColor } from '$lib/color-contrast';
	import CopyButton from '$lib/components/ui/CopyButton.svelte';

	let {
		id,
		label,
		description,
		value = $bindable(),
		onchange
	}: {
		id: string;
		label: string;
		description: string;
		value: string;
		onchange?: (value: string) => void;
	} = $props();

	let draft = $state(value);
	let swatchText = $derived(accessibleTextColor([value]));

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

<div class="picker">
	<div class="swatch" style={`background: ${value};`}>
		<input
			id={`${id}-swatch`}
			class="color-input"
			type="color"
			value={value}
			aria-label={`Choose ${label}`}
			oninput={pickColor}
		/>
		<span class="swatch-label" style={`color: ${swatchText};`}>{label}</span>
		<span class="swatch-icon" style={`color: ${swatchText};`} aria-hidden="true">
			<Icon icon={colorSwatchIcon} width="20" height="20" />
		</span>
	</div>
	<div class="hex-field">
		<input
			{id}
			class="hex"
			type="text"
			bind:value={draft}
			maxlength="7"
			inputmode="text"
			spellcheck="false"
			oninput={enterColor}
			onblur={() => (draft = value)}
		/>
		<label class="hex-label" for={id}>{description}</label>
		<CopyButton class="hex-copy" value={draft} label={`Copy ${label} hex value`} />
	</div>
</div>

<style>
	.picker {
		display: grid;
		grid-template-columns: minmax(0, 1fr);
		align-items: center;
		gap: 0.5rem;
		width: 100%;
	}

	.swatch {
		position: relative;
		grid-column: 1;
		grid-row: 1;
		width: 100%;
		height: 4.5rem;
		border-radius: 0.75rem;
		box-shadow: inset 0 0 0 1px var(--color-border);
		overflow: hidden;
	}

	.swatch-label {
		position: absolute;
		bottom: 0.625rem;
		left: 1rem;
		font-size: 1rem;
		font-weight: 400;
		line-height: 1.25;
		pointer-events: none;
	}

	.swatch-icon {
		position: absolute;
		right: 1rem;
		bottom: 0.625rem;
		display: grid;
		place-items: center;
		opacity: 0.4;
		pointer-events: none;
		transition: opacity 0.18s;
	}

	.swatch:hover .swatch-icon,
	.swatch:focus-within .swatch-icon {
		opacity: 1;
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

	.hex-field {
		position: relative;
		grid-column: 1;
		grid-row: 2;
		width: calc(100% - 0.375rem);
		height: 2.75rem;
		margin: 0.1875rem;
	}

	.hex {
		width: 100%;
		height: 100%;
		padding: 0.5rem 3.25rem 0.5rem 0.75rem;
		border: 0;
		border-radius: 0.625rem;
		outline: none;
		background: rgb(var(--color-background));
		color: rgb(var(--color-text));
		box-shadow: 0 0 0 1px rgb(var(--color-primary));
		font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
		font-size: 1rem;
		font-weight: 400;
		text-transform: uppercase;
		transition: box-shadow 0.18s;
	}

	.hex-label {
		position: absolute;
		top: 0;
		left: 0.625rem;
		max-width: calc(100% - 4rem);
		overflow: hidden;
		padding: 0 0.25rem;
		transform: translateY(-50%);
		background: rgb(var(--color-background));
		color: rgb(var(--color-text));
		font-size: 0.72rem;
		line-height: 1;
		text-overflow: ellipsis;
		white-space: nowrap;
		pointer-events: none;
	}

	.hex-field :global(.hex-copy) {
		position: absolute;
		top: 0;
		right: 0;
		width: 2.75rem;
		height: 100%;
		border-radius: 0.625rem;
		background: transparent;
	}

	.hex-field :global(.hex-copy:hover:not(:disabled)),
	.hex-field :global(.hex-copy:active:not(:disabled)) {
		background: transparent;
		color: rgb(var(--color-primary));
	}

	.hex:focus {
		box-shadow:
			0 0 0 1px rgb(var(--color-primary)),
			0 0 0 3px rgb(var(--color-primary) / 0.25);
	}
</style>
