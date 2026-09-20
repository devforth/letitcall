<script lang="ts">
	import Icon from '@iconify/svelte';
	import externalLinkIcon from '@iconify-icons/charm/link-external';
	import dotsVerticalIcon from '@iconify-icons/tabler/dots-vertical';
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
			dateStyle: 'long'
		}).format(new Date(value));
	}

	function localTime(value: string): string {
		return new Intl.DateTimeFormat(undefined, { timeStyle: 'short' }).format(new Date(value));
	}

	function localWeekday(value: string): string {
		return new Intl.DateTimeFormat(undefined, { weekday: 'long' }).format(new Date(value));
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
		<article class="booking-row relative grid gap-4 px-4 pb-4 pt-3 sm:grid-cols-[13.5rem_minmax(0,1fr)_auto] sm:items-start sm:gap-x-8 sm:px-5 sm:pb-5 sm:pt-4">
			<div class="hidden justify-items-end gap-1 sm:grid">
				<span class="status-chip mb-2 mt-[3px] inline-flex sm:justify-self-start" class:canceled={!!booking.canceledAt}>
					{historical && booking.canceledAt ? 'Canceled' : relativeTime(booking.time)}
				</span>
				<p class="text-sm font-bold" style="color: rgb(var(--color-text) / 0.75);">{localDate(booking.time)}</p>
				<p class="text-sm font-bold" style="color: rgb(var(--color-text) / 0.75);">{localTime(booking.time)}</p>
				<p class="text-xs" style="color: rgb(var(--color-text) / 0.65);">{localWeekday(booking.time)}</p>
			</div>

			<div class="min-w-0">
				<div class="mb-1 flex min-w-0 items-center gap-2 pr-14 sm:mb-3 sm:pr-0">
					<span class="status-chip mt-[3px] shrink-0 sm:hidden" class:canceled={!!booking.canceledAt}>
						{historical && booking.canceledAt ? 'Canceled' : relativeTime(booking.time)}
					</span>
					<span class="shrink-0 sm:hidden" style="color: rgb(var(--color-text) / 0.4);" aria-hidden="true">·</span>
					<h3 class="min-w-0 truncate font-semibold" style="color: rgb(var(--color-text));">{booking.title}</h3>
				</div>
				<p class="mb-3 text-sm font-bold sm:hidden" style="color: rgb(var(--color-text) / 0.75);">
					{localWeekday(booking.time)}, {localDate(booking.time)} on {localTime(booking.time)}
				</p>
				<div
					class="sm:-ml-4 sm:border-l sm:pl-[15px]"
					style="border-color: color-mix(in srgb, var(--color-border) 65%, transparent);"
				>
					<p class="mt-0.5 truncate text-sm" style="color: rgb(var(--color-text) / 0.65);">
						{booking.attendeeName} · {booking.attendeeEmail}
					</p>
					<div class="mt-3"><HostBadges hosts={bookingHosts(booking)} {users} /></div>
				</div>
			</div>

			<div class="booking-action-slot absolute right-4 top-3 flex items-center justify-end sm:relative sm:right-auto sm:top-auto sm:self-start">
				{#if booking.manageURL}
					<span class="booking-action-hint" aria-hidden="true">
						<Icon icon={dotsVerticalIcon} width="22" height="22" />
					</span>
					<a
						class="booking-action"
						href={booking.manageURL}
						aria-label={`Manage ${booking.title}`}
						title="Open booking page"
					>
						<Icon icon={externalLinkIcon} width="20" height="20" />
					</a>
				{/if}
			</div>
		</article>
	{/each}
</div>

<style>
	.booking-row {
		border-bottom: 1px solid var(--color-border);
		transition: background 0.15s ease;
	}

	.booking-row:last-child {
		border-bottom: 0;
	}

	.booking-row:hover {
		background: rgb(var(--color-primary) / 0.045);
	}

	.status-chip {
		border: 1px solid var(--color-border);
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

	.booking-action {
		display: grid;
		width: 2.5rem;
		height: 2.5rem;
		place-items: center;
		border-radius: 10px;
		background: rgb(var(--color-primary) / 0.12);
		color: rgb(var(--color-primary));
		transition: background 0.15s ease;
	}

	.booking-action-hint {
		display: none;
	}

	@media (min-width: 40rem) {
		.booking-action-slot {
			display: flex;
			min-width: 2.5rem;
			min-height: 2.5rem;
			align-items: center;
			justify-content: flex-end;
		}

		.booking-action-hint {
			position: absolute;
			right: 0;
			display: grid;
			width: 2.5rem;
			height: 2.5rem;
			place-items: center;
			color: rgb(var(--color-text) / 0.6);
			pointer-events: none;
			transition:
				opacity 0.18s ease,
				transform 0.18s ease;
		}

		.booking-action {
			opacity: 0;
			pointer-events: none;
			transform: translateX(0.5rem);
			transition:
				opacity 0.18s ease,
				transform 0.18s ease;
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
	}

	.booking-action:focus-visible {
		outline: 2px solid rgb(var(--color-text) / 0.65);
		outline-offset: 2px;
	}

	@media (prefers-reduced-motion: reduce) {
		.booking-action,
		.booking-action-hint {
			transition: none;
		}
	}

</style>
