import { base } from '$app/paths';
import type { ApiError, PublicConfig, SessionUser } from '$lib/types';
import { showError } from '$lib/notifications';

export function appPath(path: string): string {
	return `${base}${path}`;
}

export function avatarURL(filename: string): string {
	return appPath(`/content/avatars/${filename}`);
}

export async function callApi<T>(path: string, init?: RequestInit, reportError = true): Promise<T> {
	const headers = new Headers(init?.headers);
	if (init?.body && !headers.has('content-type')) {
		headers.set('content-type', 'application/json');
	}

	let response: Response;
	try {
		response = await fetch(appPath(path), {
			...init,
			headers,
			credentials: 'same-origin'
		});
	} catch (cause) {
		const message = cause instanceof Error ? cause.message : 'Unable to reach the server';
		if (reportError) showError(message);
		throw cause;
	}

	if (!response.ok) {
		let message = `Request failed (${response.status})`;
		try {
			const body = (await response.json()) as ApiError;
			if (body.error) message = body.error;
		} catch {
			// Preserve the status fallback for failures outside the API contract.
		}
		if (reportError) showError(message);
		throw new Error(message);
	}

	if (response.status === 204) return undefined as T;
	return (await response.json()) as T;
}

export function getSession(reportError = true): Promise<{ user: SessionUser }> {
	return callApi('/api/auth/session', undefined, reportError);
}

export function getPublicConfig(): Promise<PublicConfig> {
	return callApi('/api/config/public');
}
