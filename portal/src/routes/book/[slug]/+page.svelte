<script lang="ts">
	import { page } from '$app/state';
	import { onDestroy, onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import clockIcon from '@iconify-icons/tabler/clock';
	import arrowLeftIcon from '@iconify-icons/tabler/arrow-left';
	import calendarIcon from '@iconify-icons/tabler/calendar';
	import worldIcon from '@iconify-icons/tabler/world';
	import { avatarURL, callApi } from '$lib/api';
	import { generateBookingSlots, timezoneDateKey } from '$lib/booking';
	import type { Booking, PublicEventType } from '$lib/types';
	import { getLocalTimezones } from '$lib/timezones';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import MonthCalendar from '$lib/components/ui/MonthCalendar.svelte';
	import SearchableSelect from '$lib/components/ui/SearchableSelect.svelte';
	import Textarea from '$lib/components/ui/Textarea.svelte';

	let eventType = $state<PublicEventType | null>(null);
	let loading = $state(true);
	let notFound = $state(false);
	let timezoneInput = $state('UTC');
	let timezones = $state<string[]>(['UTC']);
	let localTimezone = $state('UTC');
	let month = $state('');
	let selectedDate = $state('');
	let selectedTime = $state('');
	let attendeeName = $state('');
	let attendeeEmail = $state('');
	let notes = $state('');
	let booking = $state<Booking | null>(null);
	let saving = $state(false);
	let now = $state(new Date());
	let clock: number | undefined;

	const timezone = $derived(timezones.includes(timezoneInput) ? timezoneInput : localTimezone);
	const minimumMonth = $derived(timezoneDateKey(now, timezone).slice(0, 7));
	const slotsByDate = $derived(
		eventType && month ? generateBookingSlots(eventType, timezone, month, now) : {}
	);
	const availableDates = $derived(Object.keys(slotsByDate));
	const selectedSlots = $derived(selectedDate ? (slotsByDate[selectedDate] ?? []) : []);
	const selectedDateLabel = $derived(
		selectedDate
			? new Intl.DateTimeFormat(undefined, { dateStyle: 'full', timeZone: 'UTC' }).format(
					new Date(`${selectedDate}T00:00:00Z`)
				)
			: 'Select a date'
	);
	const selectedTimeLabel = $derived.by(() => {
		if (!selectedTime || !eventType) return '';
		const start = new Date(selectedTime);
		const end = new Date(start.getTime() + eventType.durationMinutes * 60_000);
		const times = new Intl.DateTimeFormat(undefined, {
			timeZone: timezone,
			hour: 'numeric',
			minute: '2-digit'
		});
		const date = new Intl.DateTimeFormat(undefined, {
			timeZone: timezone,
			weekday: 'long',
			month: 'long',
			day: 'numeric',
			year: 'numeric'
		}).format(start);
		return `${times.format(start)} – ${times.format(end)}, ${date}`;
	});

	onMount(async () => {
		const local = getLocalTimezones();
		localTimezone = local.current;
		timezoneInput = local.current;
		timezones = local.options;
		selectedDate = timezoneDateKey(now, local.current);
		month = selectedDate.slice(0, 7);
		clock = window.setInterval(() => (now = new Date()), 60_000);
		try {
			const response = await callApi<{ eventType: PublicEventType }>(
				`/api/public/event-types/${encodeURIComponent(page.params.slug!)}`
			);
			eventType = response.eventType;
		} catch (cause) {
			notFound = cause instanceof Error && cause.message === 'event type not found';
		} finally {
			loading = false;
		}
	});

	onDestroy(() => {
		if (clock !== undefined) window.clearInterval(clock);
	});

	async function createBooking(event: SubmitEvent) {
		event.preventDefault();
		saving = true;
		try {
			const response = await callApi<{ booking: Booking }>('/api/bookings', {
				method: 'POST',
				body: JSON.stringify({
					eventSlug: eventType!.eventSlug,
					time: selectedTime,
					attendeeName,
					attendeeEmail,
					notes
				})
			});
			booking = response.booking;
		} catch {
			// callApi reports the error globally.
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head><title>{eventType?.name ?? 'Book'} · Let It Call</title></svelte:head>

{#if loading}
	<main class="grid min-h-screen place-items-center p-6"><p class="text-sm">Loading booking page…</p></main>
{:else if notFound || !eventType}
	<main class="grid min-h-screen place-items-center p-6">
		<section class="border border-black p-8 text-center">
			<h1 class="text-2xl font-semibold">Event not found</h1>
			<p class="mt-2 text-sm">This booking link is not available.</p>
		</section>
	</main>
{:else}
	<main class="min-h-screen p-4 sm:p-8 lg:p-10">
		<div class="mx-auto grid min-h-[calc(100vh-5rem)] max-w-7xl border border-black lg:grid-cols-[21rem_1fr]">
			<aside class="flex flex-col border-b border-black p-6 lg:border-r lg:border-b-0 lg:p-8">
				{#if selectedTime && !booking}
					<button
						type="button"
						class="mb-8 grid size-12 place-items-center border border-black hover:bg-black hover:text-white"
						onclick={() => (selectedTime = '')}
						aria-label="Back to date and time selection"
					>
						<Icon icon={arrowLeftIcon} width="24" height="24" />
					</button>
				{/if}
				<div class="flex -space-x-3">
					{#each eventType.hosts as host (host.email)}
						{#if host.avatarPath}
							<img src={avatarURL(host.avatarPath)} alt="" class="size-20 border border-black bg-white object-cover" />
						{/if}
					{/each}
				</div>
				<p class="mt-6 text-sm font-medium">{eventType.hosts.map((host) => host.email).join(', ')}</p>
				<h1 class="mt-2 text-3xl font-semibold tracking-tight">{eventType.name}</h1>
				<p class="mt-7 flex items-center gap-2 text-sm font-medium">
					<Icon icon={clockIcon} width="22" height="22" />
					{eventType.durationMinutes} min
				</p>
				{#if selectedTime}
					<p class="mt-5 flex items-start gap-2 text-sm font-medium">
						<Icon icon={calendarIcon} width="22" height="22" class="mt-0.5 shrink-0" />
						{selectedTimeLabel}
					</p>
					<p class="mt-4 flex items-center gap-2 text-sm font-medium">
						<Icon icon={worldIcon} width="22" height="22" />
						{timezone}
					</p>
				{/if}
				<p class="mt-auto hidden pt-12 text-xs lg:block">Powered by Let It Call</p>
			</aside>

			<section class="p-6 lg:p-10" aria-labelledby="booking-title">
				{#if booking}
					<div class="grid min-h-[32rem] place-items-center">
						<div class="max-w-lg border border-black p-8 text-center">
							<h2 id="booking-title" class="text-2xl font-semibold">Booking confirmed</h2>
							<p class="mt-3 text-sm">{eventType.name}</p>
							<p class="mt-1 text-sm">
								{new Intl.DateTimeFormat(undefined, { dateStyle: 'full', timeStyle: 'short', timeZone: timezone }).format(new Date(booking.time))}
							</p>
						</div>
					</div>
				{:else if selectedTime}
					<div class="max-w-2xl">
						<h2 id="booking-title" class="text-3xl font-semibold tracking-tight">Enter Details</h2>
						<form class="mt-7 grid gap-6" onsubmit={createBooking}>
							<Input id="attendee-name" label="Name" bind:value={attendeeName} required autocomplete="name" />
							<Input id="attendee-email" label="Email" type="email" bind:value={attendeeEmail} required autocomplete="email" />
							<Textarea
								id="booking-notes"
								label="Please share anything that will help prepare for our meeting."
								bind:value={notes}
								maxlength={2000}
							/>
							<div>
								<Button type="submit" disabled={saving}>{saving ? 'Scheduling…' : 'Schedule event'}</Button>
							</div>
						</form>
					</div>
				{:else}
					<h2 id="booking-title" class="text-3xl font-semibold tracking-tight">Select a Date & Time</h2>
					<div class="mt-9 grid gap-10 xl:grid-cols-[minmax(24rem,1fr)_minmax(18rem,0.75fr)]">
						<div>
							<MonthCalendar bind:month bind:selected={selectedDate} {availableDates} {minimumMonth} />
							<div class="mt-8 border-t border-black pt-6">
								<div class="mb-3 flex items-center gap-2 font-medium">
									<Icon icon={worldIcon} width="20" height="20" />
									<span>Time zone</span>
								</div>
								<SearchableSelect id="booking-timezone" label="Timezone" options={timezones} bind:value={timezoneInput} required />
							</div>
						</div>

						<div>
							<h3 class="border-b border-black pb-4 text-lg font-medium">{selectedDateLabel}</h3>
							{#if selectedSlots.length === 0}
								<div class="mt-5 border border-black p-6 text-center">
									<p class="font-semibold">No times available</p>
									<p class="mt-1 text-sm">Please select another date.</p>
								</div>
							{:else}
								<div class="mt-5 grid gap-2">
									{#each selectedSlots as slot (slot.time)}
										<button
											type="button"
											class={`min-h-12 border border-black px-4 text-sm font-medium ${selectedTime === slot.time ? 'bg-black text-white' : 'bg-white text-black hover:bg-black hover:text-white'}`}
											onclick={() => (selectedTime = slot.time)}
										>
											{slot.label}
										</button>
									{/each}
								</div>
							{/if}
						</div>
					</div>
				{/if}
			</section>
		</div>
	</main>
{/if}
