<script lang="ts">
	import { page } from '$app/state';
	import { onDestroy, onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import clockIcon from '@iconify-icons/tabler/clock';
	import calendarIcon from '@iconify-icons/tabler/calendar';
	import calendarOffIcon from '@iconify-icons/tabler/calendar-off';
	import worldIcon from '@iconify-icons/tabler/world';
	import { callApi } from '$lib/api';
	import { firstAvailableDate, generateBookingSlots, timezoneDateKey } from '$lib/booking';
	import type { Booking, PublicEventType } from '$lib/types';
	import { getLocalTimezones } from '$lib/timezones';
	import Button from '$lib/components/ui/Button.svelte';
	import GuestEmailFields from '$lib/components/GuestEmailFields.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import { branding } from '$lib/stores/branding.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import MonthCalendar from '$lib/components/ui/MonthCalendar.svelte';
	import SearchableSelect from '$lib/components/ui/SearchableSelect.svelte';
	import BrandLogo from '$lib/components/BrandLogo.svelte';
	import Textarea from '$lib/components/ui/Textarea.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';

	const blockStyle =
		'background: rgb(var(--color-foreground)); box-shadow: var(--shadow-small);';
	const dividerStyle = 'border-color: rgb(var(--color-border));';
	const asideStyle =
		'background: rgb(var(--color-primary)); color: rgb(var(--color-contrast-text)); box-shadow: 0 0 0 1px rgb(var(--color-border)), var(--shadow-small);';

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
	let guestEmails = $state<string[]>([]);
	let notes = $state('');
	let booking = $state<Booking | null>(null);
	let manageURL = $state('');
	let saving = $state(false);
	let currentStep = $state(0);
	let furthestStep = $state(0);
	let now = $state(new Date());
	let clock: number | undefined;
	let availabilityClock: number | undefined;

	const bookingSteps = [
		{ title: 'Date and Time' },
		{ title: 'Contact Information' },
		{ title: 'Confirmation' }
	];
	const timezone = $derived(timezones.includes(timezoneInput) ? timezoneInput : localTimezone);
	const minimumMonth = $derived(timezoneDateKey(now, timezone).slice(0, 7));
	const slotsByDate = $derived(
		eventType && month ? generateBookingSlots(eventType, timezone, month, now) : {}
	);
	const availableDates = $derived(Object.keys(slotsByDate));
	const selectedSlots = $derived.by(() => {
		if (!selectedDate || !eventType) return [];
		// Slots come from the selected date's own month, so browsing to a different
		// month in the calendar doesn't clear the times already shown.
		const selectedMonth = selectedDate.slice(0, 7);
		const map =
			selectedMonth === month ? slotsByDate : generateBookingSlots(eventType, timezone, selectedMonth, now);
		return map[selectedDate] ?? [];
	});
	const hosts = $derived(eventType ? [...eventType.requiredHosts, ...eventType.optionalHosts] : []);

	const guestLimit = $derived.by(() => {
		if (!eventType || eventType.inviteeLimit === null || !selectedTime) return null;
		const remaining = eventType.remainingInvitees[selectedTime] ?? eventType.inviteeLimit;
		return Math.max(0, remaining - 1);
	});
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
			await refreshEventType();
			// Land on the nearest day (today or later) that actually has bookable times.
			selectedDate = firstAvailableDate(eventType!, timezone, now);
			month = selectedDate.slice(0, 7);
			availabilityClock = window.setInterval(refreshEventType, 20_000);
		} catch (cause) {
			notFound = cause instanceof Error && cause.message === 'event type not found';
		} finally {
			loading = false;
		}
	});

	onDestroy(() => {
		if (clock !== undefined) window.clearInterval(clock);
		if (availabilityClock !== undefined) window.clearInterval(availabilityClock);
	});

	async function refreshEventType() {
		const response = await callApi<{ eventType: PublicEventType }>(
			`/api/public/event-types/${encodeURIComponent(page.params.slug!)}`
		);
		if (!booking && selectedTime && isBusy(response.eventType, selectedTime)) {
			selectedTime = '';
			currentStep = 0;
			furthestStep = 0;
		}
		eventType = response.eventType;
	}

	function isBusy(candidate: PublicEventType, time: string) {
		const start = new Date(time);
		const end = new Date(start.getTime() + candidate.durationMinutes * 60_000);
		return candidate.busyRanges.some(
			(range) => start < new Date(range.end) && end > new Date(range.start)
		);
	}

	async function createBooking(event: SubmitEvent) {
		event.preventDefault();
		saving = true;
		try {
			const response = await callApi<{ booking: Booking; manageURL: string }>('/api/bookings', {
				method: 'POST',
				body: JSON.stringify({
					eventSlug: eventType!.eventSlug,
					time: selectedTime,
					attendeeName,
					attendeeEmail,
					attendeeTimezone: timezone,
					guestEmails,
					notes
				})
			});
			booking = response.booking;
			manageURL = response.manageURL;
		} catch {
			// callApi reports the error globally.
		} finally {
			saving = false;
		}
	}

	function selectTime(time: string) {
		selectedTime = time;
		furthestStep = 0;
		if (guestLimit !== null) guestEmails = guestEmails.slice(0, guestLimit);
	}

	function confirmDateAndTime() {
		furthestStep = Math.max(furthestStep, 1);
		currentStep = 1;
	}

	function confirmContactInformation(event: SubmitEvent) {
		event.preventDefault();
		furthestStep = 2;
		currentStep = 2;
	}

	function goToStep(target: number) {
		// Only completed steps are navigable, and never once the booking is confirmed.
		if (booking || target > furthestStep) return;
		if (target === 0) month = selectedDate.slice(0, 7);
		currentStep = target;
	}

	function stepState(index: number) {
		if (index === currentStep) return 'is-active';
		return index < furthestStep ? 'is-done' : 'is-upcoming';
	}
</script>

<PageTitle title={eventType?.name ?? 'Book'} />

{#if loading}
	<main class="grid min-h-screen place-items-center p-6"><p class="text-sm">Loading booking page…</p></main>
{:else if notFound || !eventType}
	<main class="grid min-h-screen place-items-center p-6">
		<section class="rounded-2xl p-8 text-center" style={blockStyle}>
			<h1 class="text-2xl font-semibold">Event not found</h1>
			<p class="mt-2 text-sm">This booking link is not available.</p>
		</section>
	</main>
{:else}
	<main class="min-h-screen p-4 sm:p-8 lg:p-10">
		<div class="mx-auto grid min-h-[calc(100vh-5rem)] max-w-7xl overflow-hidden rounded-2xl lg:grid-cols-[21rem_1fr]" style={blockStyle}>
			<aside class="relative flex flex-col rounded-b-2xl p-6 lg:rounded-bl-none lg:rounded-tr-2xl lg:rounded-br-2xl lg:p-8" style={asideStyle}>
				<h1 class="text-3xl font-semibold tracking-tight">{eventType.name}</h1>
				<div class="mt-8 flex -space-x-4">
					{#each hosts as host (host.email)}
						<Avatar
							name={host.fullName}
							email={host.email}
							avatarPath={host.avatarPath}
							size={76}
							rounded="full"
							class="bg-none! bg-[rgb(var(--color-foreground))]! shadow-[0_0_0_4px_rgb(var(--color-primary))]"
						/>
					{/each}
				</div>
				<p class="mt-6 text-sm font-medium">{eventType.requiredHosts.map((host) => host.fullName || host.email).join(', ')}</p>
				{#if eventType.optionalHosts.length > 0}
					<p class="mt-1 text-xs">Optional: {eventType.optionalHosts.map((host) => host.fullName || host.email).join(', ')}</p>
				{/if}
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
				<div class="mt-auto flex items-center gap-3 pt-12 text-sm font-semibold">
					<BrandLogo class="size-10 rounded-xl object-cover" />
					<span>{branding.name}</span>
				</div>
				<div class="theme-toggle-contrast absolute bottom-6 right-6 lg:bottom-8 lg:right-8">
					<ThemeToggle />
				</div>
			</aside>

			<section class="p-6 lg:p-10" aria-label="Book a meeting">
				<ol class="bk-stepper">
					{#each bookingSteps as step, i (step.title)}
		<li
			class="bk-step {stepState(i)}"
							aria-current={i === currentStep ? 'step' : undefined}
						>
							<button
								type="button"
								class="bk-head"
								onclick={() => goToStep(i)}
				disabled={!!booking || i > furthestStep}
							>
								<span class="bk-ind">{i + 1}</span>
								<span class="bk-label">
									<span class="bk-title">{step.title}</span>
								</span>
							</button>

							{#if i === currentStep}
								<div class="bk-content">
									{#if i === 0}
										<div class="grid gap-10 xl:grid-cols-[minmax(20rem,1fr)_minmax(15rem,0.7fr)]">
											<div class="xl:border-r-2 xl:pr-10" style={dividerStyle}>
												<MonthCalendar bind:month bind:selected={selectedDate} {availableDates} {minimumMonth} today={timezoneDateKey(now, timezone)} />
											</div>
											<div>
												<h3 class="text-lg font-medium">{selectedDateLabel}</h3>
												{#if selectedSlots.length === 0}
													<div
														class="mt-5 flex flex-col items-center gap-3 rounded-2xl border-2 border-dashed px-6 py-10 text-center"
														style="border-color: rgb(var(--color-border));"
													>
														<span
															class="grid size-12 place-items-center rounded-full"
															style="background: rgb(var(--color-primary) / 0.1); color: rgb(var(--color-primary));"
														>
															<Icon icon={calendarOffIcon} width="24" height="24" />
														</span>
														<div>
															<p class="font-semibold">No times available</p>
															<p class="mt-1 text-sm" style="color: rgb(var(--color-text) / 0.6);">Please select another date.</p>
														</div>
													</div>
												{:else}
													<div class="mt-5 grid grid-cols-2 gap-2">
														{#each selectedSlots as slot (slot.time)}
															<Button
																variant={slot.time === selectedTime ? 'primary' : 'secondary'}
																fullWidth
																disabled={slot.busy}
																onclick={() => selectTime(slot.time)}
															>
																<span class="flex min-h-8 flex-col items-center justify-center">
																	<span>{slot.label}</span>
																	{#if slot.busy}<span class="text-xs font-normal">Busy</span>{/if}
																</span>
															</Button>
														{/each}
													</div>
												{/if}
												<div class="mt-8">
													<SearchableSelect id="booking-timezone" label="Timezone" icon={worldIcon} options={timezones} bind:value={timezoneInput} required />
													</div>
												</div>
											</div>
											<div class="mt-8">
												<Button disabled={!selectedTime} onclick={confirmDateAndTime}>Confirm date and time</Button>
											</div>
										{:else if i === 1}
											<form class="grid max-w-xl gap-6" onsubmit={confirmContactInformation}>
											<Input id="attendee-name" label="Name" bind:value={attendeeName} required autocomplete="name" />
											<Input id="attendee-email" label="Email" type="email" bind:value={attendeeEmail} required autocomplete="email" />
											<GuestEmailFields idPrefix="booking-guest" bind:emails={guestEmails} limit={guestLimit} />
											<Textarea
												id="booking-notes"
												label="Please share anything that will help prepare for our meeting."
												bind:value={notes}
												maxlength={2000}
												/>
												<div>
													<Button type="submit">Confirm contact information</Button>
											</div>
										</form>
											{:else if booking}
												<div class="max-w-lg rounded-2xl p-6" style={blockStyle}>
													<p class="text-sm font-medium">{eventType.name}</p>
											<p class="mt-1 text-sm">
												{new Intl.DateTimeFormat(undefined, { dateStyle: 'full', timeStyle: 'short', timeZone: timezone }).format(new Date(booking.time))}
											</p>
											<a class="mt-6 inline-block rounded-xl bg-[rgb(var(--color-primary)/0.12)] px-4 py-3 text-sm font-semibold text-[rgb(var(--color-primary))] transition-colors hover:bg-[rgb(var(--color-primary)/0.2)]" href={manageURL}>Cancel or update event</a>
													<a class="mt-4 block text-sm font-medium underline hover:no-underline" href={page.url.pathname} data-sveltekit-reload>Make another booking</a>
												</div>
											{:else}
												<div class="grid max-w-xl gap-6">
													<div class="rounded-2xl border-2 p-5" style={dividerStyle}>
														<h3 class="text-lg font-semibold">Review your booking</h3>
														<div class="mt-5 grid gap-4 text-sm">
															<div>
																<p class="text-xs font-medium" style="color: rgb(var(--color-text) / 0.6);">Date and time</p>
																<p class="mt-1 font-medium">{selectedTimeLabel}</p>
															</div>
															<div>
																<p class="text-xs font-medium" style="color: rgb(var(--color-text) / 0.6);">Timezone</p>
																<p class="mt-1 font-medium">{timezone}</p>
															</div>
															<div>
																<p class="text-xs font-medium" style="color: rgb(var(--color-text) / 0.6);">Contact</p>
																<p class="mt-1 font-medium">{attendeeName}</p>
																<p>{attendeeEmail}</p>
															</div>
															{#if guestEmails.length > 0}
																<div>
																	<p class="text-xs font-medium" style="color: rgb(var(--color-text) / 0.6);">Additional guests</p>
																	<p class="mt-1">{guestEmails.join(', ')}</p>
																</div>
															{/if}
															<div>
																<p class="text-xs font-medium" style="color: rgb(var(--color-text) / 0.6);">Notes</p>
																<p class="mt-1 whitespace-pre-wrap">{notes || 'None'}</p>
															</div>
														</div>
													</div>
													<form onsubmit={createBooking}>
														<Button type="submit" disabled={saving}>{saving ? 'Scheduling…' : 'Confirm booking'}</Button>
													</form>
												</div>
											{/if}
								</div>
							{/if}
						</li>
					{/each}
				</ol>
			</section>
		</div>
	</main>
{/if}

<style>
	/* Toggle on the primary aside: no fill, contrast-color border and icon. */
	.theme-toggle-contrast :global(.toggle-switch) {
		background: transparent !important;
		border-color: rgb(var(--color-contrast-text)) !important;
		color: rgb(var(--color-contrast-text)) !important;
		box-shadow: none !important;
	}

	/* Nuxt UI-style vertical progress stepper for the booking flow. */
	.bk-stepper {
		display: flex;
		flex-direction: column;
		list-style: none;
		padding: 0;
	}
	.bk-step {
		position: relative;
		display: grid;
		grid-template-columns: 34px 1fr;
		align-items: start;
		column-gap: 14px;
		padding-bottom: 28px;
	}
	.bk-step:last-child {
		padding-bottom: 0;
	}
	/* connector runs from the bottom of this indicator down to the next */
	.bk-step::before {
		content: '';
		position: absolute;
		left: 16px; /* centre of the 34px indicator */
		top: 40px; /* 6px gap below the circle */
		bottom: 6px; /* 6px gap above the next circle */
		border-left: 2px dashed rgb(var(--color-text) / 0.15);
	}
	.bk-step:last-child::before {
		display: none;
	}
	.bk-step.is-done::before {
		border-color: rgb(var(--color-primary));
	}
	/* Header spans participate in the .bk-step grid directly. */
	.bk-head {
		display: contents;
		font: inherit;
		color: inherit;
		text-align: left;
	}
	.bk-head:not(:disabled) .bk-ind,
	.bk-head:not(:disabled) .bk-label {
		cursor: pointer;
	}
	.bk-head:not(:disabled):hover .bk-title {
		text-decoration: underline;
	}
	.bk-ind {
		grid-column: 1;
		grid-row: 1;
		width: 34px;
		height: 34px;
		border-radius: 9999px;
		display: flex;
		align-items: center;
		justify-content: center;
		line-height: 1;
		font-weight: 600;
		font-size: 18px;
		font-variant-numeric: tabular-nums;
		transition:
			background 0.3s,
			color 0.3s,
			box-shadow 0.3s;
	}
	.bk-step.is-done .bk-ind {
		background: rgb(var(--color-primary));
		color: rgb(var(--color-contrast-text));
	}
	.bk-step.is-active .bk-ind {
		background: transparent;
		color: rgb(var(--color-primary));
		box-shadow: inset 0 0 0 2px rgb(var(--color-primary));
	}
	.bk-step.is-upcoming .bk-ind {
		background: transparent;
		color: rgb(var(--color-text) / 0.6);
		box-shadow: inset 0 0 0 2px rgb(var(--color-border));
	}
	.bk-label {
		grid-column: 2;
		grid-row: 1;
		display: flex;
		flex-direction: column;
		justify-content: center;
		gap: 2px;
		min-height: 34px;
	}
	.bk-content {
		grid-column: 2;
		grid-row: 2;
		margin-top: 18px;
	}
	.bk-title {
		font-size: 22px;
		font-weight: 600;
		line-height: 1.2;
		color: rgb(var(--color-text));
	}
	.bk-step.is-upcoming .bk-title {
		color: rgb(var(--color-text) / 0.6);
	}
</style>
