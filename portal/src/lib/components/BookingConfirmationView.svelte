<script lang="ts">
	import Icon from '@iconify/svelte';
	import pencilIcon from '@iconify-icons/tabler/pencil';
	import xIcon from '@iconify-icons/tabler/x';
	import BookingDetailsCard from '$lib/components/BookingDetailsCard.svelte';
	import BookingStatusHeading from '$lib/components/BookingStatusHeading.svelte';
	import BookingSubtitle from '$lib/components/BookingSubtitle.svelte';

	let {
		title,
		dateLabel,
		timeLabel,
		timezone,
		attendeeName,
		attendeeEmail,
		guestEmails,
		notes,
		attendeeLabel = 'Attendee (you)',
		editHref,
		cancelHref,
		newBookingHref,
		onedit,
		oncancel,
		reloadNewBooking = false
	}: {
		title: string;
		dateLabel: string;
		timeLabel: string;
		timezone: string;
		attendeeName: string;
		attendeeEmail: string;
		guestEmails: string[];
		notes?: string;
		attendeeLabel?: string;
		editHref: string;
		cancelHref: string;
		newBookingHref: string;
		onedit?: (event: MouseEvent) => void;
		oncancel?: (event: MouseEvent) => void;
		reloadNewBooking?: boolean;
	} = $props();
</script>

<section class="booking-confirmed" aria-labelledby="booking-confirmed-title">
	<BookingStatusHeading text="Booking confirmed" />
	<BookingSubtitle id="booking-confirmed-title" text={`You’re booked for ${title}`} />
	<BookingDetailsCard
		{dateLabel}
		{timeLabel}
		{timezone}
		{attendeeName}
		{attendeeEmail}
		{guestEmails}
		{notes}
		{attendeeLabel}
	/>
	<div class="booking-confirmed-actions">
		<div class="booking-event-actions">
			<a class="booking-event-action booking-edit-action" href={editHref} onclick={onedit}><Icon icon={pencilIcon} width="18" height="18" />Edit event</a>
			<a class="booking-event-action booking-cancel-action" href={cancelHref} onclick={oncancel}><Icon icon={xIcon} width="18" height="18" />Cancel event</a>
		</div>
		{#if reloadNewBooking}
			<a class="booking-new-link" href={newBookingHref} data-sveltekit-reload>Make another booking</a>
		{:else}
			<a class="booking-new-link" href={newBookingHref}>Make another booking</a>
		{/if}
	</div>
</section>

<style>
	.booking-confirmed-actions {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		margin-top: 1.5rem;
	}

	.booking-event-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
	}

	.booking-event-action {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		min-height: 3rem;
		border-radius: 11px;
		padding: 0.5rem 1rem;
		font-size: 0.875rem;
		font-weight: 600;
		transition: background 0.2s ease, color 0.2s ease;
	}

	.booking-edit-action {
		background: rgb(var(--color-primary) / 0.14);
		color: rgb(var(--color-primary));
	}

	.booking-edit-action:hover {
		background: rgb(var(--color-primary) / 0.2);
	}

	.booking-cancel-action {
		background: rgb(var(--error) / 0.14);
		color: rgb(var(--error));
	}

	.booking-cancel-action:hover {
		background: rgb(var(--error) / 0.2);
	}

	.booking-new-link {
		font-size: 0.875rem;
		font-weight: 600;
		color: rgb(var(--color-primary));
		text-decoration: underline;
	}

	.booking-new-link:hover {
		text-decoration: none;
	}

	@media (max-width: 640px) {
		.booking-confirmed-actions {
			align-items: stretch;
			flex-direction: column;
		}

		.booking-event-actions {
			flex-direction: column;
		}
	}
</style>
