import { writable } from 'svelte/store';

export type Notification = {
	id: number;
	message: string;
};

export const notifications = writable<Notification[]>([]);

let nextID = 1;

export function showError(message: string) {
	const id = nextID++;
	notifications.update((items) => [{ id, message }, ...items]);
	window.setTimeout(() => dismissNotification(id), 6000);
}

export function dismissNotification(id: number) {
	notifications.update((items) => items.filter((item) => item.id !== id));
}
