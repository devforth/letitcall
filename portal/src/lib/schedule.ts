import type { ScheduleDay, TimeRange } from '$lib/types';

export function availabilityRanges(day: ScheduleDay): TimeRange[] {
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

const wallTimeFormat = new Intl.DateTimeFormat('en-US', { hour: '2-digit', minute: '2-digit', timeZone: 'UTC' });

export function formatWallTime(value: string): string {
	const [hours, minutes] = value.split(':').map(Number);
	return wallTimeFormat.format(Date.UTC(1970, 0, 1, hours, minutes));
}
