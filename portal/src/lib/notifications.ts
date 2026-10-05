import { writable } from 'svelte/store';

export type NotificationVariant = 'error' | 'success';

export type Notification = {
	id: number;
	message: string;
	variant: NotificationVariant;
};

export const notifications = writable<Notification[]>([]);

export const notificationDurationMs = 6000;

let nextID = 1;

function showNotification(variant: NotificationVariant, message: string) {
	const id = nextID++;
	notifications.update((items) => [{ id, message, variant }, ...items]);
	window.setTimeout(() => dismissNotification(id), notificationDurationMs);
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
