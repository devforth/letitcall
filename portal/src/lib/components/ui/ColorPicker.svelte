<script lang="ts">
	const hexPattern = '#[0-9A-Fa-f]{6}';

	let {
		id,
		label,
		value = $bindable()
	}: {
		id: string;
		label: string;
		value: string;
	} = $props();

	let text = $state(value);

	$effect(() => {
		text = value;
	});

	function pickColor(event: Event) {
		value = (event.currentTarget as HTMLInputElement).value.toUpperCase();
	}

	function typeColor(event: Event) {
		text = (event.currentTarget as HTMLInputElement).value;
		if (/^#[0-9A-Fa-f]{6}$/.test(text)) value = text.toUpperCase();
	}
</script>

<div class="picker">
	<label class="swatch" for={id} style={`background: ${value}`}>
		<span class="sr-only">Choose {label}</span>
		<input id={id} type="color" value={value} oninput={pickColor} />
	</label>
	<label class="sr-only" for={`${id}-hex`}>{label} hex value</label>
	<input
		id={`${id}-hex`}
		class="hex"
		type="text"
		value={text}
		oninput={typeColor}
		pattern={hexPattern}
		maxlength="7"
		required
		spellcheck="false"
		aria-label={`${label} hex value`}
	/>
</div>

<style>
	.picker {
		display: grid;
		grid-template-columns: 2.75rem minmax(0, 8rem);
		align-items: center;
		gap: 0.625rem;
	}

	.swatch {
		position: relative;
		display: block;
		width: 2.75rem;
		height: 2.75rem;
		border: 2px solid rgb(var(--color-border));
		border-radius: 10px;
		box-shadow: inset 0 0 0 2px rgb(var(--color-foreground));
		cursor: pointer;
		overflow: hidden;
	}

	.swatch:focus-within {
		outline: 2px solid rgb(var(--color-primary));
		outline-offset: 2px;
	}

	.swatch input {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		opacity: 0;
		cursor: pointer;
	}

	.hex {
		width: 100%;
		min-height: 2.75rem;
		padding: 0.625rem 0.75rem;
		border: 2px solid rgb(var(--color-border));
		border-radius: 10px;
		outline: none;
		background: rgb(var(--color-foreground));
		color: rgb(var(--color-text));
		font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
		font-size: 0.875rem;
		text-transform: uppercase;
	}

	.hex:focus {
		border-color: rgb(var(--color-primary));
		box-shadow: 0 0 0 3px rgb(var(--color-primary) / 0.2);
	}

	.hex:invalid {
		border-style: dashed;
	}

	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip: rect(0, 0, 0, 0);
		white-space: nowrap;
		border: 0;
	}
</style>
