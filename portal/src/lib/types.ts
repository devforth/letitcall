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
