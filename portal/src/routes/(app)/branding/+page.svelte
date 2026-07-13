<script lang="ts">
	import { onMount } from 'svelte';
	import { callApi, logoURL } from '$lib/api';
	import { branding as publicBranding } from '$lib/stores/branding.svelte';
	import type { Branding } from '$lib/types';
	import Button from '$lib/components/ui/Button.svelte';
	import ImageSelector from '$lib/components/ImageSelector.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';

	let name = $state('');
	let logoPath = $state('');
	let imageSelector = $state<ImageSelector | null>(null);
	let loading = $state(true);
	let saving = $state(false);

	onMount(async () => {
		try {
			const response = await callApi<{ branding: Branding }>('/api/branding');
			name = response.branding.name;
			logoPath = response.branding.logoPath ?? '';
		} catch {
			// callApi reports the error globally.
		} finally {
			loading = false;
		}
	});

	async function saveBranding(event: SubmitEvent) {
		event.preventDefault();
		saving = true;
		try {
			const logo = (await imageSelector?.exportImage()) ?? '';
			const response = await callApi<{ branding: Branding }>('/api/branding', {
				method: 'PUT',
				body: JSON.stringify({ name, ...(logo ? { logo } : {}) })
			});
			name = response.branding.name;
			logoPath = response.branding.logoPath ?? '';
			publicBranding.name = name;
			publicBranding.logoPath = logoPath;
		} catch {
			// callApi reports the error globally.
		} finally {
			saving = false;
		}
	}
</script>

<PageTitle title="Branding" />

<section aria-labelledby="branding-title">
	<div class="mb-6">
		<h1 id="branding-title" class="text-2xl font-semibold tracking-tight">Branding</h1>
		<p class="mt-2 text-sm">Set the name and logo shown across the portal and booking pages.</p>
	</div>

	{#if loading}
		<p class="border border-black p-6 text-sm">Loading branding…</p>
	{:else}
		<form class="grid max-w-3xl gap-5 border border-black p-5" onsubmit={saveBranding}>
			<Input id="brand-name" label="Brand name" bind:value={name} required />
			{#if logoPath}
				<div class="grid gap-2 text-sm">
					<span class="font-medium">Current logo</span>
					<img src={logoURL(logoPath)} alt={`${name} logo`} class="size-24 border border-black object-cover" />
				</div>
			{/if}
			<ImageSelector id="brand-logo" legend="Logo" bind:this={imageSelector} />
			<div><Button type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save branding'}</Button></div>
		</form>
	{/if}
</section>
