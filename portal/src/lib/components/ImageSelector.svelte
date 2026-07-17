<script lang="ts">
	import { onDestroy, tick } from 'svelte';
	import Cropper from 'cropperjs';
	import Icon from '@iconify/svelte';
	import photoIcon from '@iconify-icons/tabler/photo';
	import refreshIcon from '@iconify-icons/tabler/refresh';
	import rotateClockwiseIcon from '@iconify-icons/tabler/rotate-clockwise';
	import uploadIcon from '@iconify-icons/tabler/upload';
	import zoomInIcon from '@iconify-icons/tabler/zoom-in';
	import zoomOutIcon from '@iconify-icons/tabler/zoom-out';
	import editIcon from '@iconify-icons/tabler/edit';
	import trashIcon from '@iconify-icons/tabler/trash';
	import Button from '$lib/components/ui/Button.svelte';
	import IconButton from '$lib/components/ui/IconButton.svelte';

	let {
		id,
		legend,
		current = '',
		ondelete
	}: { id: string; legend: string; current?: string; ondelete?: () => void } = $props();

	let editing = $state(false);

	const imageTemplate = `
		<cropper-canvas background scale-step="0.1">
			<cropper-image initial-center-size="cover" rotatable scalable translatable></cropper-image>
			<cropper-handle action="move" plain></cropper-handle>
			<cropper-selection class="round-selection" initial-aspect-ratio="1" aspect-ratio="1" initial-coverage="0.8" theme-color="#000" outlined>
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
	let isDragOver = $state(false);

	async function selectImage(event: Event) {
		const file = (event.currentTarget as HTMLInputElement).files?.[0];
		if (!file) return;
		await loadImage(file);
	}

	async function loadImage(file: File) {
		destroyCropper();
		source = URL.createObjectURL(file);
		filename = file.name;
		await tick();
		if (!container || !image) return;
		cropper = new Cropper(image, { container, template: imageTemplate });
		await cropper.getCropperImage()?.$ready();
	}

	function dragOver(event: DragEvent) {
		event.preventDefault();
		isDragOver = true;
	}

	function dragLeave() {
		isDragOver = false;
	}

	async function dropImage(event: DragEvent) {
		event.preventDefault();
		isDragOver = false;
		const file = event.dataTransfer?.files[0];
		if (!file) return;
		await loadImage(file);
	}

	function zoom(amount: number) {
		cropper?.getCropperImage()?.$zoom(amount);
	}

	function rotate(degrees: number) {
		cropper?.getCropperImage()?.$rotate(`${degrees}deg`);
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
		// Backend requires a 512×512 JPEG; the round look is applied cosmetically in the UI.
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

<fieldset class="image-selector">
	<legend class="selector-legend">{legend}</legend>
	{#if current && !source && !editing}
		<div class="current-avatar">
			<img src={current} alt={`Current ${legend.toLowerCase()}`} />
			<div class="current-actions">
				<IconButton tone="primary" label={`Edit ${legend.toLowerCase()}`} onclick={() => (editing = true)}>
					<Icon icon={editIcon} width="20" height="20" />
				</IconButton>
				<IconButton tone="danger" label={`Delete ${legend.toLowerCase()}`} onclick={() => ondelete?.()}>
					<Icon icon={trashIcon} width="20" height="20" />
				</IconButton>
			</div>
		</div>
	{:else}
	<div class="selector-body" class:has-image={source}>
	<div
		role="group"
		aria-label={`${legend} image upload`}
		class:dragging={isDragOver}
		class="upload-surface"
		ondragover={dragOver}
		ondragleave={dragLeave}
		ondrop={dropImage}
	>
		<div class="upload-mark">
			<Icon icon={photoIcon} width="22" height="22" />
		</div>
		<div class="min-w-0">
			<p class="upload-title">{source ? 'Replace selected image' : `Upload ${legend.toLowerCase()}`}</p>
			<p class="upload-hint">Drop a JPG, PNG, or WebP here, or choose one to crop before saving.</p>
		</div>
		<label class="file-trigger" for={id}>
			<Icon icon={uploadIcon} width="17" height="17" />
			{source ? 'Replace' : 'Choose image'}
		</label>
		<input
			{id}
			type="file"
			accept="image/jpeg,image/png,image/webp"
			onchange={selectImage}
			class="sr-only"
		/>
	</div>
	{#if source}
		<div class="crop-editor">
			<div class="crop-editor-header">
				<div>
					<p class="crop-title">Crop image</p>
					<p class="crop-hint">Drag to pan and resize the frame to crop.</p>
				</div>
				<div class="flex gap-1">
					<Button variant="ghost" class="size-9 !min-h-0 !p-0" onclick={() => zoom(-0.1)}>
						<Icon icon={zoomOutIcon} width="19" height="19" class="shrink-0" />
						<span class="sr-only">Zoom out</span>
					</Button>
					<Button variant="ghost" class="size-9 !min-h-0 !p-0" onclick={() => zoom(0.1)}>
						<Icon icon={zoomInIcon} width="19" height="19" class="shrink-0" />
						<span class="sr-only">Zoom in</span>
					</Button>
					<Button variant="ghost" class="size-9 !min-h-0 !p-0" onclick={() => rotate(-90)}>
						<Icon icon={rotateClockwiseIcon} width="19" height="19" class="-scale-x-100 shrink-0" />
						<span class="sr-only">Rotate left</span>
					</Button>
					<Button variant="ghost" class="size-9 !min-h-0 !p-0" onclick={() => rotate(90)}>
						<Icon icon={rotateClockwiseIcon} width="19" height="19" class="shrink-0" />
						<span class="sr-only">Rotate right</span>
					</Button>
					<Button variant="ghost" class="size-9 !min-h-0 !p-0" onclick={resetCrop}>
						<Icon icon={refreshIcon} width="19" height="19" class="shrink-0" />
						<span class="sr-only">Reset crop</span>
					</Button>
				</div>
			</div>
			<div class="cropper-host" bind:this={container}>
				<img bind:this={image} src={source} alt={`Crop ${filename}`} />
			</div>
		</div>
	{/if}
	</div>
	{/if}
</fieldset>

<style>
	.image-selector {
		border: 2px solid rgb(var(--color-border));
		border-radius: 8px;
		padding: 1rem;
	}

	.selector-legend {
		padding: 0 0.5rem;
		font-size: 0.875rem;
		font-weight: 600;
		color: rgb(var(--color-text));
	}

	.current-avatar {
		display: flex;
		align-items: center;
		gap: 1rem;
		margin-bottom: 1rem;
		font-size: 0.8125rem;
		color: rgb(var(--color-text) / 0.75);
	}

	.current-actions {
		display: flex;
		gap: 0.5rem;
	}

	.current-avatar img {
		width: 6rem;
		height: 6rem;
		border-radius: 50%;
		object-fit: cover;
		box-shadow: 0 0 0 1px rgb(var(--color-border));
	}

	.selector-body {
		display: flex;
		flex-wrap: wrap;
		align-items: flex-start;
		gap: 1rem;
	}

	.upload-surface {
		flex: 1 1 16rem;
		display: grid;
		grid-template-columns: auto minmax(0, 1fr) auto;
		align-items: center;
		gap: 0.875rem;
		min-height: 5.25rem;
		padding: 0.75rem;
		border: 1px dashed rgb(var(--color-border));
		border-radius: 8px;
		background: rgb(var(--color-text) / 0.06);
		transition: background 0.18s, border-color 0.18s;
	}

	.upload-surface.dragging {
		border-color: rgb(var(--color-primary));
		background: rgb(var(--color-primary) / 0.08);
	}

	.upload-mark {
		display: grid;
		width: 2.75rem;
		height: 2.75rem;
		place-items: center;
		border-radius: 8px;
		background: rgb(var(--color-primary) / 0.1);
		color: rgb(var(--color-primary));
	}

	.upload-title,
	.crop-title {
		margin: 0;
		font-weight: 600;
		color: rgb(var(--color-text));
	}

	.upload-hint,
	.crop-hint {
		margin: 0.1875rem 0 0;
		font-size: 0.75rem;
		color: rgb(var(--color-text) / 0.65);
	}

	.file-trigger {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 0.375rem;
		min-height: 2.25rem;
		padding: 0.375rem 0.625rem;
		border: 2px solid rgb(var(--color-primary));
		border-radius: 8px;
		background: rgb(var(--color-foreground));
		color: rgb(var(--color-text));
		font-size: 0.8125rem;
		font-weight: 700;
		cursor: pointer;
		transition: background 0.18s, color 0.18s;
	}

	.file-trigger:hover {
		background: rgb(var(--color-primary) / 0.1);
		color: rgb(var(--color-primary));
	}

	.crop-editor {
		flex: 1 1 20rem;
		min-width: 0;
		max-width: 26rem;
	}

	.crop-editor-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 0.75rem;
	}

	.cropper-host {
		height: 19rem;
		overflow: hidden;
		border: 1px solid rgb(var(--color-border));
		border-radius: 8px;
	}

	.cropper-host :global(cropper-canvas) {
		height: 100%;
	}

	.cropper-host :global(cropper-selection.round-selection) {
		border-radius: 50%;
		overflow: hidden;
	}

	@media (max-width: 480px) {
		.upload-surface {
			grid-template-columns: auto minmax(0, 1fr);
		}

		.file-trigger {
			grid-column: 2;
			justify-self: start;
		}
	}
</style>
