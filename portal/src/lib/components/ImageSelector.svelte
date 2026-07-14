<script lang="ts">
	import { onDestroy, tick } from 'svelte';
	import Cropper from 'cropperjs';
	import Button from '$lib/components/ui/Button.svelte';

	let { id, legend }: { id: string; legend: string } = $props();

	const imageTemplate = `
		<cropper-canvas background scale-step="0.1">
			<cropper-image initial-center-size="cover" scalable translatable></cropper-image>
			<cropper-handle action="move" plain></cropper-handle>
			<cropper-selection initial-aspect-ratio="1" aspect-ratio="1" initial-coverage="0.8" theme-color="#000" outlined>
				<cropper-grid role="grid" covered theme-color="rgba(0, 0, 0, 0.45)"></cropper-grid>
				<cropper-crosshair centered theme-color="#000"></cropper-crosshair>
				<cropper-handle action="move" theme-color="rgba(0, 0, 0, 0.35)"></cropper-handle>
			</cropper-selection>
		</cropper-canvas>
	`;

	let container = $state<HTMLDivElement>();
	let image = $state<HTMLImageElement>();
	let cropper: Cropper | null = null;
	let source = $state('');
	let filename = $state('');

	async function selectImage(event: Event) {
		const file = (event.currentTarget as HTMLInputElement).files?.[0];
		if (!file) return;
		destroyCropper();
		source = URL.createObjectURL(file);
		filename = file.name;
		await tick();
		if (!container || !image) return;
		cropper = new Cropper(image, { container, template: imageTemplate });
		await cropper.getCropperImage()?.$ready();
	}

	function zoom(amount: number) {
		cropper?.getCropperImage()?.$zoom(amount);
	}

	function resetCrop() {
		cropper?.getCropperImage()?.$resetTransform().$center('cover');
		cropper?.getCropperSelection()?.$reset();
	}

	function destroyCropper() {
		cropper?.destroy();
		cropper = null;
		if (source) URL.revokeObjectURL(source);
	}

	export async function exportImage(): Promise<string> {
		const selection = cropper?.getCropperSelection();
		if (!selection) return '';
		const canvas = await selection.$toCanvas({
			width: 512,
			height: 512,
			beforeDraw: (context, output) => {
				context.fillStyle = '#fff';
				context.fillRect(0, 0, output.width, output.height);
			}
		});
		return canvas.toDataURL('image/jpeg', 0.9);
	}

	onDestroy(destroyCropper);
</script>

<fieldset class="fieldset grid gap-3">
	<legend class="field-label px-2 text-sm">{legend}</legend>
	<label class="field" for={id}>
		<span class="field-label">Choose image</span>
		<input
			{id}
			type="file"
			accept="image/jpeg,image/png,image/webp"
			onchange={selectImage}
			class="file-input"
		/>
	</label>
	{#if source}
		<p class="text-xs">Drag to pan, use the frame to crop, and scroll or use the buttons to zoom.</p>
		<div class="cropper-host" bind:this={container}>
			<img bind:this={image} src={source} alt={`Crop ${filename}`} />
		</div>
		<div class="flex flex-wrap gap-2">
			<Button variant="secondary" onclick={() => zoom(-0.1)}>Zoom out</Button>
			<Button variant="secondary" onclick={() => zoom(0.1)}>Zoom in</Button>
			<Button variant="secondary" onclick={resetCrop}>Reset</Button>
		</div>
	{/if}
</fieldset>

<style>
	.fieldset {
		border: 2px solid rgb(var(--color-border));
		border-radius: 12px;
		padding: 1rem;
	}

	.field {
		display: grid;
		gap: 6px;
		font-size: 0.875rem;
	}

	.field-label {
		font-weight: 600;
		color: rgb(var(--color-text));
	}

	.file-input {
		width: 100%;
		min-height: 44px;
		font: inherit;
		font-size: 0.9rem;
		color: rgb(var(--color-text));
		background: rgb(var(--color-foreground));
		border: 2px solid rgb(var(--color-border));
		border-radius: 10px;
		padding: 6px 10px;
		outline: none;
		transition: border-color 0.18s, box-shadow 0.18s;
	}

	.file-input:focus {
		border-color: rgb(var(--color-primary));
		box-shadow: 0 0 0 3px rgb(var(--color-primary) / 0.25);
	}

	.file-input::file-selector-button {
		margin-right: 12px;
		padding: 6px 12px;
		border: 2px solid rgb(var(--color-primary));
		border-radius: 8px;
		background: rgb(var(--color-foreground));
		color: rgb(var(--color-text));
		font: inherit;
		font-size: 0.85rem;
		font-weight: 700;
		cursor: pointer;
		transition: background 0.18s;
	}

	.file-input::file-selector-button:hover {
		background: rgb(var(--color-primary) / 0.1);
	}

	.cropper-host {
		height: 20rem;
		border: 2px solid rgb(var(--color-border));
		border-radius: 12px;
	}

	.cropper-host :global(cropper-canvas) {
		height: 100%;
	}
</style>
