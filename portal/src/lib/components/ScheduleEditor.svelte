<script lang="ts">
	import Icon from '@iconify/svelte';
	import plusIcon from '@iconify-icons/tabler/plus';
	import xIcon from '@iconify-icons/tabler/x';
	import AvailabilityCopyMenu from '$lib/components/AvailabilityCopyMenu.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Checkbox from '$lib/components/ui/Checkbox.svelte';
	import IconButton from '$lib/components/ui/IconButton.svelte';
	import TimeInput from '$lib/components/ui/TimeInput.svelte';
	import type { ScheduleDay, TimeRange } from '$lib/types';

	let { schedule = $bindable() }: { schedule: ScheduleDay[] } = $props();

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

<section class="grid gap-4 border border-black p-4" aria-labelledby="schedule-title">
	<div>
		<h2 id="schedule-title" class="font-semibold">Weekly schedule</h2>
		<p class="mt-1 text-xs">Start with one range, then customize only the days that differ.</p>
	</div>

	<div class="grid gap-4 border border-black p-4">
		<h3 class="text-sm font-semibold">Quick setup</h3>
		<div class="flex flex-wrap gap-x-6">
			<Checkbox id="quick-weekdays" label="Weekdays" bind:checked={applyWeekdays} />
			<Checkbox id="quick-weekends" label="Weekends" bind:checked={applyWeekends} />
		</div>
		<div class="grid gap-3 sm:grid-cols-[1fr_1fr_auto] sm:items-end">
			<TimeInput id="quick-start" label="From" bind:value={quickStart} />
			<TimeInput id="quick-end" label="To" bind:value={quickEnd} />
			<Button variant="secondary" onclick={applyQuickHours}>Apply to selected days</Button>
		</div>
	</div>

	<details class="border border-black">
		<summary class="cursor-pointer px-4 py-3 text-sm font-medium">Customize individual days</summary>
		<div class="grid border-t border-black">
			{#each schedule as day (day.day)}
				{@const ranges = availabilityRanges(day)}
				<div class="grid gap-4 border-b border-black p-4 last:border-b-0 lg:grid-cols-[9rem_1fr_auto] lg:items-start">
					<div class="flex min-h-11 items-center gap-3">
						<span class="grid size-9 shrink-0 place-items-center rounded-full bg-black text-xs font-semibold text-white" aria-hidden="true">
							{labels[day.day].slice(0, 1)}
						</span>
						<span class="text-sm font-medium">{labels[day.day]}</span>
					</div>

					{#if day.enabled}
						<div class="grid gap-3">
							{#each ranges as range, index (`${day.day}-${index}`)}
								<div class="grid gap-3 sm:grid-cols-[1fr_auto_1fr_auto] sm:items-end">
									<TimeInput
										id={`${day.day}-${index}-start`}
										label="From"
										value={range.start}
										onchange={(value) => updateRange(day, index, 'start', value)}
									/>
									<span class="hidden min-h-11 items-center sm:flex" aria-hidden="true">–</span>
									<TimeInput
										id={`${day.day}-${index}-end`}
										label="To"
										value={range.end}
										onchange={(value) => updateRange(day, index, 'end', value)}
									/>
									<IconButton tone="danger" label={`Remove ${labels[day.day]} range ${index + 1}`} onclick={() => removeRange(day, index)}>
										<Icon icon={xIcon} width="22" height="22" />
									</IconButton>
								</div>
							{/each}
						</div>
					{:else}
						<p class="flex min-h-11 items-center text-sm">Unavailable</p>
					{/if}

					<div class="flex gap-2 lg:pt-6">
						<IconButton tone="primary" label={`Add ${labels[day.day]} range`} onclick={() => addRange(day)}>
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
</section>
