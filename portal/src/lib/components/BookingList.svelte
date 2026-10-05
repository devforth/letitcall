<script lang="ts">
	import Icon from '@iconify/svelte';
	import externalLinkIcon from '@iconify-icons/charm/link-external';
	import HostBadges from '$lib/components/HostBadges.svelte';
	import TimedActions from '$lib/components/ui/TimedActions.svelte';
	import type { Booking, EventType, ManagedUser } from '$lib/types';
	import { browserTimeFormatter } from '$lib/time-format';

	let {
		bookings,
		eventTypes,
		users,
		now
	}: {
		bookings: Booking[];
		eventTypes: EventType[];
		users: ManagedUser[];
		now: Date;
	} = $props();

	function localTime(value: string): string {
		return browserTimeFormatter({ timeStyle: 'short' }).format(new Date(value));
	}

	function localYear(value: string): string {
		return new Intl.DateTimeFormat(undefined, { year: 'numeric' }).format(new Date(value));
	}

	function localDayMonth(value: string): string {
		return new Intl.DateTimeFormat(undefined, { day: 'numeric', month: 'short' }).format(new Date(value));
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

<div class="grid gap-5">
	{#each bookings as booking (booking.id)}
		<article
			class="booking-row grid items-start gap-x-4 gap-y-2 rounded-lg p-4 sm:gap-y-4 sm:grid-cols-[8.75rem_minmax(0,1fr)_2.5rem] sm:px-5 sm:pb-5 sm:pt-4"
			data-timed-actions-row
		>
			<div class="-mx-4 -mt-4 sm:m-0">
				<time class="booking-date-card" datetime={booking.time}>
					<span class="booking-date-main">
						{#if new Date(booking.time).getFullYear() !== now.getFullYear()}
							<span class="booking-year">{localYear(booking.time)}</span>
						{/if}
						<span class="booking-day-month">{localDayMonth(booking.time)}</span>
						<span class="booking-weekday">{localWeekday(booking.time)}</span>
					</span>
					<span class="booking-time">{localTime(booking.time)}</span>
				</time>
			</div>

			<div class="min-w-0">
				<h3 class="mb-4 mt-1 min-w-0 text-xl leading-none sm:mb-1 sm:mt-2" style="color: rgb(var(--color-text));">{booking.title}</h3>
				<div class="mt-3 grid gap-3">
					<p class="text-sm" style="color: rgb(var(--color-text) / 0.65);">
						<span class="block leading-4">Attendee:</span>
						<span class="ml-3 block" style="color: rgb(var(--color-text));">
							<span class="whitespace-nowrap">{booking.attendeeName} ·</span>
							<span class="whitespace-nowrap">{booking.attendeeEmail}</span>
						</span>
					</p>
					{#if booking.guestEmails.length > 0}
						<p class="text-sm" style="color: rgb(var(--color-text) / 0.65);">
							<span class="block leading-4">Guests:</span>
							<span class="ml-3 block" style="color: rgb(var(--color-text));">
								{#each booking.guestEmails as email (email)}
									<span class="block">{email}</span>
								{/each}
							</span>
						</p>
					{/if}
					<div>
						<span class="block text-sm leading-4" style="color: rgb(var(--color-text) / 0.65);">Hosts:</span>
						<div class="ml-3 mt-0.5"><HostBadges hosts={bookingHosts(booking)} {users} /></div>
					</div>
				</div>
			</div>

			{#if booking.manageURL}
				<div class="justify-self-end">
					<TimedActions
						label={`Show actions for ${booking.title}`}
						controlsId={`booking-actions-${booking.id}`}
						actionsVisibleOnSmallScreens
					>
						<a
							class="booking-action"
							href={booking.manageURL}
							aria-label={`Manage ${booking.title}`}
							title="Open booking page"
						>
							<Icon icon={externalLinkIcon} width="20" height="20" />
						</a>
					</TimedActions>
				</div>
			{/if}
		</article>
	{/each}
</div>

<style>
	.booking-row {
		background: rgb(var(--color-background));
		box-shadow: 0 0 0 1px var(--color-border);
		transition: color 0.15s ease;
	}

	.booking-action:hover {
		background: rgb(var(--color-primary));
		color: rgb(var(--color-background));
	}

	.booking-date-card {
		display: flex;
		overflow: hidden;
		border: 4px solid rgb(var(--color-text) / 0.15);
		border-radius: 0.5rem 0.5rem 0 0;
		color: rgb(var(--color-text));
	}

	.booking-date-main {
		display: flex;
		flex: 1;
		flex-direction: column;
		gap: 0.125rem;
		padding: 0.375rem 1rem 0.75rem;
		background: rgb(var(--color-text) / 0.15);
	}

	.booking-year {
		font-size: 0.75rem;
		line-height: 1.2;
	}

	.booking-day-month {
		font-size: 2rem;
		letter-spacing: -0.035em;
		line-height: 1.1;
		white-space: nowrap;
	}

	.booking-weekday {
		font-size: 0.75rem;
		line-height: 1.25;
	}

	.booking-time {
		display: flex;
		align-items: center;
		border-left: 1px solid var(--color-border);
		padding: 0.75rem 1rem;
		font-size: 0.875rem;
		font-weight: 600;
		line-height: 1;
	}

	@media (min-width: 40rem) {
		.booking-date-card {
			width: 8.5rem;
			flex-direction: column;
			border-radius: 0.75rem;
		}

		.booking-time {
			border-top: 1px solid var(--color-border);
			border-left: 0;
		}
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

	.booking-action:focus-visible {
		outline: 2px solid rgb(var(--color-text) / 0.65);
		outline-offset: 2px;
	}

	@media (prefers-reduced-motion: reduce) {
		.booking-action {
			transition: none;
		}
	}

</style>
