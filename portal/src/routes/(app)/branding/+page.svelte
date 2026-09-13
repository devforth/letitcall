<script lang="ts">
	import { onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import paintIcon from '@iconify-icons/tabler/paint-filled';
	import { callApi, logoURL } from '$lib/api';
	import { defaultBrandingTheme, loadBranding } from '$lib/stores/branding.svelte';
	import { generateThemeColors } from '$lib/theme-colors';
	import type { Branding, BrandingTheme, ThemeColors } from '$lib/types';
	import { showSuccess } from '$lib/notifications';
	import Button from '$lib/components/ui/Button.svelte';
	import ColorPicker from '$lib/components/ui/ColorPicker.svelte';
	import ImageSelector from '$lib/components/ImageSelector.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import Tooltip from '$lib/components/ui/Tooltip.svelte';

	const colorFields: { key: keyof ThemeColors; label: string; description: string }[] = [
		{
			key: 'primary',
			label: 'Primary color',
			description: 'Color of buttons and active elements. This can be your brand color.'
		},
		{
			key: 'primaryContrast',
			label: 'Primary color contrast',
			description: 'Color of text shown on primary buttons and active elements.'
		},
		{
			key: 'foreground',
			label: 'Foreground',
			description: 'Color of foreground surfaces such as panels, menus and cards.'
		},
		{
			key: 'text',
			label: 'Foreground text color',
			description: 'Color of text on foreground surfaces and the page background.'
		},
		{
			key: 'background',
			label: 'Background',
			description: 'Background color of the page.'
		},
		{
			key: 'border',
			label: 'Border color',
			description: 'Color of borders and dividers around surfaces and controls.'
		},
		{
			key: 'shadow',
			label: 'Shadow color',
			description: 'Color used for shadows below surfaces and controls.'
		}
	];

	let name = $state('');
	let logoPath = $state('');
	let brandingTheme = $state<BrandingTheme>(structuredClone(defaultBrandingTheme));
	let imageSelector = $state<ImageSelector | null>(null);
	let loading = $state(true);
	let saving = $state(false);

	function setForm(loaded: Branding) {
		name = loaded.name;
		logoPath = loaded.logoPath ?? '';
		brandingTheme = structuredClone(loaded.theme);
	}

	onMount(async () => {
		try {
			setForm(await loadBranding());
		} catch {
			// callApi reports the error globally.
		} finally {
			loading = false;
		}
	});

	function generate(mode: 'light' | 'dark') {
		brandingTheme[mode] = generateThemeColors(brandingTheme[mode].primary, mode);
	}

	async function saveBranding(event: SubmitEvent) {
		event.preventDefault();
		saving = true;
		try {
			const logo = (await imageSelector?.exportImage()) ?? '';
			await callApi<{ branding: Branding }>('/api/branding', {
				method: 'PUT',
				body: JSON.stringify({ name, theme: brandingTheme, ...(logo ? { logo } : {}) })
			});
			setForm(await loadBranding());
			showSuccess('Branding applied');
		} catch {
			// callApi reports the error globally.
		} finally {
			saving = false;
		}
	}
</script>

<PageTitle title="Branding" />

<section aria-labelledby="branding-title" class="flex flex-col gap-6">
	<div class="mb-2">
		<div class="flex min-w-0 items-center gap-4">
			<div
				class="grid size-12 shrink-0 place-items-center rounded-lg"
				style="background: rgb(var(--color-primary) / 0.12); color: rgb(var(--color-primary));"
			>
				<Icon icon={paintIcon} width="24" height="24" />
			</div>
			<div>
				<h1 id="branding-title" class="text-2xl font-semibold tracking-tight" style="color: rgb(var(--color-text));">Branding</h1>
				<p class="text-sm" style="color: rgb(var(--color-text) / 0.65);">Set the identity and light and dark color themes shown across the portal and booking pages.</p>
			</div>
		</div>
	</div>

	{#if loading}
		<p class="loading-panel p-6 text-sm">Loading branding…</p>
	{:else}
		<form class="branding-form" onsubmit={saveBranding}>
			<fieldset class="section">
				<legend>Identity</legend>
				<div class="identity-fields">
					<Input id="brand-name" label="Brand name" bind:value={name} required />
					{#if logoPath}
						<div class="current-logo">
							<span>Current logo</span>
							<img src={logoURL(logoPath)} alt={`${name} logo`} />
						</div>
					{/if}
					<ImageSelector id="brand-logo" legend="Logo" bind:this={imageSelector} />
				</div>
			</fieldset>

			<fieldset class="section">
				<legend>Color theme</legend>
				<p class="section-description">Choose a color swatch or enter a six-digit hex value. Generate creates an accessible palette from the primary color.</p>

				<div class="theme-table-wrap">
					<table class="theme-table">
						<thead>
							<tr>
								<th scope="col">Color</th>
								<th scope="col">Light theme</th>
								<th scope="col">Dark theme</th>
							</tr>
						</thead>
						<tbody>
							{#each colorFields as field}
								<tr>
									<th scope="row">
										<span>{field.label}</span>
										<Tooltip text={field.description} />
									</th>
									<td>
										<ColorPicker
											id={`light-${field.key}`}
											label={`Light theme ${field.label}`}
											bind:value={brandingTheme.light[field.key]}
										/>
										{#if field.key === 'primary'}
											<div class="generate"><Button size="small" variant="secondary" onclick={() => generate('light')}>Generate</Button></div>
										{/if}
									</td>
									<td>
										<ColorPicker
											id={`dark-${field.key}`}
											label={`Dark theme ${field.label}`}
											bind:value={brandingTheme.dark[field.key]}
										/>
										{#if field.key === 'primary'}
											<div class="generate"><Button size="small" variant="secondary" onclick={() => generate('dark')}>Generate</Button></div>
										{/if}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</fieldset>

			<div class="branding-actions"><Button type="submit" disabled={saving}>{saving ? 'Applying…' : 'Apply'}</Button></div>
		</form>
	{/if}
</section>

<style>
	.loading-panel {
		border: 0;
		border-radius: 8px;
		background: rgb(var(--color-foreground));
		box-shadow: 0 0 0 1px rgb(var(--color-border));
	}

	.branding-form {
		display: grid;
		gap: 2rem;
	}

	.section {
		min-width: 0;
		padding: 0;
		border: 0;
	}

	.section legend {
		padding: 0;
		font-size: 1rem;
		font-weight: 600;
	}

	.identity-fields {
		display: grid;
		gap: 1rem;
		margin-top: 1rem;
		max-width: 36rem;
	}

	.current-logo {
		display: grid;
		gap: 0.5rem;
		font-size: 0.875rem;
		font-weight: 500;
	}

	.current-logo img {
		width: 6rem;
		height: 6rem;
		object-fit: cover;
	}

	.section-description {
		margin: 0.5rem 0 1rem;
		font-size: 0.875rem;
		color: rgb(var(--color-text) / 0.75);
	}

	.theme-table-wrap {
		overflow: hidden;
		border-radius: 8px;
		box-shadow: 0 0 0 1px rgb(var(--color-border));
	}

	.theme-table {
		width: 100%;
		min-width: 43rem;
		border-collapse: collapse;
		text-align: left;
	}

	.theme-table th,
	.theme-table td {
		padding: 0.875rem;
		border-bottom: 1px solid rgb(var(--color-border));
		vertical-align: top;
	}

	.theme-table thead th {
		background: rgb(var(--color-text) / 0.06);
		font-size: 0.75rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.theme-table tbody th {
		width: 32%;
		font-size: 0.875rem;
	}

	.theme-table tbody th > span:first-child {
		margin-right: 0.375rem;
	}

	.theme-table tbody tr:last-child th,
	.theme-table tbody tr:last-child td {
		border-bottom: 0;
	}

	.generate {
		margin-top: 0.5rem;
	}

	.branding-actions {
		display: flex;
		justify-content: flex-end;
		margin-top: 0.5rem;
	}

	@media (max-width: 800px) {
		.theme-table-wrap {
			overflow-x: auto;
		}
	}

</style>
