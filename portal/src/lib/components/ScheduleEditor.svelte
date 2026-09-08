<script lang="ts">
	import Icon from '@iconify/svelte';
	import chevronRightIcon from '@iconify-icons/tabler/chevron-right';
	import plusIcon from '@iconify-icons/tabler/plus';
	import xIcon from '@iconify-icons/tabler/x';
	import AvailabilityCopyMenu from '$lib/components/AvailabilityCopyMenu.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Checkbox from '$lib/components/ui/Checkbox.svelte';
	import IconButton from '$lib/components/ui/IconButton.svelte';
	import TimeInput from '$lib/components/ui/TimeInput.svelte';
	import type { ScheduleDay, TimeRange } from '$lib/types';

	let {
		schedule = $bindable(),
		embedded = false
	}: {
		schedule: ScheduleDay[];
		embedded?: boolean;
	} = $props();

	let applyWeekdays = $state(true);
	let applyWeekends = $state(false);
	let quickStart = $state('10:00');
	let quickEnd = $state('16:00');

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

	function availabilityRanges(day: ScheduleDay): TimeRange[] {
		if (!day.enabled) return [];
		const ranges: TimeRange[] = [];
		let start = day.start ?? '';
		for (const pause of day.breaks) {
			ranges.push({ start, end: pause.start });
			start = pause.end;
		}
		ranges.push({ start, end: day.end ?? '' });
		return ranges;
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

{#snippet durationChip(day: ScheduleDay)}
	<span class="inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium" style="background: rgb(var(--color-foreground)); border-color: rgb(var(--color-border)); color: rgb(var(--color-text) / 0.65);">
		{availabilityHours(day)}
	</span>
{/snippet}

<section
	class={`grid gap-4 px-4 pt-4 pb-0 sm:px-5 sm:pt-5 ${embedded ? '' : 'rounded-lg border-2'}`}
	aria-labelledby="schedule-title"
	style={embedded
		? undefined
		: 'background: rgb(var(--color-foreground)); border-color: rgb(var(--color-border)); box-shadow: var(--shadow-small);'}
>
	<div class="grid gap-1">
		<h2 id="schedule-title" class="font-semibold" style="color: rgb(var(--color-text));">Weekly schedule</h2>
		<p class="text-sm" style="color: rgb(var(--color-text) / 0.65);">Start with one range, then customize only the days that differ.</p>
	</div>

	<div class="overflow-hidden rounded-md border" style="border-color: rgb(var(--color-border));">
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
			style={`border-color: rgb(var(--color-border)); ${embedded ? '' : 'background: rgb(var(--color-text) / 0.035);'}`}
		>
			<div class="flex flex-wrap gap-x-6">
				<Checkbox id="quick-weekdays" label="Weekdays" bind:checked={applyWeekdays} />
				<Checkbox id="quick-weekends" label="Weekends" bind:checked={applyWeekends} />
			</div>
			<div class="grid gap-3 sm:grid-cols-[12rem_12rem_auto] sm:items-center">
				<TimeInput id="quick-start" label="From" bind:value={quickStart} />
				<TimeInput id="quick-end" label="To" bind:value={quickEnd} />
				<Button class="justify-self-start" variant="secondary" onclick={applyQuickHours}>Apply</Button>
			</div>
		</div>
		</details>

		<details class="border-t" style="border-color: rgb(var(--color-border));">
		<summary
			class="schedule-summary flex cursor-pointer items-center gap-2 px-3 py-3 text-sm font-normal"
			style="background: rgb(var(--color-text) / 0.06); color: rgb(var(--color-text));"
		>
			<Icon icon={chevronRightIcon} width="18" height="18" class="details-chevron" aria-hidden="true" />
			Customize individual days
		</summary>
		<div class="grid border-t" style="border-color: rgb(var(--color-border));">
			{#each schedule as day (day.day)}
				{@const ranges = availabilityRanges(day)}
				<div class="grid gap-4 border-b p-4 last:border-b-0 lg:grid-cols-[9rem_1fr_auto] lg:items-start" style="border-color: rgb(var(--color-border));">
					<div class="flex min-h-11 items-center gap-2">
						<span class="text-sm font-medium" style="color: rgb(var(--color-text));">{labels[day.day]}</span>
						{#if day.enabled && ranges.length > 1}
							{@render durationChip(day)}
						{/if}
					</div>

					{#if day.enabled}
						<div class="grid gap-3 lg:grid-cols-[auto_1fr] lg:items-center">
							<div class="grid gap-3">
								{#each ranges as range, index (`${day.day}-${index}`)}
									<div class="grid gap-3 sm:grid-cols-[7.5rem_7.5rem_auto] sm:items-center sm:justify-start">
										<TimeInput
											id={`${day.day}-${index}-start`}
											label="From"
											value={range.start}
											onchange={(value) => updateRange(day, index, 'start', value)}
										/>
										<TimeInput
											id={`${day.day}-${index}-end`}
											label="To"
											value={range.end}
											onchange={(value) => updateRange(day, index, 'end', value)}
										/>
										<IconButton filled tone="danger" label={`Remove ${labels[day.day]} range ${index + 1}`} onclick={() => removeRange(day, index)}>
											<Icon icon={xIcon} width="22" height="22" />
										</IconButton>
									</div>
								{/each}
								</div>
							{#if ranges.length === 1}
								<div class="w-20 justify-self-center">
									{@render durationChip(day)}
								</div>
							{/if}
						</div>
					{:else}
						<p class="flex min-h-11 items-center text-sm" style="color: rgb(var(--color-text) / 0.65);">Unavailable</p>
					{/if}

					<div class="flex min-h-11 items-center gap-2">
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
