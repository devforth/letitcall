<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import calendarEventIcon from '@iconify-icons/tabler/calendar-event';
	import clockIcon from '@iconify-icons/tabler/clock';
	import historyIcon from '@iconify-icons/tabler/history';
	import { callApi } from '$lib/api';
	import type { Booking, EventType, ManagedUser } from '$lib/types';
	import BookingList from '$lib/components/BookingList.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import SegmentedControl from '$lib/components/ui/SegmentedControl.svelte';

	let bookings = $state<Booking[]>([]);
	let eventTypes = $state<EventType[]>([]);
	let users = $state<ManagedUser[]>([]);
	let loading = $state(true);
	let now = $state(new Date());
	let bookingView = $state('upcoming');
	let clock: number | undefined;

	const blockStyle =
		'background: rgb(var(--color-background)); box-shadow: 0 0 0 1px var(--color-border);';

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
	const bookingViews = $derived([
		{ value: 'upcoming', label: 'Upcoming', icon: clockIcon, suffix: `· ${upcoming.length}` },
		{ value: 'history', label: 'History', icon: historyIcon, suffix: `· ${history.length}` }
	]);

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

<section aria-labelledby="bookings-title" class="flex flex-col gap-6">
	<div class="mb-2">
		<PageHeader
			id="bookings-title"
			title="Bookings"
			description="Upcoming appointments and booking history."
			icon={calendarEventIcon}
		>
			<SegmentedControl
				options={bookingViews}
				value={bookingView}
				label="Booking view"
				onchange={(value) => (bookingView = value)}
			/>
		</PageHeader>
	</div>

	{#if loading}
		<div class="rounded-lg" style={blockStyle}>
			<p class="p-8 text-sm" style="color: rgb(var(--color-text) / 0.65);">Loading bookings…</p>
		</div>
	{:else if bookings.length === 0}
		<div class="rounded-lg" style={blockStyle}>
			<p class="empty-state">No bookings yet</p>
		</div>
	{:else}
		{#if bookingView === 'upcoming'}
			<section aria-labelledby="upcoming-title">
				<h2 id="upcoming-title" class="sr-only">Upcoming</h2>
				{#if upcoming.length > 0}
					<BookingList bookings={upcoming} {eventTypes} {users} {now} />
				{:else}
					<p class="empty-state rounded-lg" style={blockStyle}>No upcoming bookings</p>
				{/if}
			</section>
		{:else}
			<section aria-labelledby="history-title">
				<h2 id="history-title" class="sr-only">Booking history</h2>
				{#if history.length > 0}
					<BookingList bookings={history} {eventTypes} {users} {now} />
				{:else}
					<p class="empty-state rounded-lg" style={blockStyle}>No booking history</p>
				{/if}
			</section>
		{/if}
	{/if}
</section>

<style>
	.empty-state {
		padding: 2.5rem 1rem;
		color: rgb(var(--color-text) / 0.65);
		font-size: 0.875rem;
		text-align: center;
	}
</style>
