<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import calendarEventIcon from '@iconify-icons/tabler/calendar-event';
	import clockIcon from '@iconify-icons/tabler/clock';
	import usersIcon from '@iconify-icons/tabler/users';
	import { callApi } from '$lib/api';
	import type { Booking } from '$lib/types';
	import Button from '$lib/components/ui/Button.svelte';
	import ConfirmationDialog from '$lib/components/ui/ConfirmationDialog.svelte';
	import GuestEmailFields from '$lib/components/GuestEmailFields.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import { branding } from '$lib/stores/branding.svelte';
	import Textarea from '$lib/components/ui/Textarea.svelte';

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
	<main class="min-h-screen p-4 sm:p-8">
		<div class="mx-auto max-w-4xl border border-black">
			<header class="border-b border-black p-6 sm:p-8">
				<p class="text-sm font-semibold">{branding.name}</p>
				<h1 class="mt-5 text-3xl font-semibold tracking-tight">{booking.title}</h1>
				{#if authenticated}<p class="mt-3 text-sm">You are signed in. Changes are recorded on your behalf.</p>{/if}
			</header>

			<section class="grid gap-5 border-b border-black p-6 sm:grid-cols-2 sm:p-8">
				<div class="flex items-start gap-3">
					<Icon icon={calendarEventIcon} width="22" height="22" class="mt-0.5 shrink-0" />
					<div><p class="text-xs font-semibold uppercase">Invitee</p><p class="mt-1">{booking.attendeeName}</p><p class="text-sm">{booking.attendeeEmail}</p></div>
				</div>
				<div class="flex items-start gap-3">
					<Icon icon={clockIcon} width="22" height="22" class="mt-0.5 shrink-0" />
					<div><p class="text-xs font-semibold uppercase">Date and time</p><p class="mt-1">{localDate(booking.time)}</p><p class="text-sm">{booking.attendeeTimezone}</p></div>
				</div>
			</section>

			{#if booking.canceledAt}
				<section class="p-6 sm:p-8">
					<h2 class="text-2xl font-semibold">Canceled</h2>
					<p class="mt-3 text-sm">Canceled by {booking.canceledBy?.name} ({booking.canceledBy?.email}) on {localDate(booking.canceledAt)}.</p>
					{#if booking.cancellationReason}<p class="mt-4 border border-black p-4">{booking.cancellationReason}</p>{/if}
				</section>
			{:else}
				<form class="grid gap-6 border-b border-black p-6 sm:p-8" onsubmit={save}>
					<div class="flex items-center gap-2"><Icon icon={usersIcon} width="22" height="22" /><h2 class="text-xl font-semibold">Event details</h2></div>
					<Textarea id="event-notes" label="Description" bind:value={notes} maxlength={2000} />
					<GuestEmailFields idPrefix="event-guest" bind:emails={guestEmails} limit={guestLimit} />
					<p class="text-xs">{inviteeLimit === null ? 'There is no guest limit.' : `This time allows ${inviteeLimit} invitees in total.`}</p>
					<div><Button type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save changes'}</Button></div>
				</form>

				<section class="p-6 sm:p-8">
					<h2 class="text-xl font-semibold">Cancel event</h2>
					<p class="mt-2 text-sm">Enter a reason before confirming cancellation. The reason is optional.</p>
					<div class="mt-5"><Textarea id="cancellation-reason" label="Cancellation reason" bind:value={reason} maxlength={2000} /></div>
					<div class="mt-5"><Button variant="secondary" disabled={canceling} onclick={() => (showCancelDialog = true)}>Cancel event</Button></div>
				</section>
			{/if}
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
