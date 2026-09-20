<script lang="ts">
	import { onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import aiSparklesIcon from '@iconify-icons/hugeicons/ai-sparkles';
	import checkIcon from '@iconify-icons/tabler/check';
	import colorSwatchIcon from '@iconify-icons/tabler/color-swatch';
	import refreshIcon from '@iconify-icons/tabler/refresh';
	import { callApi, logoURL } from '$lib/api';
	import { defaultBrandingTheme, loadBranding } from '$lib/stores/branding.svelte';
	import { generateThemeColors } from '$lib/theme-colors';
	import type { Branding, BrandingTheme, ThemeColors } from '$lib/types';
	import { showSuccess } from '$lib/notifications';
	import Button from '$lib/components/ui/Button.svelte';
	import ColorPicker from '$lib/components/ui/ColorPicker.svelte';
	import ImageSelector from '$lib/components/ImageSelector.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';

	const colorFields: { key: keyof ThemeColors; label: string; description: string }[] = [
		{
			key: 'primary',
			label: 'Primary color',
			description: 'Color of buttons and active elements; this can be your brand color. Click "Generate" to create an accessible palette from the primary color'
		},
		{
			key: 'primaryContrast',
			label: 'Primary color contrast',
			description: 'Color of text shown on primary buttons and active elements'
		},
		{
			key: 'text',
			label: 'Text color',
			description: 'Color of text across the theme'
		},
		{
			key: 'background',
			label: 'Background',
			description: 'Color of surfaces, panels, menus, cards, and the page'
		},
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

	function resetPalette(mode: 'light' | 'dark') {
		brandingTheme[mode] = structuredClone(defaultBrandingTheme[mode]);
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

{#snippet paletteActions(mode: 'light' | 'dark')}
	<div class="palette-actions">
		<Button size="small" variant="primary-outline" onclick={() => generate(mode)}>
			<span class="flex items-center gap-2"><Icon icon={aiSparklesIcon} width="18" height="18" />Generate</span>
		</Button>
		<Button size="small" variant="outline" onclick={() => resetPalette(mode)}>
			<span class="flex items-center gap-2"><Icon icon={refreshIcon} width="18" height="18" />Reset</span>
		</Button>
	</div>
{/snippet}

<PageTitle title="Branding" />

<section aria-labelledby="branding-title" class="flex flex-col gap-6">
	<div class="mb-2">
		<PageHeader
			id="branding-title"
			title="Branding"
			description="Set the identity and light and dark color themes shown across the portal and booking pages."
			icon={colorSwatchIcon}
		/>
	</div>

	{#if loading}
		<p class="loading-panel p-6 text-sm">Loading branding…</p>
	{:else}
		<form class="branding-form" onsubmit={saveBranding}>
			<fieldset class="section">
				<legend>Identity</legend>
				<div class="identity-fields">
					<div class="brand-name-field">
						<Input id="brand-name" label="Brand name" bind:value={name} required />
					</div>
					<ImageSelector
						id="brand-logo"
						legend="Logo"
						current={logoPath ? logoURL(logoPath) : ''}
						ondelete={() => (logoPath = '')}
						bind:this={imageSelector}
					/>
				</div>
			</fieldset>

			<fieldset class="section">
				<legend>Color theme</legend>
				<p class="section-description">Choose a color swatch or enter a six-digit hex value; Generate creates an accessible palette from the primary color</p>

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
										<span class="theme-field-label">{field.label}</span>
										<p class="theme-field-description">{field.description}</p>
									</th>
									<td>
										<ColorPicker
											id={`light-${field.key}`}
											label={`Light theme ${field.label}`}
											bind:value={brandingTheme.light[field.key]}
										/>
										{#if field.key === 'primary'}
											{@render paletteActions('light')}
										{/if}
									</td>
									<td>
										<ColorPicker
											id={`dark-${field.key}`}
											label={`Dark theme ${field.label}`}
											bind:value={brandingTheme.dark[field.key]}
										/>
										{#if field.key === 'primary'}
											{@render paletteActions('dark')}
										{/if}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</fieldset>

			<div class="branding-actions">
				<Button type="submit" rounded class="primary-action-button" disabled={saving}>
					<span class="flex items-center gap-2">
						<Icon icon={checkIcon} width="20" height="20" />
						{saving ? 'Applying…' : 'Apply'}
					</span>
				</Button>
			</div>
		</form>
	{/if}
</section>

<style>
	.loading-panel {
		border: 0;
		border-radius: 8px;
		background: rgb(var(--color-background));
		box-shadow: 0 0 0 1px var(--color-border);
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
	}

	.brand-name-field {
		width: 50%;
	}

	.section-description {
		margin: 0.5rem 0 1rem;
		font-size: 0.875rem;
		color: rgb(var(--color-text) / 0.75);
	}

	.theme-table-wrap {
		width: fit-content;
		max-width: 100%;
		overflow: hidden;
		border-radius: 8px;
		box-shadow: 0 0 0 1px var(--color-border);
	}

	.theme-table {
		width: max-content;
		min-width: 47rem;
		border-collapse: collapse;
		text-align: left;
	}

	.theme-table th,
	.theme-table td {
		padding: 0.875rem;
		border-bottom: 1px solid var(--color-border);
		vertical-align: top;
	}

	.theme-table thead th {
		background: rgb(var(--color-text) / 0.06);
		font-size: 0.75rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.theme-table tbody th {
		width: 22rem;
		font-size: 0.875rem;
	}

	.theme-field-label {
		display: block;
	}

	.theme-field-description {
		margin: 0.25rem 0 0;
		color: rgb(var(--color-text) / 0.65);
		font-size: 0.875rem;
		font-weight: 400;
		line-height: 1.35;
	}

	.theme-table tbody tr:last-child th,
	.theme-table tbody tr:last-child td {
		border-bottom: 0;
	}

	.palette-actions {
		display: grid;
		gap: 0.5rem;
		width: 11rem;
		margin-top: 0.5rem;
	}

	.palette-actions :global(button) {
		width: 100%;
		min-height: 2.75rem;
		border-radius: 0.75rem;
		font-size: 0.875rem;
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
