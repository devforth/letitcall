<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import calendarEventIcon from '@iconify-icons/tabler/calendar-event';
	import clockIcon from '@iconify-icons/tabler/clock';
	import historyIcon from '@iconify-icons/tabler/history';
	import externalLinkIcon from '@iconify-icons/tabler/external-link';
	import { callApi } from '$lib/api';
	import type { Booking, EventType, ManagedUser } from '$lib/types';
	import HostBadges from '$lib/components/HostBadges.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';

	let bookings = $state<Booking[]>([]);
	let eventTypes = $state<EventType[]>([]);
	let users = $state<ManagedUser[]>([]);
	let loading = $state(true);
	let now = $state(new Date());
	let clock: number | undefined;

	const upcoming = $derived(
		bookings
			.filter((booking) => !booking.canceledAt && new Date(booking.time) >= now)
			.sort((left, right) => left.time.localeCompare(right.time))
	);
	const history = $derived(
		bookings
			.filter((booking) => Boolean(booking.canceledAt) || new Date(booking.time) < now)
			.sort((left, right) => right.time.localeCompare(left.time))
	);

	onMount(async () => {
		clock = window.setInterval(() => (now = new Date()), 60_000);
		try {
			const [bookingsResponse, eventTypesResponse, usersResponse] = await Promise.all([
				callApi<{ bookings: Booking[] }>('/api/bookings'),
				callApi<{ eventTypes: EventType[] }>('/api/event-types'),
				callApi<{ users: ManagedUser[] }>('/api/users')
			]);
			bookings = bookingsResponse.bookings;
			eventTypes = eventTypesResponse.eventTypes;
			users = usersResponse.users;
		} catch {
			// callApi reports the error globally.
		} finally {
			loading = false;
		}
	});

	onDestroy(() => {
		if (clock !== undefined) window.clearInterval(clock);
	});

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

<PageTitle title="Bookings" />

<section aria-labelledby="bookings-title">
	<div>
		<h1 id="bookings-title" class="text-2xl font-semibold tracking-tight">Bookings</h1>
		<p class="mt-2 text-sm">Upcoming appointments and booking history.</p>
	</div>

	{#if loading}
		<p class="mt-6 border border-black p-6 text-sm">Loading bookings…</p>
	{:else if bookings.length === 0}
		<div class="grid min-h-[60vh] place-items-center text-center">
			<div>
				<Icon icon={calendarEventIcon} width="36" height="36" class="mx-auto" />
				<p class="mt-4 font-medium">No booking yet</p>
			</div>
		</div>
	{:else}
		<section class="mt-8" aria-labelledby="upcoming-title">
			<div class="flex items-center gap-2 border-b border-black pb-3">
				<Icon icon={clockIcon} width="20" height="20" />
				<h2 id="upcoming-title" class="text-lg font-semibold">Upcoming</h2>
			</div>
			<div class="mt-4 grid gap-3">
				{#each upcoming as booking (booking.id)}
					<article class="grid gap-4 border border-black p-4 sm:grid-cols-[1fr_auto] sm:items-center">
						<div>
							<h3 class="font-semibold">{booking.title}</h3>
							<p class="mt-1 text-sm">{booking.attendeeName} · {booking.attendeeEmail}</p>
							<p class="mt-2 text-xs">{localDate(booking.time)}</p>
							<div class="mt-3"><HostBadges hosts={bookingHosts(booking)} {users} /></div>
						</div>
						<div class="flex items-center gap-2">
							<p class="w-fit border border-black px-3 py-2 text-sm font-medium">{relativeTime(booking.time)}</p>
							{#if booking.manageURL}
								<a class="grid size-10 place-items-center border border-black hover:bg-black hover:text-white" href={booking.manageURL} aria-label={`Manage ${booking.title}`} title="Open booking page">
									<Icon icon={externalLinkIcon} width="18" height="18" />
								</a>
							{/if}
						</div>
					</article>
				{:else}
					<p class="border border-black p-6 text-sm">No upcoming bookings.</p>
				{/each}
			</div>
		</section>

		{#if history.length > 0}
			<section class="mt-12" aria-labelledby="history-title">
				<div class="flex items-center gap-2 border-b border-black pb-3">
					<Icon icon={historyIcon} width="20" height="20" />
					<h2 id="history-title" class="text-lg font-semibold">Booking history</h2>
				</div>
				<div class="mt-4 grid gap-3">
					{#each history as booking (booking.id)}
						<article class="grid gap-3 border border-black p-4 opacity-60 sm:grid-cols-[1fr_auto] sm:items-center">
							<div>
								<h3 class="font-semibold">{booking.title}</h3>
								<p class="mt-1 text-sm">{booking.attendeeName} · {booking.attendeeEmail}</p>
								<p class="mt-2 text-xs">{localDate(booking.time)}</p>
								<div class="mt-3"><HostBadges hosts={bookingHosts(booking)} {users} /></div>
							</div>
							<div class="flex items-center gap-3">
								<p class="text-sm">{booking.canceledAt ? 'Canceled' : relativeTime(booking.time)}</p>
								{#if booking.manageURL}
									<a class="grid size-10 place-items-center border border-black hover:bg-black hover:text-white" href={booking.manageURL} aria-label={`Manage ${booking.title}`} title="Open booking page">
										<Icon icon={externalLinkIcon} width="18" height="18" />
									</a>
								{/if}
							</div>
						</article>
					{/each}
				</div>
			</section>
		{/if}
	{/if}
</section>
