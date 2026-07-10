import { base } from '$app/paths';
import type { ApiError, PublicConfig, SessionUser } from '$lib/types';

export function appPath(path: string): string {
	return `${base}${path}`;
}

export async function api<T>(path: string, init?: RequestInit): Promise<T> {
	const headers = new Headers(init?.headers);
	if (init?.body && !headers.has('content-type')) {
		headers.set('content-type', 'application/json');
	}

	const response = await fetch(appPath(path), {
		...init,
		headers,
		credentials: 'same-origin'
	});

	if (!response.ok) {
		let message = `Request failed (${response.status})`;
		try {
			const body = (await response.json()) as ApiError;
			if (body.error) message = body.error;
		} catch {
			// Preserve the status-based fallback when the response has no JSON body.
		}
		throw new Error(message);
	}

	if (response.status === 204) return undefined as T;
	return (await response.json()) as T;
}

export function getSession(): Promise<{ user: SessionUser }> {
	return api('/api/auth/session');
}

export function getPublicConfig(): Promise<PublicConfig> {
	return api('/api/config/public');
}
