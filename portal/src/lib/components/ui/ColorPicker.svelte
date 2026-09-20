<script lang="ts">
	let {
		id,
		label,
		value = $bindable()
	}: {
		id: string;
		label: string;
		value: string;
	} = $props();

	const textColor = $derived(Number.parseInt(value.slice(1, 3), 16) * 299 + Number.parseInt(value.slice(3, 5), 16) * 587 + Number.parseInt(value.slice(5, 7), 16) * 114 > 160000 ? '#111111' : '#ffffff');

	function pickColor(event: Event) {
		value = (event.currentTarget as HTMLInputElement).value.toUpperCase();
	}
</script>

<div class="picker">
	<label class="swatch" for={id} style={`background: ${value}; color: ${textColor};`}>
		<span class="sr-only">Choose {label}</span>
		<span class="hex" aria-hidden="true">{value}</span>
		<input id={id} type="color" value={value} oninput={pickColor} />
	</label>
</div>

<style>
	.picker {
		width: 11rem;
	}

	.swatch {
		position: relative;
		display: flex;
		width: 100%;
		height: 4.5rem;
		align-items: flex-end;
		justify-content: space-between;
		border: 0;
		border-radius: 0.75rem;
		padding: 0.75rem;
		box-shadow:
			inset 0 0 0 1px rgb(var(--color-primary)),
			inset 0 0 0 3px rgb(var(--color-contrast-text));
		cursor: pointer;
		overflow: hidden;
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
		position: relative;
		z-index: 1;
		pointer-events: none;
		font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
		font-size: 1rem;
		font-weight: 400;
		text-transform: uppercase;
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
