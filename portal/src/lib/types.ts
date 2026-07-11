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
	brandName: string;
	googleLoginEnabled: boolean;
};

export type ApiError = {
	error: string;
};

export type EventType = {
	eventSlug: string;
	name: string;
	durationMinutes: number;
	bookingWindowDays: number;
	inviteeLimit: number | null;
	timezone: string;
	requiredHostEmails: string[];
	optionalHostEmails: string[];
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
	requiredHosts: { email: string; fullName: string; avatarPath?: string }[];
	optionalHosts: { email: string; fullName: string; avatarPath?: string }[];
	busyRanges: { start: string; end: string }[];
	remainingInvitees: Record<string, number>;
};

export type Booking = {
	id: string;
	eventSlug: string;
	time: string;
	endTime: string;
	attendeeName: string;
	attendeeEmail: string;
	attendeeTimezone: string;
	guestEmails: string[];
	notes?: string;
	title: string;
	recipientEmails: string[];
	createdAt: string;
	updatedAt: string;
	manageURL?: string;
	canceledAt?: string;
	canceledBy?: { name: string; email: string };
	cancellationReason?: string;
};
