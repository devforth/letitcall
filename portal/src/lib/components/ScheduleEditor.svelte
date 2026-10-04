<script lang="ts">
	import Icon from '@iconify/svelte';
	import chevronRightIcon from '@iconify-icons/tabler/chevron-right';
	import copyIcon from '@iconify-icons/tabler/copy';
	import plusIcon from '@iconify-icons/tabler/plus';
	import xIcon from '@iconify-icons/tabler/x';
	import AvailabilityCopyMenu from '$lib/components/AvailabilityCopyMenu.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Checkbox from '$lib/components/ui/Checkbox.svelte';
	import IconButton from '$lib/components/ui/IconButton.svelte';
	import TimeInput from '$lib/components/ui/TimeInput.svelte';
	import { availabilityRanges } from '$lib/schedule';
	import type { ScheduleDay, TimeRange } from '$lib/types';

	let {
		schedule = $bindable(),
		stepMinutes
	}: {
		schedule: ScheduleDay[];
		stepMinutes: number;
	} = $props();

	let applyWeekdays = $state(true);
	let applyWeekends = $state(false);
	let quickStart = $state('10:00');
	let quickEnd = $state('16:00');
	const quickInvalid = $derived(endsBeforeStart({ start: quickStart, end: quickEnd }));

	const labels: Record<string, string> = {
		monday: 'Monday',
		tuesday: 'Tuesday',
		wednesday: 'Wednesday',
		thursday: 'Thursday',
		friday: 'Friday',
		saturday: 'Saturday',
		sunday: 'Sunday'
	};

	function applyQuickHours() {
		for (const day of schedule) {
			const weekend = day.day === 'saturday' || day.day === 'sunday';
			const enabled = weekend ? applyWeekends : applyWeekdays;
			day.enabled = enabled;
			day.start = enabled ? quickStart : undefined;
			day.end = enabled ? quickEnd : undefined;
			day.breaks = [];
		}
	}

	function endsBeforeStart(range: TimeRange) {
		return !!range.start && !!range.end && range.end <= range.start;
	}

	function availabilityHours(day: ScheduleDay) {
		const minutes = availabilityRanges(day).reduce((total, range) => {
			if (!range.start || !range.end) return total;
			const [startHours, startMinutes] = range.start.split(':').map(Number);
			const [endHours, endMinutes] = range.end.split(':').map(Number);
			return total + endHours * 60 + endMinutes - startHours * 60 - startMinutes;
		}, 0);
		return minutes % 60 ? `${Math.floor(minutes / 60)}h ${minutes % 60}m` : `${minutes / 60}h`;
	}

	function setAvailabilityRanges(day: ScheduleDay, ranges: TimeRange[]) {
		if (ranges.length === 0) {
			day.enabled = false;
			day.start = '';
			day.end = '';
			day.breaks = [];
			return;
		}
		day.enabled = true;
		day.start = ranges[0].start;
		day.end = ranges[ranges.length - 1].end;
		day.breaks = ranges.slice(0, -1).map((range, index) => ({
			start: range.end,
			end: ranges[index + 1].start
		}));
	}

	function updateRange(day: ScheduleDay, index: number, field: keyof TimeRange, value: string) {
		const ranges = availabilityRanges(day);
		ranges[index] = { ...ranges[index], [field]: value };
		setAvailabilityRanges(day, ranges);
	}

	function addRange(day: ScheduleDay) {
		const ranges = availabilityRanges(day);
		ranges.push(day.enabled ? { start: '', end: '' } : { start: quickStart, end: quickEnd });
		setAvailabilityRanges(day, ranges);
	}

	function removeRange(day: ScheduleDay, index: number) {
		setAvailabilityRanges(
			day,
			availabilityRanges(day).filter((_, rangeIndex) => rangeIndex !== index)
		);
	}

	function copyRanges(source: ScheduleDay, targetDays: string[]) {
		const ranges = availabilityRanges(source);
		for (const target of schedule.filter(({ day }) => targetDays.includes(day))) {
			setAvailabilityRanges(target, ranges.map((range) => ({ ...range })));
		}
	}
</script>

<section
	class="grid gap-4 pt-4 pb-0 sm:pt-5"
	aria-labelledby="schedule-title"
>
	<div class="grid gap-1">
		<h2 id="schedule-title" class="font-semibold" style="color: rgb(var(--color-text));">Weekly schedule</h2>
		<p class="text-sm" style="color: rgb(var(--color-text) / 0.65);">Start with one range, then customize only the days that differ</p>
	</div>

	<div class="rounded-md border" style="border-color: var(--color-border);">
		<details open>
			<summary
				id="quick-preset-title"
				class="schedule-summary flex cursor-pointer items-center gap-2 px-3 py-3 text-sm font-normal"
				style="background: rgb(var(--color-text) / 0.06); color: rgb(var(--color-text));"
			>
				<Icon icon={chevronRightIcon} width="18" height="18" class="details-chevron" aria-hidden="true" />
				Set quick preset
			</summary>
		<div
			class="grid gap-4 border-t p-4"
			aria-labelledby="quick-preset-title"
			style={`border-color: var(--color-border);`}
		>
			<div class="flex flex-wrap gap-x-6">
				<Checkbox id="quick-weekdays" label="Weekdays" bind:checked={applyWeekdays} />
				<Checkbox id="quick-weekends" label="Weekends" bind:checked={applyWeekends} />
			</div>
			<div class="grid gap-3 sm:grid-cols-[12rem_12rem_auto] sm:items-center">
				<TimeInput id="quick-start" label="From" bind:value={quickStart} step={stepMinutes * 60} />
				<TimeInput id="quick-end" label="To" bind:value={quickEnd} invalid={quickInvalid} step={stepMinutes * 60} />
				<Button
					class="outlined-action-button justify-self-start"
					variant="primary-outline"
					style="padding-right: 1.125rem !important; padding-bottom: 0.5rem !important;"
					disabled={quickInvalid}
					onclick={applyQuickHours}
				>
					<span class="flex items-center gap-2">
						<Icon icon={copyIcon} width="20" height="20" />
						Apply
					</span>
				</Button>
			</div>
		</div>
		</details>

		<details class="border-t" style="border-color: var(--color-border);">
		<summary
			class="schedule-summary flex cursor-pointer items-center gap-2 px-3 py-3 text-sm font-normal"
			style="background: rgb(var(--color-text) / 0.06); color: rgb(var(--color-text));"
		>
			<Icon icon={chevronRightIcon} width="18" height="18" class="details-chevron" aria-hidden="true" />
			Customize individual days
		</summary>
		<div class="grid border-t" style="border-color: var(--color-border);">
			{#each schedule as day (day.day)}
				{@const ranges = availabilityRanges(day)}
				<div class="grid grid-cols-[minmax(0,1fr)_auto] gap-4 border-b p-4 last:border-b-0 lg:grid-cols-[9rem_1fr_auto] lg:items-start" style="border-color: var(--color-border);">
					<div class="flex min-h-11 flex-col justify-center">
						<span class="text-sm font-medium" style="color: rgb(var(--color-text));">{labels[day.day]}</span>
						{#if day.enabled}
							<span class="text-xs" style="color: rgb(var(--color-text) / 0.65);">{availabilityHours(day)}</span>
						{/if}
					</div>

					{#if day.enabled}
						<div class="order-last col-span-2 grid gap-3 lg:order-none lg:col-span-1 lg:grid-cols-[auto_1fr] lg:items-center">
							<div class="grid gap-3">
								{#each ranges as range, index (`${day.day}-${index}`)}
									<div class="grid grid-cols-[minmax(0,7.5rem)_minmax(0,7.5rem)_auto] items-center justify-start gap-3">
										<TimeInput
											id={`${day.day}-${index}-start`}
											label="From"
											value={range.start}
											step={stepMinutes * 60}
											onchange={(value) => updateRange(day, index, 'start', value)}
										/>
										<TimeInput
											id={`${day.day}-${index}-end`}
											label="To"
											value={range.end}
											step={stepMinutes * 60}
											invalid={endsBeforeStart(range)}
											onchange={(value) => updateRange(day, index, 'end', value)}
										/>
										<IconButton filled tone="danger" label={`Remove ${labels[day.day]} range ${index + 1}`} onclick={() => removeRange(day, index)}>
											<Icon icon={xIcon} width="22" height="22" />
										</IconButton>
									</div>
								{/each}
								</div>
						</div>
					{:else}
						<p class="hidden min-h-11 items-center text-sm lg:flex" style="color: rgb(var(--color-text) / 0.65);">—</p>
					{/if}

					<div class="flex min-h-11 items-center gap-2 lg:w-[5.5rem]">
						<IconButton filled tone="primary" label={`Add ${labels[day.day]} range`} onclick={() => addRange(day)}>
							<Icon icon={plusIcon} width="22" height="22" />
						</IconButton>
						{#if day.enabled}
							<AvailabilityCopyMenu
								sourceDay={day.day}
								days={schedule.filter((target) => target.day !== day.day).map((target) => ({
									day: target.day,
									label: labels[target.day]
								}))}
								oncopy={(days) => copyRanges(day, days)}
							/>
						{/if}
					</div>
				</div>
			{/each}
		</div>
		</details>
	</div>
</section>

<style>
	.schedule-summary::-webkit-details-marker {
		display: none;
	}

	.details-chevron {
		transition: transform 0.15s ease;
	}

	details[open] :global(.details-chevron) {
		transform: rotate(90deg);
	}
</style>
