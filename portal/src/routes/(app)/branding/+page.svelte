<script lang="ts">
	import { onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import alertTriangleIcon from '@iconify-icons/tabler/alert-triangle';
	import checkIcon from '@iconify-icons/tabler/check';
	import colorSwatchIcon from '@iconify-icons/tabler/color-swatch';
	import moonIcon from '@iconify-icons/tabler/moon';
	import refreshIcon from '@iconify-icons/tabler/refresh';
	import sparklesIcon from '@iconify-icons/tabler/sparkles';
	import archiveRestoreIcon from '@iconify-icons/lucide/archive-restore';
	import sunIcon from '@iconify-icons/tabler/sun';
	import { callApi, logoURL } from '$lib/api';
	import { contrastRatio, wcagAAContrast } from '$lib/color-contrast';
	import { defaultBrandingTheme, loadBranding } from '$lib/stores/branding.svelte';
	import { generateThemeColors } from '$lib/theme-colors';
	import type { Branding, BrandingTheme, ImageSource } from '$lib/types';
	import { showSuccess } from '$lib/notifications';
	import BookingPagePreview from '$lib/components/BookingPagePreview.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import ColorPicker from '$lib/components/ui/ColorPicker.svelte';
	import IconButton from '$lib/components/ui/IconButton.svelte';
	import ImageSelector from '$lib/components/ImageSelector.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import SegmentedControl from '$lib/components/ui/SegmentedControl.svelte';

	const themeModes = ['light', 'dark'] as const;
	const themePresets: Record<string, BrandingTheme> = {
		ocean: {
			light: { primary: '#1075A8', text: '#242F34', background: '#FFFFFF' },
			dark: { primary: '#0284C7', text: '#F5F6F7', background: '#000000' }
		},
		forest: {
			light: { primary: '#28643A', text: '#1C2A20', background: '#FFFFFF' },
			dark: { primary: '#55C274', text: '#F4FBF5', background: '#000000' }
		},
		violet: {
			light: { primary: '#65558F', text: '#25232B', background: '#FFFFFF' },
			dark: { primary: '#D0BCFF', text: '#F7F2FA', background: '#000000' }
		},
		amber: {
			light: { primary: '#985A00', text: '#302814', background: '#FFFCF5' },
			dark: { primary: '#FFB95C', text: '#FFF8F0', background: '#000000' }
		},
		teal: {
			light: { primary: '#0F6B65', text: '#1D2928', background: '#FFFFFF' },
			dark: { primary: '#4FD1C5', text: '#F0FDFA', background: '#000000' }
		},
		indigo: {
			light: { primary: '#3F4FA8', text: '#25273A', background: '#FFFFFF' },
			dark: { primary: '#A5B4FC', text: '#F5F7FF', background: '#000000' }
		},
		rose: {
			light: { primary: '#B4235A', text: '#33242A', background: '#FFFFFF' },
			dark: { primary: '#FF9CBD', text: '#FFF1F5', background: '#000000' }
		},
		coral: {
			light: { primary: '#C2413C', text: '#3A1D1B', background: '#FFFFFF' },
			dark: { primary: '#FF8A80', text: '#FFF5F4', background: '#1F0806' }
		},
		lime: {
			light: { primary: '#4D6B00', text: '#253000', background: '#FFFFFF' },
			dark: { primary: '#B7D957', text: '#F7FBEF', background: '#0D1204' }
		},
		graphite: {
			light: { primary: '#374151', text: '#1F2937', background: '#FFFFFF' },
			dark: { primary: '#D1D5DB', text: '#F9FAFB', background: '#111827' }
		}
	};
	const themePresetTiles = [
		{ value: 'ocean', label: 'Ocean' },
		{ value: 'forest', label: 'Forest' },
		{ value: 'violet', label: 'Violet' },
		{ value: 'amber', label: 'Amber' },
		{ value: 'teal', label: 'Teal' },
		{ value: 'indigo', label: 'Indigo' },
		{ value: 'rose', label: 'Rose' },
		{ value: 'coral', label: 'Coral' },
		{ value: 'lime', label: 'Lime' },
		{ value: 'graphite', label: 'Graphite' }
	];
	const themeSourceOptions = [
		{ value: 'custom', label: 'Custom' },
		{ value: 'preset', label: 'Preset' }
	];

	let name = $state('');
	let logoPath = $state('');
	let logoSource = $state<ImageSource>();
	let brandingTheme = $state<BrandingTheme>(structuredClone(defaultBrandingTheme));
	let selectedThemePreset = $state('custom');
	let themeSource = $state('custom');
	let imageSelector = $state<ImageSelector | null>(null);
	let loading = $state(true);
	let saving = $state(false);
	let savedBranding: Branding;
	let savedForm = $state('');
	let logoChanged = $state(false);
	let formState = $derived(JSON.stringify({ name, logoPath, theme: brandingTheme, preset: selectedThemePreset }));
	let hasUnsavedChanges = $derived(logoChanged || formState !== savedForm);
	let contrastFailures = $derived.by(() => {
		return (['light', 'dark'] as const).flatMap((mode) => {
			const colors = brandingTheme[mode];
			return [
				{ label: 'text and background', ratio: contrastRatio(colors.text, colors.background) },
				{ label: 'brand and background', ratio: contrastRatio(colors.primary, colors.background) }
			]
				.filter(({ ratio }) => ratio < wcagAAContrast)
				.map(({ label, ratio }) => `${mode === 'light' ? 'Light' : 'Dark'} theme ${label}: ${ratio.toFixed(2)}:1`);
		});
	});

	function setForm(loaded: Branding) {
		savedBranding = structuredClone(loaded);
		name = loaded.name;
		logoPath = loaded.logoPath ?? '';
		logoSource = loaded.logoSource;
		brandingTheme = structuredClone(loaded.theme);
		selectedThemePreset = loaded.preset && themePresets[loaded.preset] ? loaded.preset : 'custom';
		themeSource = 'custom';
		logoChanged = false;
		savedForm = JSON.stringify({ name, logoPath, theme: brandingTheme, preset: selectedThemePreset });
	}

	function revertChanges() {
		setForm(savedBranding);
		imageSelector?.showCurrent();
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
		useCustomPalette();
	}

	function resetPalette(mode: 'light' | 'dark') {
		brandingTheme[mode] = structuredClone(defaultBrandingTheme[mode]);
		useCustomPalette();
	}

	function chooseThemePreset(preset: string) {
		brandingTheme = structuredClone(themePresets[preset]);
		selectedThemePreset = preset;
		themeSource = 'preset';
	}

	function useCustomPalette() {
		selectedThemePreset = 'custom';
		themeSource = 'custom';
	}

	function showThemeSource(source: string) {
		themeSource = source;
		if (source === 'custom') useCustomPalette();
	}

	async function saveBranding(event: SubmitEvent) {
		event.preventDefault();
		saving = true;
		try {
			const logo = await imageSelector?.exportImage();
		await callApi<{ branding: Branding }>('/api/branding', {
				method: 'PUT',
				body: JSON.stringify({ name, theme: brandingTheme, preset: selectedThemePreset, ...(logo ? { logo } : {}) })
			});
			setForm(await loadBranding());
			imageSelector?.showCurrent();
			showSuccess('Branding applied');
		} catch {
			// callApi reports the error globally.
		} finally {
			saving = false;
		}
	}
</script>

{#snippet presetTile(value: string, label: string, theme: BrandingTheme)}
	<button
		type="button"
		class="preset-tile"
		class:selected={selectedThemePreset === value}
		aria-pressed={selectedThemePreset === value}
		onclick={() => chooseThemePreset(value)}
	>
		<span class="preset-swatches" aria-hidden="true">
			{#each themeModes as mode}
				<span class="preset-swatch" style={`background: ${theme[mode].primary};`}></span>
				<span class="preset-swatch" style={`background: ${theme[mode].text};`}></span>
				<span class="preset-swatch" style={`background: ${theme[mode].background};`}></span>
			{/each}
		</span>
		<span class="preset-tile-label">
			{#if selectedThemePreset === value}<Icon icon={checkIcon} width="16" height="16" />{/if}
			{label}
		</span>
	</button>
{/snippet}

{#snippet themeCard(mode: 'light' | 'dark')}
	<article class="theme-card">
		<header class="theme-card-header">
			<div class="theme-card-title">
				<Icon icon={mode === 'light' ? sunIcon : moonIcon} width="20" height="20" />
				<h3>{mode === 'light' ? 'Light theme' : 'Dark theme'}</h3>
			</div>
			<div class="theme-card-actions">
				<IconButton
					filled
					tone="primary"
					label="Match text and background to the brand color"
					onclick={() => generate(mode)}
				>
					<Icon icon={sparklesIcon} width="20" height="20" />
				</IconButton>
				<IconButton filled tone="primary" label={`Reset ${mode} theme to default`} onclick={() => resetPalette(mode)}>
					<Icon icon={archiveRestoreIcon} width="20" height="20" />
				</IconButton>
			</div>
		</header>

		<div class="theme-card-body">
			<ColorPicker
				id={`${mode}-primary`}
					label="Brand color"
				description="Color of clicable elements (hex)"
				bind:value={brandingTheme[mode].primary}
				onchange={useCustomPalette}
			/>

			<div class="supporting-colors">
				<ColorPicker
					id={`${mode}-text`}
					label="Text color"
					description="Text on background (hex)"
					bind:value={brandingTheme[mode].text}
					onchange={useCustomPalette}
				/>
				<ColorPicker
					id={`${mode}-background`}
					label="Background"
					description="Surfaces color (hex)"
					bind:value={brandingTheme[mode].background}
					onchange={useCustomPalette}
				/>
			</div>
		</div>
	</article>
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
		<form class="branding-form" class:has-unsaved-changes={hasUnsavedChanges} onsubmit={saveBranding}>
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
						original={logoSource ? logoURL(logoSource.path) : ''}
						editor={logoSource?.editor}
						showCurrentCopy={false}
						onchange={() => (logoChanged = true)}
						ondelete={() => (logoPath = '')}
						bind:this={imageSelector}
					/>
				</div>
			</fieldset>

			<fieldset class="section">
				<legend>Color theme</legend>
				<div class="theme-source-control">
					<SegmentedControl
						options={themeSourceOptions}
						value={themeSource}
						label="Color theme source"
						onchange={showThemeSource}
					/>
				</div>

				{#if themeSource === 'preset'}
					<div class="preset-grid" aria-label="Theme presets">
						{#each themePresetTiles as preset (preset.value)}
							{@render presetTile(preset.value, preset.label, themePresets[preset.value])}
						{/each}
					</div>
				{:else}
					<div class="theme-cards">
						{#each themeModes as mode}
							{@render themeCard(mode)}
						{/each}
					</div>
				{/if}

				{#if contrastFailures.length}
					<div class="contrast-warning" role="alert">
						<div>
							<p class="contrast-warning-title">
								<Icon icon={alertTriangleIcon} width="18" height="18" />
								Palette does not meet WCAG AA
							</p>
							<p class="contrast-warning-details">
								{contrastFailures.join('; ')}. Required contrast is {wcagAAContrast}:1. Use Generate or adjust the colors before applying
							</p>
						</div>
					</div>
				{/if}

				<div class="booking-preview-wrap">
					<BookingPagePreview theme={brandingTheme} />
				</div>
			</fieldset>

			<div class="branding-submit">
				{#if hasUnsavedChanges}
					<div class="unsaved-panel-position">
						<div class="unsaved-panel-boundary">
							<div class="unsaved-panel" aria-live="polite">
								<p class="unsaved-title">You have unsaved changes</p>
								<div class="unsaved-actions">
									<Button variant="primary-outline" class="outlined-action-button" onclick={revertChanges} disabled={saving}>
										<span class="flex items-center gap-2">
											<Icon icon={refreshIcon} width="20" height="20" />
											Revert
										</span>
									</Button>
									<Button type="submit" rounded class="primary-action-button" disabled={saving}>
										<span class="flex items-center gap-2">
											<Icon icon={checkIcon} width="20" height="20" />
											{saving ? 'Applying…' : 'Apply'}
										</span>
									</Button>
								</div>
							</div>
						</div>
					</div>
				{/if}
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

	.branding-form.has-unsaved-changes {
		padding-bottom: 5rem;
	}

	.section {
		min-width: 0;
		padding: 0;
		border: 0;
	}

	.section legend {
		padding: 0;
		font-size: 1.25rem;
		font-weight: 600;
	}

	.identity-fields {
		display: grid;
		grid-template-columns: minmax(0, 1fr);
		gap: 1rem;
		margin-top: 1rem;
	}

	.brand-name-field,
	.identity-fields > :global(fieldset.image-selector) {
		min-width: 0;
		width: 36rem;
		max-width: 100%;
	}

	.brand-name-field {
		width: 20rem;
	}

	.identity-fields > :global(fieldset.image-selector.current-state) {
		width: fit-content;
	}

	.identity-fields > :global(fieldset.image-selector.editing-state) {
		width: 100%;
	}

	.theme-source-control {
		display: flex;
		margin-top: 1rem;
		margin-bottom: 1.25rem;
	}

	.preset-grid {
		display: grid;
		grid-template-columns: repeat(5, minmax(0, 1fr));
		gap: 1rem;
	}

	.booking-preview-wrap {
		margin-top: 1.625rem;
	}

	.preset-tile {
		min-width: 0;
		min-height: 5.5rem;
		border: 1px solid var(--color-border);
		border-radius: 0.875rem;
		background: rgb(var(--color-background));
	}

	.preset-tile {
		display: grid;
		align-content: start;
		gap: 0.625rem;
		padding: 0.625rem;
		color: rgb(var(--color-text));
		text-align: left;
		cursor: pointer;
		transition: box-shadow 0.18s;
		overflow: hidden;
	}

	.preset-tile:hover,
	.preset-tile:focus-visible {
		outline: none;
		box-shadow: 0 0 0 2px rgb(var(--color-primary) / 0.35);
	}

	.preset-tile.selected {
		box-shadow: 0 0 0 2px rgb(var(--color-primary));
	}

	.preset-swatches {
		display: grid;
		grid-template-columns: 2fr 1fr 1fr;
		gap: 0.375rem;
	}

	.preset-swatch {
		height: 1.5rem;
		border-radius: 0.375rem;
		box-shadow: inset 0 0 0 1px var(--color-border);
	}

	.preset-tile-label {
		display: flex;
		align-items: center;
		gap: 0.375rem;
		font-size: 0.875rem;
		font-weight: 600;
	}

	.preset-tile-label :global(svg) {
		flex: none;
		color: rgb(var(--color-primary));
	}

	.theme-cards {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 1.625rem;
	}

	.theme-card {
		min-width: 0;
		border: 1px solid var(--color-border);
		border-radius: 0.75rem;
		background: rgb(var(--color-background));
		overflow: hidden;
	}

	.theme-card-header,
	.theme-card-title {
		display: flex;
		align-items: center;
	}

	.theme-card-header {
		min-height: 3.5rem;
		justify-content: space-between;
		gap: 1rem;
		padding: 0.5rem 1.125rem;
		border-bottom: 1px solid var(--color-border);
	}

	.theme-card-title {
		min-width: 0;
		gap: 0.75rem;
	}

	.theme-card-actions {
		display: flex;
		gap: 0.5rem;
	}

	.theme-card-title :global(svg) {
		flex: none;
		color: rgb(var(--color-text));
	}

	.theme-card-title h3 {
		margin: 0;
		font-size: 1rem;
		font-weight: 600;
	}

	.theme-card-body {
		display: grid;
		gap: 0.625rem;
		padding: 1.375rem 1.125rem;
	}

	.supporting-colors {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(13rem, 100%), 1fr));
		gap: 1.25rem;
		margin-top: 0.75rem;
		margin-inline: -1.125rem;
		padding: 1.375rem 1.125rem 0;
		border-top: 1px solid var(--color-border);
	}

	.branding-submit {
		display: grid;
		gap: 0.75rem;
	}

	.contrast-warning {
		width: 100%;
		min-height: 44px;
		margin-top: 1.625rem;
		border-left: 4px solid rgb(var(--error));
		padding: 0.5rem 0.875rem;
		background: rgb(var(--error) / 0.08);
		color: rgb(var(--error));
		font-size: 0.8125rem;
		font-weight: 600;
		line-height: 1.25;
		animation: contrast-warning-pulse 480ms ease-in-out;
	}

	.contrast-warning-title,
	.contrast-warning-details {
		margin: 0;
	}

	.contrast-warning-title {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font: inherit;
	}

	.contrast-warning-title > :global(svg) {
		flex: none;
	}

	.contrast-warning-details {
		margin-top: 0.25rem;
		font: inherit;
	}

	@keyframes contrast-warning-pulse {
		0%, 100% {
			background: rgb(var(--error) / 0.08);
		}

		45% {
			background: rgb(var(--error) / 0.14);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.contrast-warning {
			animation: none;
		}
	}

	.unsaved-panel-position {
		position: fixed;
		z-index: 30;
		bottom: 0;
		right: 0;
		left: var(--sidebar-w, 0);
		padding: 0 2rem;
		pointer-events: none;
		transition: left 0.3s ease-out;
	}

	.unsaved-panel-boundary {
		display: flex;
		justify-content: flex-end;
		width: 100%;
		max-width: 72rem;
		margin: 0 auto;
	}

	.unsaved-panel {
		position: relative;
		isolation: isolate;
		display: flex;
		align-items: center;
		gap: 1rem;
		width: max-content;
		max-width: 100%;
		padding: 0.875rem 3rem;
		transform: translateX(2rem);
		pointer-events: auto;
	}

	.unsaved-panel::before,
	.unsaved-panel::after {
		position: absolute;
		z-index: -2;
		inset: 0;
		clip-path: polygon(6% 0, 94% 0, 100% 100%, 0 100%);
		clip-path: shape(
			from 5% 12px,
			curve to calc(6% + 12px) 0 with 6% 0,
			line to calc(94% - 12px) 0,
			curve to 95% 12px with 94% 0,
			line to 100% 100%,
			line to 0 100%,
			close
		);
		content: '';
	}

	.unsaved-panel::before {
		background: rgb(var(--color-primary));
		filter: drop-shadow(0 8px 24px rgb(var(--color-text) / 0.16));
	}

	.unsaved-panel::after {
		z-index: -1;
		inset: 1px 1px 0;
		background: rgb(var(--color-background));
	}

	.unsaved-title {
		margin: 0;
		line-height: 1.2;
		font-size: 1.125rem;
		font-weight: 400;
	}

	.unsaved-actions {
		display: flex;
		flex: none;
		gap: 0.5rem;
	}

	@media (max-width: 900px) {
		.unsaved-panel {
			align-items: stretch;
			flex-direction: column;
			gap: 0.75rem;
		}

		.unsaved-actions {
			justify-content: flex-end;
		}

		.theme-cards {
			grid-template-columns: 1fr;
		}

		.preset-grid {
			grid-template-columns: repeat(3, minmax(0, 1fr));
		}
	}

	@media (max-width: 1023px) {
		.unsaved-panel-position {
			padding-right: 1.5rem;
			padding-left: 1.5rem;
		}
	}

	@media (max-width: 767px) {
		.unsaved-panel-position {
			left: 0;
		}
	}

	@media (max-width: 520px) {
		.unsaved-panel-position {
			bottom: 0;
			left: 0;
			padding: 0 1rem;
		}

		.unsaved-panel {
			gap: 1rem;
			width: 100%;
			padding: 0.875rem 1rem;
			transform: none;
		}

		.unsaved-panel::before,
		.unsaved-panel::after {
			clip-path: none;
			border-radius: 12px 12px 0 0;
		}

		.preset-grid {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}

	}

</style>
