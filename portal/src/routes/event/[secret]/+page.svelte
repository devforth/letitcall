<script lang="ts">
	import { page } from '$app/state';
	import { onMount, tick } from 'svelte';
	import { appPath, callApi } from '$lib/api';
	import type { Booking, PublicEventType } from '$lib/types';
	import Button from '$lib/components/ui/Button.svelte';
	import ConfirmationDialog from '$lib/components/ui/ConfirmationDialog.svelte';
	import GuestEmailFields from '$lib/components/GuestEmailFields.svelte';
	import IconButton from '$lib/components/ui/IconButton.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import Textarea from '$lib/components/ui/Textarea.svelte';
	import BookingConfirmationView from '$lib/components/BookingConfirmationView.svelte';
	import EventTypeAside from '$lib/components/EventTypeAside.svelte';

	let booking = $state<Booking | null>(null);
	let eventType = $state<PublicEventType | null>(null);
	let inviteeLimit = $state<number | null>(null);
	let authenticated = $state(false);
	let loading = $state(true);
	let notFound = $state(false);
	let notes = $state('');
	let guestEmails = $state<string[]>([]);
	let guestLimit = $state<number | null>(null);
	let reason = $state('');
	let saving = $state(false);
	let canceling = $state(false);
	let showCancelDialog = $state(false);
	let view = $state<'summary' | 'edit' | 'cancel'>('summary');

	const secret = $derived(page.params.secret!);
	const blockStyle = 'background: rgb(var(--color-foreground)); box-shadow: var(--shadow-small);';
	const reasonPresets = ['Schedule conflict', 'No longer needed', 'Booked by mistake', 'Rescheduling'];
	// A preset owns the first line of the reason, so anything typed by hand survives
	// switching between chips.
	const activePreset = $derived(
		reasonPresets.find((preset) => reason.trim() === preset || reason.startsWith(`${preset}\n`)) ?? ''
	);
	const eventDateLabel = $derived.by(() =>
		booking
			? new Intl.DateTimeFormat(undefined, {
					timeZone: booking.attendeeTimezone,
					weekday: 'long',
					month: 'long',
					day: 'numeric',
					year: 'numeric'
				}).format(new Date(booking.time))
			: ''
	);
	const eventTimeLabel = $derived.by(() => {
		if (!booking) return '';
		const formatter = new Intl.DateTimeFormat(undefined, {
			timeZone: booking.attendeeTimezone,
			hour: 'numeric',
			minute: '2-digit'
		});
		return `${formatter.format(new Date(booking.time))} – ${formatter.format(new Date(booking.endTime))}`;
	});

	onMount(async () => {
		try {
			const response = await callApi<{ booking: Booking; inviteeLimit: number | null; guestLimit: number | null; authenticated: boolean }>(
				`/api/events/${encodeURIComponent(secret)}`
			);
			booking = response.booking;
			const eventTypeResponse = await callApi<{ eventType: PublicEventType }>(
				`/api/public/event-types/${encodeURIComponent(booking.eventSlug)}`
			);
			eventType = eventTypeResponse.eventType;
			inviteeLimit = response.inviteeLimit;
			guestLimit = response.guestLimit;
			authenticated = response.authenticated;
			notes = booking.notes ?? '';
			guestEmails = [...booking.guestEmails];
			syncViewWithHash();
		} catch (cause) {
			notFound = cause instanceof Error && cause.message === 'booking not found';
		} finally {
			loading = false;
		}
	});

	function localDate(value: string): string {
		return new Intl.DateTimeFormat(undefined, {
			dateStyle: 'full',
			timeStyle: 'short',
			timeZone: booking?.attendeeTimezone
		}).format(new Date(value));
	}

	function syncViewWithHash() {
		view = location.hash === '#event-details' ? 'edit' : location.hash === '#cancel-event' ? 'cancel' : 'summary';
	}

	function showSummary() {
		view = 'summary';
		history.replaceState(null, '', appPath(`/event/${encodeURIComponent(secret)}`));
	}

	async function save(event: SubmitEvent) {
		event.preventDefault();
		saving = true;
		try {
			const response = await callApi<{ booking: Booking }>(`/api/events/${encodeURIComponent(secret)}`, {
				method: 'PATCH',
				body: JSON.stringify({ notes, guestEmails })
			});
			booking = response.booking;
			notes = booking.notes ?? '';
			guestEmails = [...booking.guestEmails];
			showSummary();
		} catch {
			// callApi reports the error globally.
		} finally {
			saving = false;
		}
	}

	async function applyReasonPreset(preset: string) {
		// Focus before the value changes: the floating label rises on focus and on a filled
		// field, so writing the text first animates it twice over.
		const field = document.getElementById('cancellation-reason') as HTMLTextAreaElement | null;
		field?.focus();

		// A chip only ever writes its own text. Clicking the one already in the field leaves
		// it alone rather than emptying the reason.
		if (activePreset !== preset) {
			const typed = activePreset ? reason.slice(activePreset.length).replace(/^\n/, '') : reason.trim();
			reason = typed ? `${preset}\n${typed}` : preset;
		}

		// Leave the caret after the text so details can be typed straight away.
		await tick();
		field?.setSelectionRange(reason.length, reason.length);
	}

	async function cancelBooking() {
		canceling = true;
		try {
			const response = await callApi<{ booking: Booking }>(`/api/events/${encodeURIComponent(secret)}/cancel`, {
				method: 'POST',
				body: JSON.stringify({ reason })
			});
			booking = response.booking;
			showCancelDialog = false;
		} catch {
			// callApi reports the error globally.
		} finally {
			canceling = false;
		}
	}
</script>

<svelte:window onhashchange={syncViewWithHash} />

<PageTitle title={booking?.title ?? 'Event'} />

<!-- Shared by the edit and cancel views so both return the same way. -->
{#snippet backButton()}
	<div class="event-manage-back">
		<IconButton tone="primary" label="Back to booking" onclick={showSummary}>
			<!-- mingcute:left-line, drawn on a viewBox cropped to the glyph so the height in
			     CSS is the arrow itself rather than the icon's own margin. -->
			<svg class="event-back-chevron" viewBox="7.59 5.34 8.07 13.32" fill="currentColor" aria-hidden="true">
				<path d="M8.293 12.707a1 1 0 0 1 0-1.414l5.657-5.657a1 1 0 1 1 1.414 1.414L10.414 12l4.95 4.95a1 1 0 0 1-1.414 1.414l-5.657-5.657Z" />
			</svg>
		</IconButton>
	</div>
{/snippet}

{#if loading}
	<main class="grid min-h-screen place-items-center p-6"><p class="text-sm">Loading event…</p></main>
{:else if notFound || !booking || !eventType}
	<main class="grid min-h-screen place-items-center p-6">
		<section class="border border-black p-8 text-center">
			<h1 class="text-2xl font-semibold">Event not found</h1>
			<p class="mt-2 text-sm">This event link is not available.</p>
		</section>
	</main>
{:else}
	<main class="min-h-screen p-4 sm:p-8 lg:p-10">
		<div class="mx-auto grid min-h-[calc(100vh-5rem)] max-w-7xl overflow-hidden rounded-2xl lg:grid-cols-[21rem_1fr]" style={blockStyle}>
			<EventTypeAside
				{eventType}
				booking={view === 'edit' || view === 'cancel' ? booking : undefined}
				bookingDetails={view === 'edit' ? 'fixed' : 'all'}
			/>

			<section class="event-content p-6 lg:p-10" aria-label="Manage booking">
				{#if booking.canceledAt}
					<section class="event-status">
						<p class="event-status-label">Event canceled</p>
						<h2 class="event-page-title">This event has been canceled.</h2>
						<p class="event-page-copy">Canceled by {booking.canceledBy?.name} ({booking.canceledBy?.email}) on {localDate(booking.canceledAt)}.</p>
						{#if booking.cancellationReason}<p class="event-cancellation-reason">{booking.cancellationReason}</p>{/if}
					</section>
				{:else if view === 'summary'}
					<BookingConfirmationView
						title={booking.title}
						dateLabel={eventDateLabel}
						timeLabel={eventTimeLabel}
						timezone={booking.attendeeTimezone}
						attendeeName={booking.attendeeName}
						attendeeEmail={booking.attendeeEmail}
						guestEmails={booking.guestEmails}
						notes={booking.notes}
						attendeeLabel={authenticated ? 'Attendee' : 'Attendee (you)'}
						editHref={`${appPath(`/event/${encodeURIComponent(secret)}`)}#event-details`}
						cancelHref={`${appPath(`/event/${encodeURIComponent(secret)}`)}#cancel-event`}
						newBookingHref={appPath(`/book/${encodeURIComponent(booking.eventSlug)}`)}
						onedit={() => (view = 'edit')}
						oncancel={() => (view = 'cancel')}
					/>
				{:else if view === 'edit'}
					<header id="event-details" class="event-manage-header">
						<p class="event-manage-eyebrow">Manage booking</p>
						<div class="event-manage-title">
							{@render backButton()}
							<h2>Edit your event</h2>
						</div>
					</header>
					{#if authenticated}<p class="event-page-copy">You are signed in. Changes are recorded on your behalf.</p>{/if}

					<form class="event-form" autocomplete="off" onsubmit={save}>
						<div class="event-form-fields">
							<section>
								<h3 class="event-section-title">Guests</h3>
								<GuestEmailFields idPrefix="event-guest" bind:emails={guestEmails} limit={guestLimit} legend={null} />
								{#if inviteeLimit !== null}
									<p class="event-help">This time allows {inviteeLimit} invitees in total.</p>
								{/if}
							</section>
							<section>
								<h3 class="event-section-title">Notes</h3>
								<Textarea id="event-notes" label="Anything that will help prepare for this meeting" bind:value={notes} maxlength={2000} />
							</section>
						</div>
						<div class="event-form-actions">
							<Button type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save changes'}</Button>
						</div>
					</form>
				{:else}
					<section id="cancel-event" class="event-cancel">
						<div class="event-manage-header">
							<p class="event-manage-eyebrow">Manage booking</p>
							<div class="event-manage-title">
								{@render backButton()}
								<h2>Cancel event</h2>
							</div>
						</div>
						<div class="event-cancel-reason"><Textarea id="cancellation-reason" label="Cancellation reason (optional)" bind:value={reason} maxlength={2000} rows={8} resizable={false} /></div>
						<div class="event-cancel-presets" role="group" aria-label="Quick cancellation reasons">
							{#each reasonPresets as preset (preset)}
								<button type="button" class="event-preset-chip" onclick={() => applyReasonPreset(preset)}>{preset}</button>
							{/each}
						</div>
						<div class="event-cancel-footer">
							<p class="event-warning">
								<span>Canceling reopens the slot, emails everyone invited and removes the calendar entries.</span>
								<span>This cannot be undone.</span>
							</p>
							<div class="event-form-actions event-cancel-actions">
								<Button class="booking-next event-cancel-button gap-2" disabled={canceling} onclick={() => (showCancelDialog = true)}>
									Cancel event
								</Button>
							</div>
						</div>
					</section>
				{/if}
			</section>
		</div>
	</main>

	<ConfirmationDialog
		open={showCancelDialog}
		title="Cancel event?"
		description="This action cannot be undone."
		confirmLabel="Cancel event"
		confirmingLabel="Canceling…"
		confirming={canceling}
		onconfirm={cancelBooking}
		oncancel={() => (showCancelDialog = false)}
	/>
{/if}

<style>
	.event-content {
		display: flex;
		min-height: 0;
		flex-direction: column;
	}

	.event-status-label {
		font-size: 0.75rem;
		font-weight: 600;
		color: rgb(var(--color-primary));
	}

	.event-page-title {
		margin-top: 0.375rem;
		font-size: 1.5rem;
		font-weight: 600;
		line-height: 1.25;
	}

	.event-page-copy {
		margin-top: 0.5rem;
		font-size: 0.875rem;
		line-height: 1.5;
		color: rgb(var(--color-text) / 0.62);
	}

	.event-form {
		display: flex;
		flex: 1;
		min-height: 0;
		flex-direction: column;
		margin-top: 2rem;
	}

	.event-form-fields {
		display: grid;
		gap: 2rem;
		width: 100%;
	}

	.event-section-title {
		font-size: 1.125rem;
		font-weight: 500;
	}

	.event-form section {
		display: grid;
		gap: 0.75rem;
	}

	.event-help {
		font-size: 0.75rem;
		color: rgb(var(--color-text) / 0.6);
	}

	.event-form-actions {
		display: flex;
		justify-content: end;
		gap: 0.75rem;
		margin-top: auto;
		padding-top: 2rem;
	}

	.event-cancel {
		display: flex;
		flex: 1;
		min-height: 0;
		flex-direction: column;
		width: 100%;
	}

	/* The footer's gap already spaces this; .event-form-actions' padding is for the edit form. */
	.event-cancel-actions {
		flex: 0 0 auto;
		align-self: flex-end;
		padding-top: 0;
	}

	.event-manage-header {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		padding-bottom: 0.25rem;
	}

	/* Context label on its own line above the back button and current step. */
	.event-manage-eyebrow {
		margin: 0;
		font-size: 0.75rem;
		font-weight: 600;
		color: rgb(var(--color-text) / 0.65);
	}

	.event-manage-back {
		display: flex;
		justify-content: flex-start;
	}

	/* Matches the guest row's remove control in GuestEmailFields. */
	.event-manage-back :global(.icon-button) {
		width: 44px;
		height: 44px;
		border: 0;
		border-radius: 10px;
	}

	.event-manage-back :global(.tone-primary:not(:disabled)) {
		background: rgb(var(--color-primary) / 0.14);
		color: rgb(var(--color-primary));
	}

	/* The shared tone-primary hover matches its own resting tint, so lift it here. */
	.event-manage-back :global(.icon-button:hover:not(:disabled)) {
		background: rgb(var(--color-primary) / 0.22);
	}

	.event-back-chevron {
		display: block;
		height: 1rem;
		width: auto;
		margin-left: -3px;
	}

	/* Second line: back control beside the current step. */
	.event-manage-title {
		display: flex;
		align-items: center;
		gap: 0.875rem;
	}

	.event-manage-title h2 {
		margin: 0;
		font-size: 1.875rem;
		line-height: 2.25rem;
		font-weight: 600;
	}

	.event-cancel-presets {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		margin-top: 0.75rem;
	}

	.event-preset-chip {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-height: 1.75rem;
		border: 0;
		border-radius: 999px;
		background: rgb(var(--color-primary) / 0.12);
		padding: 0 0.875rem;
		font-size: 0.8125rem;
		font-weight: 400;
		line-height: 1;
		color: rgb(var(--color-primary));
		cursor: pointer;
		transition: background 0.2s ease, color 0.2s ease;
	}

	.event-preset-chip:hover {
		background: rgb(var(--color-primary) / 0.2);
	}

	.event-cancel-reason {
		margin-top: 1.5rem;
	}

	.event-cancel-footer {
		display: flex;
		align-items: stretch;
		flex-direction: column;
		gap: 1rem;
		margin-top: auto;
		padding-top: 1.25rem;
	}

	.event-cancel :global(.event-cancel-button) {
		border-color: rgb(var(--error)) !important;
		background: rgb(var(--error)) !important;
		color: rgb(var(--color-contrast-text)) !important;
	}

	.event-cancel :global(.event-cancel-button svg) {
		color: currentColor;
	}

	.event-cancel :global(.event-cancel-button:hover:not(:disabled)) {
		border-color: color-mix(in srgb, rgb(var(--error)), white 15%) !important;
		background: color-mix(in srgb, rgb(var(--error)), white 15%) !important;
	}

	.event-cancel :global(.event-cancel-button:active:not(:disabled)) {
		border-color: color-mix(in srgb, rgb(var(--error)), black 15%) !important;
		background: color-mix(in srgb, rgb(var(--error)), black 8%) !important;
		box-shadow: none;
	}

	.event-warning {
		display: flex;
		align-items: flex-start;
		justify-content: center;
		flex-direction: column;
		gap: 0.125rem;
		width: 100%;
		min-height: 44px;
		border-left: 4px solid rgb(var(--warning));
		background: rgb(var(--warning) / 0.08);
		padding: 0.5rem 0.875rem;
		color: rgb(var(--warning));
		font-size: 0.8125rem;
		font-weight: 600;
		line-height: 1.25;
	}

	.event-cancellation-reason {
		margin-top: 1.5rem;
		border-left: 3px solid rgb(var(--color-primary) / 0.4);
		padding: 0.125rem 0 0.125rem 0.875rem;
		font-size: 0.875rem;
		line-height: 1.5;
		white-space: pre-wrap;
	}

	@media (max-width: 640px) {
		.event-cancel-actions {
			align-self: stretch;
		}

		.event-form-actions {
			align-items: stretch;
			flex-direction: column;
		}

		.event-form-actions :global(button) {
			width: 100%;
		}
	}
</style>
