<script lang="ts">
	import { onDestroy, tick } from 'svelte';
	import { fade } from 'svelte/transition';
	import Cropper from 'cropperjs';
	import Icon from '@iconify/svelte';
	import photoIcon from '@iconify-icons/tabler/photo';
	import refreshIcon from '@iconify-icons/tabler/refresh';
	import rotateClockwiseIcon from '@iconify-icons/tabler/rotate-clockwise';
	import uploadIcon from '@iconify-icons/tabler/upload';
	import zoomInIcon from '@iconify-icons/tabler/zoom-in';
	import zoomOutIcon from '@iconify-icons/tabler/zoom-out';
	import editIcon from '@iconify-icons/mdi/edit';
	import trashIcon from '@iconify-icons/tabler/trash';
	import Button from '$lib/components/ui/Button.svelte';
	import IconButton from '$lib/components/ui/IconButton.svelte';
	import { showError } from '$lib/notifications';
	import type { ImageEditor, ImageUpload } from '$lib/types';

	let {
		id,
		legend,
		current = '',
		original = '',
		editor,
		onchange,
		ondelete
	}: {
		id: string;
		legend: string;
		current?: string;
		original?: string;
		editor?: ImageEditor;
		onchange?: () => void;
		ondelete?: () => void;
	} = $props();

	let editing = $state(false);
	let roundCrop = $derived(legend.toLowerCase().startsWith('avatar'));

	let imageTemplate = $derived(`
		<cropper-canvas background scale-step="0.1">
			<cropper-image initial-center-size="cover" rotatable scalable translatable></cropper-image>
			<cropper-shade class="${roundCrop ? 'round-shade' : ''}" theme-color="rgba(0, 0, 0, 0.35)"></cropper-shade>
			<cropper-handle action="move" plain></cropper-handle>
			<cropper-selection class="${roundCrop ? 'round-selection' : ''}" initial-aspect-ratio="1" aspect-ratio="1" initial-coverage="0.8" theme-color="#000" outlined>
				<cropper-crosshair centered theme-color="#000"></cropper-crosshair>
				<cropper-handle action="move" plain></cropper-handle>
			</cropper-selection>
		</cropper-canvas>
	`);

	let container = $state<HTMLDivElement>();
	let image = $state<HTMLImageElement>();
	let cropper: Cropper | null = null;
	let source = $state('');
	let originalData = $state('');
	let filename = $state('');
	let isDragOver = $state(false);

	async function selectImage(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		await loadImage(file);
		input.value = '';
	}

	async function loadImage(file: File) {
		let png: string;
		try {
			png = await convertToPNG(await readImage(file));
		} catch {
			showError('Unsupported image format');
			return;
		}
		originalData = png;
		await openImage(png, file.name);
		onchange?.();
	}

	function convertToPNG(source: string): Promise<string> {
		return new Promise((resolve, reject) => {
			const sourceImage = new Image();
			sourceImage.onload = () => {
				const canvas = document.createElement('canvas');
				canvas.width = sourceImage.naturalWidth;
				canvas.height = sourceImage.naturalHeight;
				canvas.getContext('2d')!.drawImage(sourceImage, 0, 0);
				resolve(canvas.toDataURL('image/png'));
			};
			sourceImage.onerror = reject;
			sourceImage.src = source;
		});
	}

	async function editImage() {
		editing = true;
		if (original) {
			originalData = '';
			await openImage(original, original, editor);
			return;
		}
		const response = await fetch(current);
		originalData = await convertToPNG(await readImage(await response.blob()));
		await openImage(originalData, current);
	}

	function readImage(blob: Blob): Promise<string> {
		return new Promise((resolve) => {
			const reader = new FileReader();
			reader.onload = () => resolve(reader.result as string);
			reader.readAsDataURL(blob);
		});
	}

	async function openImage(imageSource: string, imageName: string, savedEditor?: ImageEditor) {
		destroyCropper();
		source = imageSource;
		filename = imageName;
		await tick();
		if (!container || !image) return;
		cropper = new Cropper(image, { container, template: imageTemplate });
		const cropperImage = cropper.getCropperImage();
		await cropperImage?.$ready();
		if (cropperImage && savedEditor) {
			cropperImage.$setTransform(savedEditor.transform);
			const selection = savedEditor.selection;
			cropper.getCropperSelection()?.$change(selection.x, selection.y, selection.width, selection.height);
		}
		cropperImage?.addEventListener('transform', () => onchange?.());
		cropper.getCropperSelection()?.addEventListener('change', () => onchange?.());
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
	}

	export function showCurrent() {
		destroyCropper();
		source = '';
		originalData = '';
		filename = '';
		editing = false;
	}

	export async function exportImage(): Promise<ImageUpload | undefined> {
		const selection = cropper?.getCropperSelection();
		const cropperImage = cropper?.getCropperImage();
		if (!selection || !cropperImage) return;
		// Backend requires a 512×512 PNG; the round look is applied cosmetically in the UI.
		const canvas = await selection.$toCanvas({
			width: 512,
			height: 512,
			beforeDraw: (context, output) => {
				if (!roundCrop) return;
				context.fillStyle = '#fff';
				context.fillRect(0, 0, output.width, output.height);
			}
		});
		return {
			rendered: canvas.toDataURL('image/png'),
			...(originalData ? { original: originalData } : {}),
			editor: {
				transform: cropperImage.$getTransform() as ImageEditor['transform'],
				selection: {
					x: selection.x,
					y: selection.y,
					width: selection.width,
					height: selection.height
				}
			}
		};
	}

	onDestroy(destroyCropper);
</script>

<fieldset class="image-selector" class:current-state={Boolean(current) && !source && !editing}>
	<legend class="selector-legend">{legend}</legend>
	{#if current && !source && !editing}
		<div class="current-avatar" in:fade={{ duration: 180 }}>
			<img class:round-image={roundCrop} src={current} alt={`Current ${legend.toLowerCase()}`} />
			<div class="current-copy">
				<p class="current-title">Current {legend.toLowerCase()}</p>
				<p class="current-hint">Edit the image or remove it</p>
			</div>
			<div class="current-actions">
				<IconButton filled tone="primary" label={`Edit ${legend.toLowerCase()}`} onclick={editImage}>
					<Icon icon={editIcon} width="20" height="20" />
				</IconButton>
				<IconButton filled tone="danger" label={`Delete ${legend.toLowerCase()}`} onclick={() => ondelete?.()}>
					<Icon icon={trashIcon} width="20" height="20" />
				</IconButton>
			</div>
		</div>
	{:else}
	<div class="selector-body" class:has-image={source} in:fade={{ duration: 180 }}>
	<div
		role="group"
		aria-label={`${legend} image upload`}
		class:dragging={isDragOver}
		class="upload-surface"
		class:empty={!source}
		ondragover={dragOver}
		ondragleave={dragLeave}
		ondrop={dropImage}
	>
		<div class="upload-mark">
			<Icon icon={photoIcon} width="22" height="22" />
		</div>
		<div class="upload-copy min-w-0">
			<p class="upload-title">{source ? 'Replace selected image' : `Upload ${legend.toLowerCase()}`}</p>
			<p class="upload-hint">Drop an image here, or choose one to crop before saving</p>
		</div>
		<label class="file-trigger button-primary-outline" for={id}>
			<Icon icon={uploadIcon} width="17" height="17" />
			{source ? 'Replace' : 'Choose image'}
		</label>
		<input
			{id}
			type="file"
			accept="image/*"
			onchange={selectImage}
			class="sr-only"
		/>
	</div>
	{#if source}
		<div class="crop-editor">
			<div class="crop-editor-header">
				<div>
					<p class="crop-title">Crop image</p>
					<p class="crop-hint">Drag to pan and resize the frame to crop</p>
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
		box-sizing: border-box;
		width: 100%;
		border: 0;
		border-radius: 8px;
		padding: 1rem;
		box-shadow: 0 0 0 1px var(--color-border);
		transition: width 0.2s ease;
	}

	.image-selector.current-state {
		width: 50%;
	}

	.selector-legend {
		padding: 0 0.25rem;
		background: rgb(var(--color-background));
		font-size: 0.72rem;
		font-weight: 400;
		color: rgb(var(--color-text));
	}

	.current-avatar {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr) auto;
		align-items: center;
		gap: 0.875rem;
		min-height: 0;
		padding: 0;
		transform: translateY(-0.25rem);
		font-size: 0.8125rem;
		color: rgb(var(--color-text) / 0.75);
	}

	.current-title {
		margin: 0;
		font-weight: 600;
		color: rgb(var(--color-text));
	}

	.current-hint {
		margin: 0.1875rem 0 0;
		font-size: 0.75rem;
		color: rgb(var(--color-text) / 0.65);
	}

	.current-actions {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.current-avatar img {
		width: 4.5rem;
		height: 4.5rem;
		object-fit: cover;
	}

	.current-avatar img.round-image {
		border-radius: 50%;
	}

	.selector-body {
		display: flex;
		flex-wrap: wrap;
		align-items: flex-start;
		gap: 1rem;
	}

	.selector-body.has-image {
		align-items: stretch;
	}

	.selector-body.has-image .upload-surface {
		grid-template-columns: 1fr;
		align-content: center;
		justify-items: center;
		text-align: center;
	}

	.selector-body.has-image .upload-copy {
		max-width: 15rem;
	}

	.selector-body.has-image .file-trigger {
		grid-column: auto;
		justify-self: center;
	}

	.upload-surface {
		flex: 1 1 16rem;
		display: grid;
		grid-template-columns: auto minmax(0, 1fr) auto;
		align-items: center;
		gap: 0.875rem;
		min-height: 5.25rem;
		padding: 0.75rem;
		border: 1px dashed var(--color-border);
		border-radius: 8px;
		background: rgb(var(--color-text) / 0.06);
		transition: background 0.18s, border-color 0.18s;
	}

	.upload-surface.empty {
		align-content: center;
		min-height: 0;
		padding: 0;
		border: 0;
		background: transparent;
		transform: translateY(-0.25rem);
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
		background: rgb(var(--color-text) / 0.06);
		color: rgb(var(--color-text));
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
		padding: 0.375rem 0.75rem;
		padding-right: 1rem;
		padding-bottom: 0.5rem;
		border-radius: 9999px;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
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
		border: 1px solid var(--color-border);
		border-radius: 8px;
	}

	.cropper-host :global(cropper-canvas) {
		height: 100%;
	}

	.cropper-host :global(cropper-selection.round-selection) {
		border-radius: 50%;
		overflow: hidden;
	}

	.cropper-host :global(cropper-shade.round-shade) {
		border-radius: 50%;
	}

	@media (max-width: 640px) {
		.image-selector.current-state {
			width: 100%;
		}
	}

	@media (max-width: 480px) {
		.current-avatar {
			grid-template-columns: auto minmax(0, 1fr);
		}

		.current-actions {
			grid-column: 2;
		}

		.upload-surface {
			grid-template-columns: auto minmax(0, 1fr);
		}

		.file-trigger {
			grid-column: 2;
			justify-self: start;
		}
	}
</style>
