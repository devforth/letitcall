export type SessionUser = {
	email: string;
	fullName: string;
	timezone: string;
	googleConnected: boolean;
	avatarPath?: string;
};

export type ManagedUser = SessionUser & {
	createdAt: string;
	updatedAt: string;
};

export type PublicConfig = {
	googleLoginEnabled: boolean;
};

export type ApiError = {
	error: string;
};

export type EventType = {
	eventSlug: string;
	name: string;
	durationMinutes: number;
	bookingWindowDays: number | null;
	inviteeLimit: number | null;
	timezone: string;
	recipientEmails: string[];
	schedule: ScheduleDay[];
	createdBy: string;
	createdAt: string;
	updatedAt: string;
};

export type ScheduleDay = {
	day: string;
	enabled: boolean;
	start?: string;
	end?: string;
	breaks: TimeRange[];
};

export type TimeRange = {
	start: string;
	end: string;
};

export type PublicEventType = Pick<
	EventType,
	| 'eventSlug'
	| 'name'
	| 'durationMinutes'
	| 'bookingWindowDays'
	| 'inviteeLimit'
	| 'timezone'
	| 'schedule'
> & {
	hosts: { email: string; avatarPath?: string }[];
	unavailableTimes: string[];
};

export type Booking = {
	id: string;
	eventSlug: string;
	time: string;
	endTime: string;
	attendeeName: string;
	attendeeEmail: string;
	notes?: string;
	title: string;
	recipientEmails: string[];
	createdAt: string;
};
