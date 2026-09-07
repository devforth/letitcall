<script lang="ts">
	import Icon from '@iconify/svelte';
	import calendarEventIcon from '@iconify-icons/tabler/calendar-event';
	import dotsIcon from '@iconify-icons/tabler/dots-vertical';
	import externalLinkIcon from '@iconify-icons/charm/link-external';
	import HostBadges from '$lib/components/HostBadges.svelte';
	import type { Booking, EventType, ManagedUser } from '$lib/types';

	let {
		bookings,
		eventTypes,
		users,
		now,
		historical = false
	}: {
		bookings: Booking[];
		eventTypes: EventType[];
		users: ManagedUser[];
		now: Date;
		historical?: boolean;
	} = $props();

	function relativeTime(value: string): string {
		const seconds = (new Date(value).getTime() - now.getTime()) / 1000;
		const formatter = new Intl.RelativeTimeFormat(undefined, { numeric: 'always' });
		if (Math.abs(seconds) < 3600) return formatter.format(Math.round(seconds / 60), 'minute');
		if (Math.abs(seconds) < 86_400) return formatter.format(Math.round(seconds / 3600), 'hour');
		return formatter.format(Math.round(seconds / 86_400), 'day');
	}

	function localDate(value: string): string {
		return new Intl.DateTimeFormat(undefined, {
			dateStyle: 'full',
			timeStyle: 'short'
		}).format(new Date(value));
	}

	function bookingHosts(booking: Booking) {
		const eventType = eventTypes.find((candidate) => candidate.eventSlug === booking.eventSlug);
		return booking.recipientEmails.map((email) => ({
			email,
			role: eventType?.requiredHostEmails.includes(email)
				? ('Required' as const)
				: eventType?.optionalHostEmails.includes(email)
					? ('Optional' as const)
					: ('Host' as const)
		}));
	}
</script>

<div class="booking-list">
	{#each bookings as booking (booking.id)}
		<article class="booking-row grid gap-4 p-4 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center sm:p-5">
			<div class="flex min-w-0 items-start gap-3">
				<div class="booking-icon grid size-10 shrink-0 place-items-center rounded-xl" aria-hidden="true">
					<Icon icon={calendarEventIcon} width="20" height="20" />
				</div>
				<div class="min-w-0">
					<h3 class="truncate font-semibold" style="color: rgb(var(--color-text));">{booking.title}</h3>
					<p class="mt-0.5 truncate text-xs" style="color: rgb(var(--color-text) / 0.65);">
						{booking.attendeeName} · {booking.attendeeEmail}
					</p>
					<p class="mt-2 text-sm" style="color: rgb(var(--color-text) / 0.75);">{localDate(booking.time)}</p>
					<div class="mt-3"><HostBadges hosts={bookingHosts(booking)} {users} /></div>
				</div>
			</div>

			<div class="flex items-center justify-end gap-3">
				<span class="status-chip" class:canceled={!!booking.canceledAt}>
					{historical && booking.canceledAt ? 'Canceled' : relativeTime(booking.time)}
				</span>
				{#if booking.manageURL}
					<div class="action-slot">
						<span class="booking-action-hint" aria-hidden="true">
							<Icon icon={dotsIcon} width="22" height="22" />
						</span>
						<a
							class="booking-action"
							href={booking.manageURL}
							aria-label={`Manage ${booking.title}`}
							title="Open booking page"
						>
							<Icon icon={externalLinkIcon} width="20" height="20" />
						</a>
					</div>
				{/if}
			</div>
		</article>
	{/each}
</div>

<style>
	.booking-row {
		border-bottom: 1px solid rgb(var(--color-border));
		transition: background 0.15s ease;
	}

	.booking-row:last-child {
		border-bottom: 0;
	}

	.booking-row:hover {
		background: rgb(var(--color-primary) / 0.045);
	}

	.booking-icon {
		background: rgb(var(--color-primary) / 0.14);
		color: rgb(var(--color-primary));
	}

	.status-chip {
		border: 1px solid rgb(var(--color-border));
		border-radius: 999px;
		padding: 0.25rem 0.5rem;
		color: rgb(var(--color-text) / 0.65);
		font-size: 0.75rem;
		font-weight: 600;
		line-height: 1;
		white-space: nowrap;
	}

	.status-chip.canceled {
		color: rgb(var(--error));
	}

	.action-slot {
		position: relative;
		display: grid;
		width: 2.5rem;
		height: 2.5rem;
		place-items: center;
	}

	.booking-action-hint,
	.booking-action {
		position: absolute;
		display: grid;
		width: 2.5rem;
		height: 2.5rem;
		place-items: center;
		border-radius: 10px;
		transition:
			opacity 0.18s ease,
			transform 0.18s ease;
	}

	.booking-action-hint {
		color: rgb(var(--color-text) / 0.6);
		pointer-events: none;
	}

	.booking-action {
		background: rgb(var(--color-text) / 0.08);
		color: rgb(var(--color-text) / 0.65);
		opacity: 0;
		pointer-events: none;
		transform: translateX(0.5rem);
	}

	.booking-row:hover .booking-action,
	.booking-row:focus-within .booking-action {
		opacity: 1;
		pointer-events: auto;
		transform: translateX(0);
	}

	.booking-row:hover .booking-action-hint,
	.booking-row:focus-within .booking-action-hint {
		opacity: 0;
		transform: translateX(-0.5rem) scale(0.85);
	}

	.booking-action:focus-visible {
		outline: 2px solid rgb(var(--color-text) / 0.65);
		outline-offset: 2px;
	}

	@media (prefers-reduced-motion: reduce) {
		.booking-action-hint,
		.booking-action {
			transition: none;
		}
	}
</style>
