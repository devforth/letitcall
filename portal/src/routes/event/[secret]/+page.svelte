<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import clockIcon from '@iconify-icons/tabler/clock';
	import notesIcon from '@iconify-icons/tabler/align-left';
	import userIcon from '@iconify-icons/tabler/user';
	import usersIcon from '@iconify-icons/tabler/users';
	import { callApi } from '$lib/api';
	import type { Booking } from '$lib/types';
	import Button from '$lib/components/ui/Button.svelte';
	import ConfirmationDialog from '$lib/components/ui/ConfirmationDialog.svelte';
	import GuestEmailFields from '$lib/components/GuestEmailFields.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import { branding } from '$lib/stores/branding.svelte';
	import BrandLogo from '$lib/components/BrandLogo.svelte';
	import Textarea from '$lib/components/ui/Textarea.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';

	let booking = $state<Booking | null>(null);
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

	const secret = $derived(page.params.secret!);
	const blockStyle = 'background: rgb(var(--color-foreground)); box-shadow: var(--shadow-small);';
	const asideStyle =
		'background: rgb(var(--color-primary)); color: rgb(var(--color-contrast-text)); box-shadow: 0 0 0 1px rgb(var(--color-border)), var(--shadow-small);';

	onMount(async () => {
		try {
			const response = await callApi<{ booking: Booking; inviteeLimit: number | null; guestLimit: number | null; authenticated: boolean }>(
				`/api/events/${encodeURIComponent(secret)}`
			);
			booking = response.booking;
			inviteeLimit = response.inviteeLimit;
			guestLimit = response.guestLimit;
			authenticated = response.authenticated;
			notes = booking.notes ?? '';
			guestEmails = [...booking.guestEmails];
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
		} catch {
			// callApi reports the error globally.
		} finally {
			saving = false;
		}
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

<PageTitle title={booking?.title ?? 'Event'} />

{#if loading}
	<main class="grid min-h-screen place-items-center p-6"><p class="text-sm">Loading event…</p></main>
{:else if notFound || !booking}
	<main class="grid min-h-screen place-items-center p-6">
		<section class="border border-black p-8 text-center">
			<h1 class="text-2xl font-semibold">Event not found</h1>
			<p class="mt-2 text-sm">This event link is not available.</p>
		</section>
	</main>
{:else}
	<main class="min-h-screen p-4 sm:p-8 lg:p-10">
		<div class="mx-auto grid min-h-[calc(100vh-5rem)] max-w-7xl overflow-hidden rounded-2xl lg:grid-cols-[21rem_1fr]" style={blockStyle}>
			<aside class="relative flex flex-col rounded-b-2xl p-6 lg:rounded-bl-none lg:rounded-tr-2xl lg:rounded-br-2xl lg:p-8" style={asideStyle}>
				<h1 class="text-3xl font-semibold tracking-tight">{booking.title}</h1>
				<div class="event-aside-details">
					<section class="event-aside-detail">
						<Icon icon={clockIcon} width="22" height="22" class="mt-0.5 shrink-0" />
						<div>
							<p class="event-aside-label">Date and time</p>
							<p>{localDate(booking.time)}</p>
							<p class="event-aside-muted">{booking.attendeeTimezone}</p>
						</div>
					</section>
					<section class="event-aside-detail">
						<Icon icon={userIcon} width="22" height="22" class="mt-0.5 shrink-0" />
						<div>
							<p class="event-aside-label">Attendee</p>
							<p>{booking.attendeeName}</p>
							<p class="event-aside-muted">{booking.attendeeEmail}</p>
						</div>
					</section>
					{#if booking.guestEmails.length > 0}
						<section class="event-aside-detail">
							<Icon icon={usersIcon} width="22" height="22" class="mt-0.5 shrink-0" />
							<div>
								<p class="event-aside-label">Guests · {booking.guestEmails.length}</p>
								<ul class="event-aside-list">
									{#each booking.guestEmails as email (email)}
										<li>{email}</li>
									{/each}
								</ul>
							</div>
						</section>
					{/if}
					{#if booking.notes}
						<section class="event-aside-detail">
							<Icon icon={notesIcon} width="22" height="22" class="mt-0.5 shrink-0" />
							<div>
								<p class="event-aside-label">Notes</p>
								<p class="event-aside-notes">{booking.notes}</p>
							</div>
						</section>
					{/if}
				</div>
				<div class="mt-auto flex items-center gap-3 pt-12 text-sm font-semibold">
					<BrandLogo class="size-10 rounded-xl object-cover" />
					<span>{branding.name}</span>
				</div>
				<div class="event-theme-toggle theme-toggle-contrast absolute bottom-6 right-6 lg:bottom-8 lg:right-8">
					<ThemeToggle />
				</div>
			</aside>

			<section class="p-6 lg:p-10" aria-label="Manage booking">
				{#if booking.canceledAt}
					<section class="event-status">
						<p class="event-status-label">Event canceled</p>
						<h2 class="event-page-title">This event has been canceled.</h2>
						<p class="event-page-copy">Canceled by {booking.canceledBy?.name} ({booking.canceledBy?.email}) on {localDate(booking.canceledAt)}.</p>
						{#if booking.cancellationReason}<p class="event-cancellation-reason">{booking.cancellationReason}</p>{/if}
					</section>
				{:else}
					<header>
						<p class="event-page-label">Manage booking</p>
						<h2 class="event-page-title">Edit your event</h2>
						{#if authenticated}<p class="event-page-copy">You are signed in. Changes are recorded on your behalf.</p>{/if}
					</header>

					<form id="event-details" class="event-form" autocomplete="off" onsubmit={save}>
						<section>
							<h3 class="event-section-title">Guests</h3>
							<GuestEmailFields idPrefix="event-guest" bind:emails={guestEmails} limit={guestLimit} />
							<p class="event-help">{inviteeLimit === null ? 'There is no guest limit.' : `This time allows ${inviteeLimit} invitees in total.`}</p>
						</section>
						<section>
							<h3 class="event-section-title">Notes</h3>
							<Textarea id="event-notes" label="Anything that will help prepare for this meeting" bind:value={notes} maxlength={2000} />
						</section>
						<div class="event-form-actions"><Button type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save changes'}</Button></div>
					</form>

					<section id="cancel-event" class="event-cancel">
						<h2 class="event-section-title">Cancel event</h2>
						<p class="event-page-copy">This action cannot be undone. A cancellation reason is optional.</p>
						<div class="mt-5"><Textarea id="cancellation-reason" label="Cancellation reason" bind:value={reason} maxlength={2000} /></div>
						<div class="mt-5"><Button variant="secondary" disabled={canceling} onclick={() => (showCancelDialog = true)}>Cancel event</Button></div>
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
	.event-aside-details {
		display: grid;
		gap: 1.25rem;
		margin-top: 2.5rem;
	}

	.event-aside-detail {
		display: flex;
		align-items: start;
		gap: 0.75rem;
		font-size: 0.875rem;
		line-height: 1.5rem;
	}

	.event-aside-label {
		font-size: 0.75rem;
		font-weight: 600;
		color: rgb(var(--color-contrast-text) / 0.65);
	}

	.event-aside-muted {
		color: rgb(var(--color-contrast-text) / 0.75);
	}

	.event-aside-list {
		display: grid;
		gap: 0.125rem;
		margin: 0;
		padding: 0;
		list-style: none;
		overflow-wrap: anywhere;
	}

	.event-aside-notes {
		white-space: pre-wrap;
	}

	.event-page-label,
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
		display: grid;
		gap: 2rem;
		max-width: 40rem;
		margin-top: 2rem;
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
	}

	.event-cancel {
		max-width: 40rem;
		margin-top: 3rem;
		border-top: 1px solid rgb(var(--color-border));
		padding-top: 2rem;
	}

	.event-cancellation-reason {
		margin-top: 1.5rem;
		border-left: 3px solid rgb(var(--color-primary) / 0.4);
		padding: 0.125rem 0 0.125rem 0.875rem;
		font-size: 0.875rem;
		line-height: 1.5;
		white-space: pre-wrap;
	}

	.theme-toggle-contrast :global(.toggle-switch) {
		background: transparent !important;
		border-color: rgb(var(--color-contrast-text)) !important;
		color: rgb(var(--color-contrast-text)) !important;
		box-shadow: none !important;
	}

	@media (max-width: 640px) {
		.event-form-actions {
			justify-content: stretch;
		}

		.event-form-actions :global(button) {
			width: 100%;
		}
	}
</style>
