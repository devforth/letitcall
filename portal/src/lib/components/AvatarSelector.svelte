<script lang="ts">
	import { onDestroy, tick } from 'svelte';
	import Cropper from 'cropperjs';
	import Button from '$lib/components/ui/Button.svelte';

	let { id = 'avatar' }: { id?: string } = $props();

	const avatarTemplate = `
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

	async function selectAvatar(event: Event) {
		const file = (event.currentTarget as HTMLInputElement).files?.[0];
		if (!file) return;
		destroyCropper();
		source = URL.createObjectURL(file);
		filename = file.name;
		await tick();
		if (!container || !image) return;
		cropper = new Cropper(image, { container, template: avatarTemplate });
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

	export async function exportAvatar(): Promise<string> {
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

<fieldset class="grid gap-3 border border-black p-4">
	<legend class="px-2 text-sm font-medium">Avatar</legend>
	<label class="grid gap-2 text-sm" for={id}>
		<span class="font-medium">Choose image</span>
		<input
			{id}
			type="file"
			accept="image/jpeg,image/png,image/webp"
			onchange={selectAvatar}
			class="min-h-11 w-full border border-black bg-white px-3 py-2 text-black file:mr-3 file:border file:border-black file:bg-white file:px-3 file:py-1 file:text-black"
		/>
	</label>
	{#if source}
		<p class="text-xs">Drag to pan, use the frame to crop, and scroll or use the buttons to zoom.</p>
		<div class="cropper-host border border-black" bind:this={container}>
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
	.cropper-host {
		height: 20rem;
	}

	.cropper-host :global(cropper-canvas) {
		height: 100%;
	}
</style>
