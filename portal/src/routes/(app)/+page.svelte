<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import calendarEventIcon from '@iconify-icons/tabler/calendar-event';
	import clockIcon from '@iconify-icons/tabler/clock';
	import historyIcon from '@iconify-icons/tabler/history';
	import { callApi } from '$lib/api';
	import type { Booking, EventType, ManagedUser } from '$lib/types';
	import BookingList from '$lib/components/BookingList.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';

	let bookings = $state<Booking[]>([]);
	let eventTypes = $state<EventType[]>([]);
	let users = $state<ManagedUser[]>([]);
	let loading = $state(true);
	let now = $state(new Date());
	let clock: number | undefined;

	const blockStyle = 'background: rgb(var(--color-foreground)); box-shadow: var(--shadow-small);';

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

</script>

<PageTitle title="Bookings" />

<section aria-labelledby="bookings-title" class="flex flex-col gap-4">
	<div class="rounded-lg p-4 sm:p-5" style={blockStyle}>
		<div class="flex min-w-0 items-center gap-4">
			<div
				class="grid size-12 shrink-0 place-items-center rounded-lg"
				style="background: rgb(var(--color-primary) / 0.12); color: rgb(var(--color-primary));"
			>
				<Icon icon={calendarEventIcon} width="24" height="24" />
			</div>
			<div>
				<div class="flex items-center gap-3">
					<h1 id="bookings-title" class="text-2xl font-semibold tracking-tight" style="color: rgb(var(--color-text));">Bookings</h1>
					<span class="inline-flex items-center rounded-md px-2 py-1 text-xs font-semibold" style="background: rgb(var(--color-primary) / 0.1); color: rgb(var(--color-primary));">
						{loading ? 'Loading…' : `${bookings.length} ${bookings.length === 1 ? 'booking' : 'bookings'}`}
					</span>
				</div>
				<p class="text-sm" style="color: rgb(var(--color-text) / 0.65);">Upcoming appointments and booking history.</p>
			</div>
		</div>
	</div>

	{#if loading}
		<div class="rounded-lg" style={blockStyle}>
			<p class="p-8 text-sm" style="color: rgb(var(--color-text) / 0.65);">Loading bookings…</p>
		</div>
	{:else if bookings.length === 0}
		<div class="rounded-lg" style={blockStyle}>
			<div class="px-5 py-14 text-center">
				<div class="mx-auto flex max-w-xs flex-col items-center">
					<Icon icon={calendarEventIcon} width="30" height="30" style="color: rgb(var(--color-text) / 0.65);" />
					<p class="mt-3 font-semibold" style="color: rgb(var(--color-text));">No bookings yet</p>
					<p class="mt-1 text-xs" style="color: rgb(var(--color-text) / 0.65);">New appointments will appear here.</p>
				</div>
			</div>
		</div>
	{:else}
		<section class="booking-group overflow-hidden rounded-lg" style={blockStyle} aria-labelledby="upcoming-title">
			<div class="booking-group-heading">
				<Icon icon={clockIcon} width="18" height="18" />
				<h2 id="upcoming-title" class="text-sm font-semibold">Upcoming</h2>
				<span class="group-count">{upcoming.length}</span>
			</div>
			{#if upcoming.length > 0}
				<BookingList bookings={upcoming} {eventTypes} {users} {now} />
			{:else}
				<p class="p-5 text-sm" style="color: rgb(var(--color-text) / 0.65);">No upcoming bookings.</p>
			{/if}
		</section>

		{#if history.length > 0}
			<section class="booking-group overflow-hidden rounded-lg" style={blockStyle} aria-labelledby="history-title">
				<div class="booking-group-heading">
					<Icon icon={historyIcon} width="18" height="18" />
					<h2 id="history-title" class="text-sm font-semibold">Booking history</h2>
					<span class="group-count">{history.length}</span>
				</div>
				<BookingList bookings={history} {eventTypes} {users} {now} historical />
			</section>
		{/if}
	{/if}
</section>

<style>
	.booking-group-heading {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		border-bottom: 1px solid rgb(var(--color-border));
		padding: 0.75rem 1rem;
		background: rgb(var(--color-text) / 0.06);
		color: rgb(var(--color-text));
	}

	.group-count {
		color: rgb(var(--color-text) / 0.45);
		font-size: 0.75rem;
		font-variant-numeric: tabular-nums;
	}
</style>
