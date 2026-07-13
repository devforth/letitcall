import type { PublicEventType, ScheduleDay } from '$lib/types';

export type BookingSlot = {
	time: string;
	label: string;
	busy: boolean;
};

type DateParts = { year: number; month: number; day: number };

const datePartFormatters = new Map<string, Intl.DateTimeFormat>();

function dateParts(value: Date, timezone: string): DateParts & { hour: number; minute: number } {
	let formatter = datePartFormatters.get(timezone);
	if (!formatter) {
		formatter = new Intl.DateTimeFormat('en-CA', {
			timeZone: timezone,
			year: 'numeric',
			month: '2-digit',
			day: '2-digit',
			hour: '2-digit',
			minute: '2-digit',
			hourCycle: 'h23'
		});
		datePartFormatters.set(timezone, formatter);
	}
	const parts = Object.fromEntries(
		formatter.formatToParts(value).map((part) => [part.type, part.value])
	);
	return {
		year: Number(parts.year),
		month: Number(parts.month),
		day: Number(parts.day),
		hour: Number(parts.hour),
		minute: Number(parts.minute)
	};
}

function dateKey(parts: DateParts): string {
	return `${parts.year}-${String(parts.month).padStart(2, '0')}-${String(parts.day).padStart(2, '0')}`;
}

function addDays(parts: DateParts, days: number): DateParts {
	const value = new Date(Date.UTC(parts.year, parts.month - 1, parts.day + days));
	return { year: value.getUTCFullYear(), month: value.getUTCMonth() + 1, day: value.getUTCDate() };
}

function wallTimeToUTC(parts: DateParts, minutes: number, timezone: string): Date {
	const target = Date.UTC(parts.year, parts.month - 1, parts.day, Math.floor(minutes / 60), minutes % 60);
	let guess = new Date(target);
	for (let index = 0; index < 2; index += 1) {
		const actual = dateParts(guess, timezone);
		const represented = Date.UTC(actual.year, actual.month - 1, actual.day, actual.hour, actual.minute);
		guess = new Date(guess.getTime() + target - represented);
	}
	return guess;
}

function minuteOfDay(value: string): number {
	const [hour, minute] = value.split(':').map(Number);
	return hour * 60 + minute;
}

function daySchedule(schedule: ScheduleDay[], parts: DateParts): ScheduleDay {
	const dayIndex = new Date(Date.UTC(parts.year, parts.month - 1, parts.day)).getUTCDay();
	const dayName = ['sunday', 'monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday'][dayIndex];
	return schedule.find((day) => day.day === dayName)!;
}

export function timezoneDateKey(value: Date, timezone: string): string {
	return dateKey(dateParts(value, timezone));
}

export function generateBookingSlots(
	eventType: PublicEventType,
	timezone: string,
	month: string,
	now: Date
): Record<string, BookingSlot[]> {
	const [year, monthNumber] = month.split('-').map(Number);
	const firstCandidate = addDays({ year, month: monthNumber, day: 1 }, -2);
	const lastCandidate = addDays({ year, month: monthNumber + 1, day: 1 }, 2);
	const today = dateParts(now, eventType.timezone);
	const firstBookable = dateKey(today);
	const lastBookable = dateKey(addDays(today, eventType.bookingWindowDays));
	const slots: Record<string, BookingSlot[]> = {};
	const timeFormatter = new Intl.DateTimeFormat(undefined, {
		timeZone: timezone,
		hour: 'numeric',
		minute: '2-digit'
	});

	for (let candidate = firstCandidate; dateKey(candidate) < dateKey(lastCandidate); candidate = addDays(candidate, 1)) {
		const candidateKey = dateKey(candidate);
		if (candidateKey < firstBookable || candidateKey > lastBookable) continue;
		const schedule = daySchedule(eventType.schedule, candidate);
		if (!schedule.enabled) continue;
		const start = minuteOfDay(schedule.start!);
		const end = minuteOfDay(schedule.end!);
		for (let minutes = start; minutes + eventType.durationMinutes <= end; minutes += eventType.durationMinutes) {
			const slotEnd = minutes + eventType.durationMinutes;
			if (schedule.breaks.some((pause) => minutes < minuteOfDay(pause.end) && slotEnd > minuteOfDay(pause.start))) continue;
			const instant = wallTimeToUTC(candidate, minutes, eventType.timezone);
			const instantEnd = new Date(instant.getTime() + eventType.durationMinutes * 60_000);
			const time = instant.toISOString().replace('.000Z', 'Z');
			if (instant <= now) continue;
			const viewerDate = timezoneDateKey(instant, timezone);
			if (!viewerDate.startsWith(month)) continue;
			(slots[viewerDate] ??= []).push({
				time,
				label: timeFormatter.format(instant),
				busy: eventType.busyRanges.some(
					(range) => instant < new Date(range.end) && instantEnd > new Date(range.start)
				)
			});
		}
	}
	for (const daySlots of Object.values(slots)) {
		daySlots.sort((left, right) => left.time.localeCompare(right.time));
	}
	return slots;
}
