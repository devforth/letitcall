<script lang="ts">
	import Icon from '@iconify/svelte';
	import pencilIcon from '@iconify-icons/mdi/edit';
	import xIcon from '@iconify-icons/mingcute/close-fill';
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
		celebrate = false,
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
		celebrate?: boolean;
		reloadNewBooking?: boolean;
	} = $props();

	const bookingEventActionStyle =
		'height: 2.75rem !important; min-height: 2.75rem !important; border-radius: 9999px !important;';
</script>

<section class="booking-confirmed" aria-labelledby="booking-confirmed-title">
	{#if celebrate}
		<div class="booking-celebration" aria-hidden="true">
			{#each Array(18) as _, index}
				<span
					style={`--x: ${4 + index * 5.4}%; --drift: ${(index % 5) - 2}rem; --fall: ${7 + (index % 4) * 1.25}rem; --turn: ${180 + (index % 7) * 55}deg; --delay: ${(index % 6) * 0.045}s;`}
				></span>
			{/each}
		</div>
	{/if}
	<header class="booking-confirmed-header">
		<div>
			<BookingStatusHeading text="Booking confirmed" />
			<BookingSubtitle id="booking-confirmed-title" text={`You’re booked for ${title}`} />
		</div>
	</header>
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
			<a
				class="booking-event-action button-primary-outline outlined-action-button"
				style={`${bookingEventActionStyle} border: 2px solid rgb(var(--color-primary)) !important; box-shadow: none !important;`}
				href={cancelHref}
				onclick={oncancel}
			><Icon icon={xIcon} width="22" height="22" />Cancel event</a>
			<a
				class="booking-event-action button-primary primary-action-button"
				style={bookingEventActionStyle}
				href={editHref}
				onclick={onedit}
			><Icon icon={pencilIcon} width="22" height="22" />Edit event</a>
		</div>
		{#if reloadNewBooking}
			<a class="booking-new-link" href={newBookingHref} data-sveltekit-reload>Make another booking</a>
		{:else}
			<a class="booking-new-link" href={newBookingHref}>Make another booking</a>
		{/if}
	</div>
</section>

<style>
	.booking-confirmed {
		position: relative;
		display: flex;
		flex: 1;
		min-height: 0;
		flex-direction: column;
		isolation: isolate;
	}

	.booking-confirmed > :not(.booking-celebration) {
		position: relative;
		z-index: 1;
	}

	.booking-confirmed-header {
		margin-top: -0.5rem;
		padding: 0 0 1.25rem;
	}

	.booking-celebration {
		position: absolute;
		z-index: 0;
		top: -2rem;
		right: -1rem;
		left: -1rem;
		height: 13rem;
		overflow: hidden;
		pointer-events: none;
	}

	.booking-celebration span {
		position: absolute;
		top: 0;
		left: var(--x);
		width: 0.35rem;
		height: 0.75rem;
		border: 1px solid rgb(var(--color-primary));
		background: rgb(var(--color-primary));
		opacity: 0;
		animation: booking-hooray 1.2s cubic-bezier(0.18, 0.72, 0.28, 1) var(--delay) both;
	}

	.booking-celebration span:nth-child(even) {
		background: rgb(var(--color-text));
	}

	.booking-celebration span:nth-child(3n) {
		width: 0.45rem;
		height: 0.45rem;
		border-radius: 999px;
		background: transparent;
	}

	@keyframes booking-hooray {
		0% {
			opacity: 0;
			transform: translate3d(0, -0.75rem, 0) rotate(0deg) scale(0.6);
		}
		12% {
			opacity: 1;
		}
		100% {
			opacity: 0;
			transform: translate3d(var(--drift), var(--fall), 0) rotate(var(--turn)) scale(1);
		}
	}

	.booking-confirmed-actions {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		margin-top: auto;
		padding-top: 1.5rem;
	}

	.booking-event-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
		order: 2;
	}

	.booking-event-action {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		font-size: 0.875rem;
		font-weight: 600;
		transition: background 0.2s ease, color 0.2s ease;
	}

	.booking-new-link {
		order: 1;
		font-size: 1rem;
		font-weight: 500;
		color: rgb(var(--color-primary));
	}

	.booking-new-link:hover {
		text-decoration: underline;
		text-underline-offset: 0.25rem;
	}

	@media (prefers-reduced-motion: reduce) {
		.booking-celebration {
			display: none;
		}
	}

	@media (max-width: 640px) {
		.booking-confirmed-actions {
			align-items: stretch;
			flex-direction: column;
		}

		.booking-event-actions {
			flex-direction: column;
		}

		/* Stacked full-width buttons above it, so the link centres under them. */
		.booking-new-link {
			text-align: center;
		}
	}
</style>
