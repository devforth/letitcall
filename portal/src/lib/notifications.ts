import { writable } from 'svelte/store';

export type NotificationVariant = 'info' | 'error' | 'success';

export type Notification = {
	id: number;
	message: string;
	variant: NotificationVariant;
	progress?: number;
	subtitle?: string;
};

export const notifications = writable<Notification[]>([]);

export const notificationDurationMs = 6000;

let nextID = 1;

function showNotification(variant: NotificationVariant, message: string) {
	const id = nextID++;
	notifications.update((items) => [{ id, message, variant }, ...items]);
	window.setTimeout(() => dismissNotification(id), notificationDurationMs);
}

export function showInfo(message: string) {
	showNotification('info', message);
}

export function showError(message: string) {
	showNotification('error', message);
}

export function showSuccess(message: string) {
	showNotification('success', message);
}

export function dismissNotification(id: number) {
	notifications.update((items) => items.filter((item) => item.id !== id));
}

export function showProgress(message: string, variant: NotificationVariant = 'info', subtitle?: string) {
	const id = nextID++;
	notifications.update((items) => [{ id, message, variant, progress: 0, subtitle }, ...items]);
	return id;
}

export function updateProgress(id: number, progress: number) {
	notifications.update((items) =>
		items.map((item) => (item.id === id ? { ...item, progress: Math.min(100, Math.max(0, progress)) } : item))
	);
}

export function updateProgressWithSubtitle(id: number, progress: number, subtitle: string) {
	notifications.update((items) =>
		items.map((item) =>
			item.id === id ? { ...item, progress: Math.min(100, Math.max(0, progress)), subtitle } : item
		)
	);
}
