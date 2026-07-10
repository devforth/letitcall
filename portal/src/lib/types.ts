export type SessionUser = {
	email: string;
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
