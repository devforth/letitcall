<script lang="ts">
	import Button from '$lib/components/ui/Button.svelte';
	import Checkbox from '$lib/components/ui/Checkbox.svelte';
	import TimeInput from '$lib/components/ui/TimeInput.svelte';
	import type { ScheduleDay } from '$lib/types';

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

	function addBreak(day: ScheduleDay) {
		const start = toMinutes(day.start ?? '10:00');
		const end = toMinutes(day.end ?? '16:00');
		let pauseStart = 12 * 60;
		let pauseEnd = 13 * 60;
		if (pauseStart <= start || pauseEnd >= end) {
			const midpoint = Math.floor((start + end) / 30) * 15;
			pauseStart = Math.max(start + 15, midpoint - 30);
			pauseEnd = Math.min(end - 15, pauseStart + 60);
		}
		day.breaks = [{ start: fromMinutes(pauseStart), end: fromMinutes(pauseEnd) }];
	}

	function toMinutes(value: string) {
		const [hours, minutes] = value.split(':').map(Number);
		return hours * 60 + minutes;
	}

	function fromMinutes(value: number) {
		return `${String(Math.floor(value / 60)).padStart(2, '0')}:${String(value % 60).padStart(2, '0')}`;
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
				<div class="grid gap-3 border-b border-black p-4 last:border-b-0 lg:grid-cols-[9rem_1fr]">
					<Checkbox id={`schedule-${day.day}`} label={labels[day.day]} bind:checked={day.enabled} />
					{#if day.enabled}
						<div class="grid gap-3">
							<div class="grid gap-3 sm:grid-cols-2">
								<TimeInput id={`${day.day}-start`} label="From" bind:value={day.start} />
								<TimeInput id={`${day.day}-end`} label="To" bind:value={day.end} />
							</div>
							{#each day.breaks as pause, index (`${day.day}-${index}`)}
								<div class="grid gap-3 border border-black p-3 sm:grid-cols-[1fr_1fr_auto] sm:items-end">
									<TimeInput id={`${day.day}-break-${index}-start`} label="Break from" bind:value={pause.start} />
									<TimeInput id={`${day.day}-break-${index}-end`} label="Break to" bind:value={pause.end} />
									<Button variant="danger" onclick={() => (day.breaks = [])}>Remove break</Button>
								</div>
							{/each}
							{#if day.breaks.length === 0}
								<div><Button variant="secondary" onclick={() => addBreak(day)}>Add break</Button></div>
							{/if}
						</div>
					{/if}
				</div>
			{/each}
		</div>
	</details>
</section>
