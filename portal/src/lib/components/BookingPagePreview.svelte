<script lang="ts">
	import Icon from '@iconify/svelte';
	import arrowRightIcon from '@iconify-icons/tabler/arrow-right';
	import calendarTimeIcon from '@iconify-icons/tabler/calendar-time';
	import clockIcon from '@iconify-icons/tabler/clock';
	import moonIcon from '@iconify-icons/tabler/moon';
	import sunIcon from '@iconify-icons/tabler/sun';
	import worldIcon from '@iconify-icons/tabler/world';
	import type { BrandingTheme } from '$lib/types';
	import Button from '$lib/components/ui/Button.svelte';
	import MonthCalendar from '$lib/components/ui/MonthCalendar.svelte';
	import SearchableSelect from '$lib/components/ui/SearchableSelect.svelte';
	import SegmentedControl from '$lib/components/ui/SegmentedControl.svelte';

	let { brandName, theme }: { brandName: string; theme: BrandingTheme } = $props();

	const previewModes = [
		{ value: 'light', label: 'Light', icon: sunIcon },
		{ value: 'dark', label: 'Dark', icon: moonIcon }
	];
	const availableDates = [
		'2026-10-01', '2026-10-02', '2026-10-05', '2026-10-07', '2026-10-08',
		'2026-10-09', '2026-10-12', '2026-10-14', '2026-10-15', '2026-10-16',
		'2026-10-19', '2026-10-21', '2026-10-22', '2026-10-23', '2026-10-26',
		'2026-10-28', '2026-10-29', '2026-10-30'
	];
	const times = ['9:00 AM', '10:30 AM', '1:00 PM', '2:30 PM'];

	let mode = $state<'light' | 'dark'>('light');
	let month = $state('2026-10');
	let selectedDate = $state('2026-10-15');
	let selectedTime = $state('10:30 AM');
	let previewVisible = $state(false);
	const colors = $derived(theme[mode]);
	const displayName = $derived(brandName.trim() || 'Let It Call');
	const selectedDateLabel = $derived(
		new Intl.DateTimeFormat(undefined, {
			weekday: 'long',
			month: 'long',
			day: 'numeric',
			timeZone: 'UTC'
		}).format(new Date(selectedDate + 'T00:00:00Z'))
	);

	function channels(hex: string): string {
		return [
			Number.parseInt(hex.slice(1, 3), 16),
			Number.parseInt(hex.slice(3, 5), 16),
			Number.parseInt(hex.slice(5, 7), 16)
		].join(' ');
	}

	const previewStyle = $derived(
		'--color-primary: ' + channels(colors.primary) + '; ' +
		'--color-text: ' + channels(colors.text) + '; ' +
		'--color-background: ' + channels(colors.background) + '; ' +
		'--color-border: color-mix(in srgb, rgb(var(--color-text)) 26%, rgb(var(--color-background))); ' +
		'--shadow-small: 0 8px 24px rgb(0 0 0 / 0.12); color-scheme: ' + mode + ';'
	);

	function selectMode(value: string) {
		mode = value as 'light' | 'dark';
	}
</script>

<section aria-labelledby="booking-preview-title">
	<div class="preview-shell overflow-hidden rounded-xl">
		<div class="preview-toolbar flex min-h-15 items-center justify-between gap-4 px-4 py-2.5" class:preview-hidden={!previewVisible}>
			<div class="min-w-0">
				<h3 id="booking-preview-title" class="m-0 text-base font-semibold">Booking page preview</h3>
				<p class="m-0 mt-1 text-sm opacity-60">Mocked event data · changes update before you apply</p>
			</div>
			{#if previewVisible}
				<SegmentedControl
				options={previewModes}
				value={mode}
				label="Preview color theme"
				onchange={selectMode}
				/>
			{/if}
			<Button
				variant="primary-outline"
				class="preview-visibility-button outlined-action-button"
				style="padding-left: 1.25rem !important; padding-right: 1.25rem !important;"
				onclick={() => (previewVisible = !previewVisible)}
			>
				{previewVisible ? 'Hide preview' : 'Show preview'}
			</Button>
		</div>

		<div id="booking-preview-content" class="preview-page grid" style={previewStyle} hidden={!previewVisible} inert>
			<aside class="preview-aside flex flex-col justify-between gap-8 p-8">
				<div>
					<h4 class="m-0 text-3xl font-semibold tracking-tight">Discovery Call</h4>
					<div class="mt-8 flex items-end">
						<span class="avatar grid size-10 shrink-0 place-items-center rounded-full text-xs font-bold" aria-hidden="true">AM</span>
					</div>
					<p class="mt-3 text-sm font-medium">Alex Morgan</p>
					<p class="mt-7 flex items-center gap-2 text-sm font-medium">
						<Icon icon={clockIcon} width="22" height="22" />30 min
					</p>
				</div>
				<div>
					<p class="m-0 text-xl font-semibold">{displayName}</p>
					{#if displayName !== 'Let It Call'}
						<span class="text-xs opacity-70">Powered by <strong>Let It Call</strong></span>
					{/if}
				</div>
			</aside>

			<div class="preview-main flex min-w-0 flex-col p-8">
				<div class="preview-step-summary">
					<span class="preview-step-icon preview-step-icon-desktop" aria-hidden="true">
						<Icon icon={calendarTimeIcon} width="28" height="28" />
					</span>
					<div class="preview-step-heading">
						<h4>
							<span class="preview-step-icon preview-step-icon-mobile" aria-hidden="true">
								<Icon icon={calendarTimeIcon} width="22" height="22" />
							</span>
							Date and Time
						</h4>
						<p>
							<span class="preview-step-count-mobile">Step 1 of 3 · </span>
							Choose a date and time that works best for you
						</p>
					</div>
					<span class="preview-step-count">Step 1 of 3</span>
				</div>
				<div class="progress-bars grid grid-cols-3 gap-1.5" aria-hidden="true">
					<i class="current"></i><i></i><i></i>
				</div>

				<div class="schedule mt-8 grid gap-8">
					<MonthCalendar
						bind:month
						bind:selected={selectedDate}
						{availableDates}
						minimumMonth="2026-10"
						today="2026-10-15"
					/>

					<section aria-label="Mock available times">
						<p class="m-0 text-xs opacity-60">Available times for</p>
						<h5 class="m-0 mt-0.5 text-sm font-semibold">{selectedDateLabel}</h5>
						<div class="time-grid mt-3 grid grid-cols-2 gap-2">
							{#each times as time}
								<button
									type="button"
									class:selected={selectedTime === time}
									aria-pressed={selectedTime === time}
									onclick={() => (selectedTime = time)}
								>{time}</button>
							{/each}
						</div>
						<div class="mt-4">
							<SearchableSelect
								id="preview-timezone"
								label="Timezone"
								icon={worldIcon}
								options={['Europe/Kyiv']}
								value="Europe/Kyiv"
								clearable={false}
							/>
						</div>
					</section>
				</div>

				<div class="preview-actions mt-auto flex items-center justify-end gap-4 pt-6">
					<Button
						class="booking-next preview-next gap-2"
						style="min-height: 2rem !important; padding: 0.25rem 0.75rem !important; font-size: 0.75rem !important;"
					>
						Next
						<Icon icon={arrowRightIcon} width="14" height="14" />
					</Button>
				</div>
			</div>
		</div>
	</div>
</section>

<style>
	.preview-shell {
		container-type: inline-size;
		border: 1px solid var(--color-border);
		background: rgb(var(--color-background));
		color: rgb(var(--color-text));
	}

	.preview-toolbar {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
		border-bottom: 1px solid var(--color-border);
		background: color-mix(in srgb, rgb(var(--color-text)) 4%, rgb(var(--color-background)));
	}

	.preview-toolbar.preview-hidden {
		border-bottom: 0;
	}

	.preview-toolbar :global(.segmented-control) {
		grid-column: 2;
	}

	.preview-toolbar :global(.preview-visibility-button) {
		grid-column: 3;
		justify-self: end;
	}

	.preview-page {
		grid-template-columns: minmax(13rem, 0.34fr) minmax(0, 1fr);
		min-height: 34rem;
		background: rgb(var(--color-background));
		color: rgb(var(--color-text));
		pointer-events: none;
		user-select: none;
	}

	.preview-aside {
		background: rgb(var(--color-primary));
		color: rgb(var(--color-background));
	}

	.avatar {
		border: 2px solid rgb(var(--color-background) / 0.75);
		background: rgb(var(--color-background));
		color: rgb(var(--color-primary));
	}

	.preview-step-summary {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.preview-step-heading {
		min-width: 0;
	}

	.preview-step-icon {
		display: inline-flex;
		flex: none;
	}

	.preview-step-icon-mobile,
	.preview-step-count-mobile {
		display: none;
	}

	.preview-step-heading h4 {
		margin: 0;
		font-size: 1.125rem;
		font-weight: 600;
		letter-spacing: -0.025em;
		line-height: 1.25;
	}

	.preview-step-heading p {
		margin: 0;
		color: rgb(var(--color-text) / 0.65);
		font-size: 0.75rem;
		line-height: 1.25;
	}

	.preview-step-count {
		align-self: flex-end;
		margin-left: auto;
		color: rgb(var(--color-text) / 0.65);
		font-size: 0.75rem;
		line-height: 1.25;
		white-space: nowrap;
	}

	.preview-step-icon :global([stroke]) {
		stroke-width: 1.5;
	}

	.progress-bars {
		margin-top: 0.3rem;
	}

	.progress-bars i {
		height: 0.125rem;
		border-radius: 999px;
		background: rgb(var(--color-text) / 0.12);
	}

	.progress-bars .current {
		background: rgb(var(--color-primary));
	}

	.schedule {
		grid-template-columns: minmax(15rem, 24rem) minmax(12rem, 1fr);
	}

	.schedule :global(.calendar-header > div:first-child) {
		padding-inline: 0.75rem;
	}

	.schedule :global(.calendar-label) {
		font-size: 1rem;
	}

	.schedule :global(.icon-button) {
		width: 2.25rem;
		height: 2.25rem;
	}

	.schedule :global(.calendar-dates) {
		padding: 0.65rem;
	}

	.schedule :global(.calendar-month) {
		gap: 0.2rem;
	}

	.schedule :global(#preview-timezone) {
		min-height: 2.25rem;
		padding-block: 0.375rem;
	}

	.time-grid button {
		min-height: 2.5rem;
		border: 0;
		border-radius: 0.6rem;
		background: rgb(var(--color-text) / 0.055);
		color: rgb(var(--color-text));
		font-size: 0.75rem;
		font-weight: 650;
		cursor: pointer;
	}

	.time-grid button.selected {
		background: rgb(var(--color-primary));
		color: rgb(var(--color-background));
		box-shadow: inset 0 0 0 2px rgb(var(--color-background) / 0.3);
	}


	@container (max-width: 900px) {
		.preview-page {
			grid-template-columns: 1fr;
		}

		.preview-aside {
			flex-direction: row;
			align-items: end;
			padding: 1.5rem;
		}

		.preview-step-icon-desktop,
		.preview-step-count {
			display: none;
		}

		.preview-step-heading h4 {
			display: flex;
			align-items: center;
			gap: 0.4rem;
		}

		.preview-step-icon-mobile {
			display: inline-flex;
		}

		.preview-step-icon-mobile :global([stroke]) {
			stroke-width: 1.75;
		}

		.preview-step-count-mobile {
			display: inline;
		}
	}

	@container (max-width: 640px) {
		.preview-toolbar {
			grid-template-columns: minmax(0, 1fr) auto;
		}

		.preview-toolbar :global(.segmented-control) {
			grid-column: 1 / -1;
			grid-row: 2;
			justify-self: center;
		}

		.preview-toolbar :global(.preview-visibility-button) {
			grid-column: 2;
			grid-row: 1;
		}

		.preview-aside,
		.preview-actions {
			align-items: stretch;
			flex-direction: column;
		}

		.preview-main {
			padding: 1.25rem;
		}

		.schedule {
			grid-template-columns: 1fr;
		}
	}
</style>
