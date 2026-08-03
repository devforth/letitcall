<script lang="ts">
	import { page } from '$app/state';
	import { onDestroy, onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import stepCalendarIcon from '@iconify-icons/humbleicons/calendar';
	import stepCheckIcon from '@iconify-icons/humbleicons/check';
	import stepUserIcon from '@iconify-icons/la/user';
	import arrowRightIcon from '@iconify-icons/tabler/arrow-right';
	import clockIcon from '@iconify-icons/tabler/clock';
	import notesIcon from '@iconify-icons/tabler/align-left';
	import pencilIcon from '@iconify-icons/tabler/pencil';
	import userIcon from '@iconify-icons/tabler/user';
	import xIcon from '@iconify-icons/tabler/x';
	import calendarOffIcon from '@iconify-icons/tabler/calendar-off';
	import usersIcon from '@iconify-icons/tabler/users';
	import checkIcon from '@iconify-icons/tabler/check';
	import lockIcon from '@iconify-icons/material-symbols/lock';
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
	import { isValidEmail } from '$lib/validation';

	const blockStyle =
		'background: rgb(var(--color-foreground)); box-shadow: var(--shadow-small);';
	const boldStepUserIcon = {
		...stepUserIcon,
		body: stepUserIcon.body.replace('<path ', '<path stroke="currentColor" stroke-width="1" stroke-linejoin="round" ')
	};
	const boldStepCheckIcon = {
		...stepCheckIcon,
		body: stepCheckIcon.body.replace('stroke-width="2"', 'stroke-width="3"')
	};
	const boldArrowRightIcon = { ...arrowRightIcon, body: arrowRightIcon.body.replace('stroke-width="2"', 'stroke-width="3"') };
	const boldCheckIcon = { ...checkIcon, body: checkIcon.body.replace('stroke-width="2"', 'stroke-width="3"') };
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
	let scheduleAttempts = $state(0);
	let contactAttempts = $state(0);
	let now = $state(new Date());
	let clock: number | undefined;
	let availabilityClock: number | undefined;

	const bookingSteps = [
		{ title: 'Date and Time', subtitle: 'Find a time that works', icon: stepCalendarIcon },
		{ title: 'Contact Information', subtitle: 'Your name, email, and guests', icon: boldStepUserIcon },
		{ title: 'Confirmation', subtitle: "Review and you're booked", icon: boldStepCheckIcon }
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
	const showGuestFields = $derived(
		guestLimit === null || guestLimit > 0 || guestEmails.length > 0
	);
	const scheduleError = $derived(
		scheduleAttempts > 0 && !selectedTime ? 'Select an available time to continue' : ''
	);
	const attendeeNameError = $derived.by(() => {
		const name = attendeeName.trim();
		if (!name) return 'Enter your name';
		return name.length > 200 ? 'Name must be 200 characters or fewer' : '';
	});
	const attendeeEmailError = $derived.by(() => {
		const email = attendeeEmail.trim().toLowerCase();
		if (!email) return 'Enter your email address';
		if (!isValidEmail(email)) return 'Enter a valid email address';
		return '';
	});
	const guestEmailErrors = $derived.by(() => {
		const attendee = attendeeEmail.trim().toLowerCase();
		const seen = new Set<string>();
		return guestEmails.map((value, index) => {
			const email = value.trim().toLowerCase();
			if (!email) return `Enter an email address for guest ${index + 1}`;
			if (!isValidEmail(email)) return `Enter a valid email address for guest ${index + 1}`;
			if (email === attendee) return 'Your email address cannot also be added as a guest';
			if (seen.has(email)) return `Guest ${index + 1} repeats a previous email address`;
			seen.add(email);
			return '';
		});
	});
	const invalidGuestEmails = $derived(
		contactAttempts > 0 ? guestEmailErrors.map((error) => !!error) : []
	);
	const contactValidationErrors = $derived.by(() => [
		...new Set([attendeeNameError, attendeeEmailError, ...guestEmailErrors].filter((error) => !!error))
	]);
	const contactErrors = $derived(contactAttempts > 0 ? contactValidationErrors : []);
	const selectedDateLabel = $derived(
		selectedDate
			? new Intl.DateTimeFormat(undefined, { dateStyle: 'full', timeZone: 'UTC' }).format(
					new Date(`${selectedDate}T00:00:00Z`)
				)
			: 'Select a date'
	);
	// The review step shows the range and the date on separate lines, so they are
	// derived apart and joined again for the places that want one string.
	const selectedTimeRangeLabel = $derived.by(() => {
		if (!selectedTime || !eventType) return '';
		const start = new Date(selectedTime);
		const end = new Date(start.getTime() + eventType.durationMinutes * 60_000);
		const times = new Intl.DateTimeFormat(undefined, {
			timeZone: timezone,
			hour: 'numeric',
			minute: '2-digit'
		});
		return `${times.format(start)} – ${times.format(end)}`;
	});
	const selectedReviewDateLabel = $derived(
		selectedTime
			? new Intl.DateTimeFormat(undefined, {
					timeZone: timezone,
					weekday: 'long',
					month: 'long',
					day: 'numeric',
					year: 'numeric'
				}).format(new Date(selectedTime))
			: ''
	);
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
		if (!selectedTime) {
			scheduleAttempts += 1;
			return;
		}
		furthestStep = Math.max(furthestStep, 1);
		currentStep = 1;
	}

	function confirmContactInformation(event: SubmitEvent) {
		event.preventDefault();
		contactAttempts += 1;
		if (contactValidationErrors.length > 0) return;
		attendeeName = attendeeName.trim();
		attendeeEmail = attendeeEmail.trim().toLowerCase();
		guestEmails = guestEmails.map((email) => email.trim().toLowerCase());
		furthestStep = 2;
		currentStep = 2;
	}

	function goToStep(target: number) {
		// Only completed steps are navigable, and never once the booking is confirmed.
		if (booking || target > furthestStep) return;
		if (target === 0) month = selectedDate.slice(0, 7);
		currentStep = target;
	}

	function rows<T>(items: T[], perRow = 2): T[][] {
		const grouped: T[][] = [];
		for (let index = 0; index < items.length; index += perRow) {
			grouped.push(items.slice(index, index + perRow));
		}
		return grouped;
	}

	function stepState(index: number) {
		if (index === currentStep) return 'is-active';
		return index < furthestStep ? 'is-done' : 'is-upcoming';
	}

	function scrollFades(node: HTMLElement) {
		const shell = node.parentElement!;
		const update = () => {
			shell.classList.toggle('show-fade-top', node.scrollTop > 1);
			shell.classList.toggle('show-fade-bottom', node.scrollTop + node.clientHeight < node.scrollHeight - 1);
		};
		const resizeObserver = new ResizeObserver(update);
		const mutationObserver = new MutationObserver(update);
		node.addEventListener('scroll', update, { passive: true });
		resizeObserver.observe(node);
		mutationObserver.observe(node, { childList: true, subtree: true });
		requestAnimationFrame(update);
		return {
			destroy() {
				node.removeEventListener('scroll', update);
				resizeObserver.disconnect();
				mutationObserver.disconnect();
			}
		};
	}
</script>

<PageTitle title={eventType?.name ?? 'Book'} />

{#snippet stepError(messages: string[])}
	<div class="booking-step-error" role="alert">
		{#if messages.length === 1}
			<span>{messages[0]}</span>
		{:else}
			<ul>
				{#each messages as message (message)}
					<li>{message}</li>
				{/each}
			</ul>
		{/if}
	</div>
{/snippet}

{#snippet bookingSummary(editable: boolean)}
	<div class="review-editorial" class:review-editorial-static={!editable}>
		<section class="review-editorial-schedule">
			<p class="review-editorial-date">{selectedReviewDateLabel}</p>
			<p class="review-editorial-time">{selectedTimeRangeLabel}</p>
			<p class="review-editorial-muted">{timezone}</p>
			{#if editable}
				<button type="button" class="review-editorial-change" aria-label="Change schedule" title="Change schedule" onclick={() => goToStep(0)}>
					<Icon icon={pencilIcon} width="22" height="22" />
				</button>
			{/if}
		</section>
		<div class="review-editorial-details">
			<section>
				<p class="review-editorial-label"><span class="review-editorial-label-icon" aria-hidden="true"><Icon icon={userIcon} width="16" height="16" /></span>Attendee (you)</p>
				<p class="review-editorial-value">{attendeeName} · {attendeeEmail}</p>
			</section>
			{#if guestEmails.length > 0}
				<section>
					<p class="review-editorial-label"><span class="review-editorial-label-icon" aria-hidden="true"><Icon icon={usersIcon} width="16" height="16" /></span>Guests · {guestEmails.length}</p>
					<ul class="review-editorial-list">
						{#each guestEmails as email (email)}
							<li>{email}</li>
						{/each}
					</ul>
				</section>
			{/if}
			{#if notes}
				<section>
					<p class="review-editorial-label"><span class="review-editorial-label-icon" aria-hidden="true"><Icon icon={notesIcon} width="16" height="16" /></span>Notes</p>
					<p class="review-editorial-notes">{notes}</p>
				</section>
			{/if}
			{#if editable}
				<button type="button" class="review-editorial-change" aria-label="Change contact information" title="Change contact information" onclick={() => goToStep(1)}>
					<Icon icon={pencilIcon} width="22" height="22" />
				</button>
			{/if}
		</div>
	</div>
{/snippet}

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
		<div class="mx-auto grid min-h-[calc(100vh-5rem)] max-w-7xl overflow-hidden rounded-2xl lg:h-[calc(100vh-5rem)] lg:min-h-0 lg:grid-cols-[21rem_1fr]" style={blockStyle}>
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
				<div class="mt-auto flex items-center gap-3 pt-12 text-sm font-semibold">
					<BrandLogo class="size-10 rounded-xl object-cover" />
					<span>{branding.name}</span>
				</div>
				<div class="theme-toggle-contrast absolute bottom-6 right-6 lg:bottom-8 lg:right-8">
					<ThemeToggle />
				</div>
			</aside>

			<section class="flex min-h-0 flex-col overflow-hidden p-6 pt-4 lg:p-10 lg:pt-6" aria-label="Book a meeting">
				<div class="bk-stepper">
					<div class="bk-rail" aria-hidden="true">
						<div class="bk-track">
							<div class="bk-fill" style="--bk-progress: {furthestStep / (bookingSteps.length - 1)};"></div>
							<div class="bk-marks">
								{#each bookingSteps as step, i (step.title)}
									<span class="bk-dot {stepState(i)}">
										{#if i === currentStep}
											<Icon icon={step.icon} width="20" height="20" />
										{/if}
									</span>
								{/each}
							</div>
						</div>
					</div>

					<ol class="bk-labels">
						{#each bookingSteps as step, i (step.title)}
							<li class="bk-step {stepState(i)}" aria-current={i === currentStep ? 'step' : undefined}>
								<button
									type="button"
									class="bk-head"
									onclick={() => goToStep(i)}
									disabled={!!booking || i > furthestStep}
								>
									<span class="bk-title">{step.title}</span>
									<span class="bk-sub">{step.subtitle}</span>
								</button>
							</li>
						{/each}
					</ol>

					<div class="bk-content">
						{#if currentStep === 0}
							<div class="flex h-full flex-col">
								<div class="grid gap-10 xl:grid-cols-[minmax(20rem,1fr)_minmax(15rem,0.7fr)]">
								<div>
									<MonthCalendar bind:month bind:selected={selectedDate} {availableDates} {minimumMonth} today={timezoneDateKey(now, timezone)} />
								</div>
								<div>
									<h3 class="text-lg font-medium">
										<span class="block text-sm font-normal" style="color: rgb(var(--color-text) / 0.65);">Available times for</span>
										{selectedDateLabel}
									</h3>
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
										<table class="mt-3 w-full" style="border-collapse: separate; border-spacing: 8px; margin-left: -8px; margin-right: -8px; width: calc(100% + 16px); table-layout: fixed;">
											<tbody>
																					{#each rows(selectedSlots, 2) as row}
													<tr>
														{#each row as slot (slot.time)}
															{@const selected = slot.time === selectedTime}
															<td
																role="button"
																tabindex={slot.busy ? -1 : 0}
																aria-disabled={slot.busy}
																class="slot-cell px-4 text-center text-sm font-bold transition"
																class:is-busy={slot.busy}
																class:is-selected={selected}
																class:cursor-pointer={!slot.busy}
																class:cursor-not-allowed={slot.busy}
																onclick={() => !slot.busy && selectTime(slot.time)}
																onkeydown={(event) => {
																	if (slot.busy) return;
																	if (event.key === 'Enter' || event.key === ' ') {
																		event.preventDefault();
																		selectTime(slot.time);
																	}
																}}
															>
																<span class="inline-flex items-center gap-1 whitespace-nowrap" class:opacity-40={slot.busy}>
																	{#if slot.busy}<Icon icon={lockIcon} width="14" height="14" />{/if}
																	<span>{slot.label}</span>
																</span>
															</td>
														{/each}
													</tr>
												{/each}
											</tbody>
										</table>
									{/if}
									<div class="mt-8">
										<SearchableSelect id="booking-timezone" label="Timezone" icon={worldIcon} options={timezones} bind:value={timezoneInput} required />
										</div>
									</div>
								</div>
								<div class="booking-step-actions mt-auto flex items-center justify-end gap-4 pt-8">
									{#if scheduleError}
										{#key scheduleAttempts}{@render stepError([scheduleError])}{/key}
									{/if}
									<Button class="booking-next gap-2" onclick={confirmDateAndTime}>
										Next
										<Icon icon={boldArrowRightIcon} width="18" height="18" />
									</Button>
								</div>
							</div>
							{:else if currentStep === 1}
								<form class="flex h-full min-h-0 flex-col" autocomplete="off" novalidate onsubmit={confirmContactInformation}>
									<div class="booking-step-scroll-shell">
										<div class="booking-step-scroll grid content-start gap-8" use:scrollFades>
										<section class="grid content-start gap-3 md:w-1/2">
											<h3 class="text-lg font-medium">
												Personal details
											</h3>
											<div class="grid gap-5">
												<Input id="attendee-name" label="Name" icon="user" bind:value={attendeeName} required autocomplete="name" invalid={contactAttempts > 0 && !!attendeeNameError} />
												<Input id="attendee-email" label="Email" type="email" bind:value={attendeeEmail} required autocomplete="email" invalid={contactAttempts > 0 && !!attendeeEmailError} />
											</div>
										</section>
										{#if showGuestFields}
											<section class="grid content-start gap-3">
												{#if guestEmails.length > 0}
													<h3 class="text-lg font-medium">
														Guests
													</h3>
												{/if}
											<GuestEmailFields idPrefix="booking-guest" bind:emails={guestEmails} limit={guestLimit} legend={null} addLabel="Add guest" invalidEmails={invalidGuestEmails} />
										</section>
										{/if}
										<section class="grid gap-3">
											<h3 class="text-lg font-medium">
												Additional info <span class="booking-optional-label">(optional)</span>
											</h3>
											<Textarea
												id="booking-notes"
												label="Anything that will help prepare for our meeting"
												bind:value={notes}
												maxlength={2000}
											/>
										</section>
										</div>
									</div>
									<div class="booking-step-actions mt-auto flex items-center justify-end gap-4 pt-8">
										{#if contactErrors.length > 0}
											{#key contactAttempts}{@render stepError(contactErrors)}{/key}
										{/if}
										<Button type="submit" class="booking-next gap-2">
											Next
											<Icon icon={boldArrowRightIcon} width="18" height="18" />
										</Button>
									</div>
								</form>
								{:else if saving}
									<div class="booking-confirming" role="status" aria-live="polite">
										<span class="booking-confirming-spinner" aria-hidden="true"></span>
										<p class="booking-confirming-title">Confirming your booking…</p>
										<p class="booking-confirming-copy">Please keep this page open.</p>
									</div>
								{:else if booking}
									<section class="booking-confirmed" aria-labelledby="booking-confirmed-title">
										<p class="booking-confirmed-eyebrow">Booking confirmed</p>
										<h2 id="booking-confirmed-title" class="booking-confirmed-title">You’re booked for {eventType.name}</h2>
										<p class="booking-confirmed-copy">A confirmation has been sent to {attendeeEmail}.</p>
										{@render bookingSummary(false)}
										<div class="booking-confirmed-actions">
											<div class="booking-event-actions">
												<a class="booking-event-action booking-edit-action" href={`${manageURL}#event-details`}><Icon icon={pencilIcon} width="18" height="18" />Edit event</a>
												<a class="booking-event-action booking-cancel-action" href={`${manageURL}#cancel-event`}><Icon icon={xIcon} width="18" height="18" />Cancel event</a>
											</div>
											<a class="booking-new-link" href={page.url.pathname} data-sveltekit-reload>Make another booking</a>
										</div>
									</section>
								{:else}
									<form class="flex h-full min-h-0 flex-col" onsubmit={createBooking}>
										<div class="booking-step-scroll-shell">
											<div class="booking-step-scroll" use:scrollFades>
												<section class="review-details-section" aria-labelledby="review-details-title">
													<h2 id="review-details-title" class="review-details-title">Review your booking</h2>
													{@render bookingSummary(true)}
												</section>
											</div>
										</div>
										<div class="review-confirm mt-auto flex justify-end pt-8">
											<Button type="submit" class="booking-next booking-confirm gap-2">
												<Icon icon={boldCheckIcon} width="20" height="20" />
												Confirm booking
											</Button>
										</div>
									</form>
								{/if}
					</div>
				</div>
			</section>
		</div>
	</main>
{/if}

<style>
	.booking-optional-label {
		margin-left: 0.25rem;
		color: rgb(var(--color-text) / 0.55);
		font-size: 0.875rem;
		font-weight: 400;
	}

	.booking-step-error {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		width: 100%;
		min-height: 44px;
		border-left: 4px solid rgb(var(--error));
		background: rgb(var(--error) / 0.08);
		padding: 0.5rem 0.875rem;
		color: rgb(var(--error));
		font-size: 0.8125rem;
		font-weight: 600;
		line-height: 1.25;
		animation: booking-error-pulse 480ms ease-in-out;
	}

	.booking-step-error ul {
		display: grid;
		gap: 0.25rem;
		margin: 0;
		padding-left: 1rem;
		list-style: disc;
	}

	.booking-step-actions {
		align-items: flex-end;
		flex-direction: column;
		gap: 0.75rem;
	}

	@keyframes booking-error-pulse {
		0%, 100% { background: rgb(var(--error) / 0.08); }
		45% {
			background: rgb(var(--error) / 0.14);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.booking-step-error { animation: none; }
	}

	.booking-confirming {
		display: grid;
		min-height: 24rem;
		place-content: center;
		justify-items: center;
		gap: 0.75rem;
		text-align: center;
	}

	.booking-confirming-spinner {
		width: 2.75rem;
		height: 2.75rem;
		border: 3px solid rgb(var(--color-text) / 0.12);
		border-top-color: rgb(var(--color-primary));
		border-radius: 999px;
		animation: booking-confirming-spin 0.8s linear infinite;
	}

	.booking-confirming-title {
		font-size: 1.125rem;
		font-weight: 600;
	}

	.booking-confirming-copy {
		font-size: 0.875rem;
		color: rgb(var(--color-text) / 0.62);
	}

	@keyframes booking-confirming-spin {
		to { transform: rotate(360deg); }
	}

	.booking-confirmed-eyebrow {
		font-size: 0.75rem;
		font-weight: 600;
		color: rgb(var(--color-primary));
	}

	.booking-confirmed-title {
		margin-top: 0.375rem;
		font-size: 1.5rem;
		font-weight: 600;
		line-height: 1.25;
	}

	.booking-confirmed-copy {
		margin-top: 0.5rem;
		font-size: 0.875rem;
		color: rgb(var(--color-text) / 0.62);
	}

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

	.review-details-title {
		font-size: 1.125rem;
		font-weight: 500;
		line-height: 1.5rem;
	}

	.review-editorial {
		display: grid;
		grid-template-columns: minmax(13rem, 0.8fr) minmax(0, 1.2fr);
		overflow: hidden;
		margin-top: 1rem;
		border: 1px solid rgb(var(--color-border));
		border-radius: 1rem;
	}

	.review-editorial-schedule {
		position: relative;
		padding: 1.25rem;
		padding-bottom: 5rem;
		border-right: 1px solid rgb(var(--color-border));
	}

	.review-editorial-static .review-editorial-schedule {
		padding-bottom: 1.25rem;
	}

	.review-editorial-change {
		display: grid;
		position: absolute;
		right: 1.25rem;
		bottom: 1.25rem;
		place-items: center;
		width: 3rem;
		height: 3rem;
		border: 0;
		border-radius: 0.875rem;
		background: rgb(var(--color-primary) / 0.12);
		padding: 0;
		color: rgb(var(--color-primary));
		cursor: pointer;
		opacity: 0;
		transition: background 0.2s ease, opacity 0.2s ease;
	}

	.review-editorial-change:hover {
		background: rgb(var(--color-primary) / 0.2);
	}

	.review-editorial-schedule:hover .review-editorial-change,
	.review-editorial-schedule:focus-within .review-editorial-change,
	.review-editorial-details:hover .review-editorial-change,
	.review-editorial-details:focus-within .review-editorial-change {
		opacity: 1;
	}

	.review-editorial-change:hover {
		text-decoration: underline;
	}

	.review-editorial-date {
		font-size: 1.5rem;
		font-weight: 600;
		line-height: 1.25;
	}

	.review-editorial-time {
		margin-top: 0.25rem;
		font-size: 1rem;
		font-weight: 600;
		line-height: 1.5rem;
	}

	.review-editorial-details section {
		padding: 0;
	}

	.review-editorial-details {
		position: relative;
		display: grid;
		align-content: start;
		gap: 1.5rem;
		padding: 1.25rem 1.25rem 5rem;
	}

	.review-editorial-static .review-editorial-details {
		padding-bottom: 1.25rem;
	}

	.review-editorial-label {
		display: flex;
		align-items: center;
		gap: 0.375rem;
		font-size: 0.75rem;
		font-weight: 400;
		color: rgb(var(--color-text) / 0.6);
	}

	.review-editorial-label-icon {
		display: grid;
		place-items: center;
		color: inherit;
	}

	.review-editorial-value {
		margin: 0.25rem 0 0 1.4rem;
		font-size: 1rem;
		font-weight: 400;
		line-height: 1.5rem;
		overflow-wrap: anywhere;
	}

	.review-editorial-muted {
		font-size: 0.8125rem;
		overflow-wrap: anywhere;
		color: rgb(var(--color-text) / 0.62);
	}

	.review-editorial-list {
		display: grid;
		gap: 0.125rem;
		margin: 0.25rem 0 0 1.4rem;
		padding: 0;
		list-style: none;
		font-size: 1rem;
		font-weight: 400;
		line-height: 1.5rem;
		color: rgb(var(--color-text));
		overflow-wrap: anywhere;
	}

	.review-editorial-notes {
		margin: 0.25rem 0 0 1.4rem;
		font-size: 1rem;
		font-weight: 400;
		line-height: 1.5;
		color: rgb(var(--color-text));
		white-space: pre-wrap;
	}

	@media (hover: none) {
		.review-editorial-change {
			opacity: 1;
		}
	}

	/* Time-slot cells: an arrow-shaped primary fill sweeps in from the left. */
	.slot-cell {
		position: relative;
		z-index: 0;
		overflow: hidden;
		height: 3.25rem;
		vertical-align: middle;
		border-radius: 10px;
		background: rgb(var(--color-text) / 0.05);
		color: rgb(var(--color-text));
	}
	.slot-cell::before {
		content: '';
		position: absolute;
		inset: 0;
		right: 100%;
		background: rgb(var(--color-primary));
		clip-path: polygon(0 0, calc(100% - 12px) 0, 100% 50%, calc(100% - 12px) 100%, 0 100%);
		transition: right 0.25s ease;
		z-index: -1;
	}
	.slot-cell.is-selected::before {
		right: -12px;
	}
	/* Hover on an unselected slot: just a little tint, not the full fill. */
	.slot-cell:not(.is-busy):not(.is-selected):hover {
		background: rgb(var(--color-primary) / 0.12) !important;
		color: rgb(var(--color-primary)) !important;
	}
	.slot-cell.is-selected {
		color: rgb(var(--color-contrast-text)) !important;
	}
	/* Click/press feedback — same as the calendar day cells. */
	.slot-cell:not(.is-busy):active {
		filter: brightness(0.92);
	}
	/* Active slot: same inset ring as the selected calendar date. */
	.slot-cell.is-selected::after {
		content: '';
		position: absolute;
		inset: 0;
		border-radius: inherit;
		box-shadow: inset 0 0 0 4px rgb(var(--color-background) / 0.3);
		pointer-events: none;
	}

	/* Toggle on the primary aside: no fill, contrast-color border and icon. */
	.theme-toggle-contrast :global(.toggle-switch) {
		background: transparent !important;
		border-color: rgb(var(--color-contrast-text)) !important;
		color: rgb(var(--color-contrast-text)) !important;
		box-shadow: none !important;
	}

	/* Booking progress: one filling rail, a marker per step, labels beneath. */
	.bk-stepper {
		display: flex;
		flex: 1;
		min-height: 0;
		flex-direction: column;
		gap: 0;
	}
	/* The rail row is as tall as a marker, so the circles never overlap the labels. */
	.bk-rail {
		display: flex;
		align-items: center;
		height: 36px;
	}
	.bk-track {
		position: relative;
		flex: 1;
		height: 6px;
		/* Ends sit under the centres of the outer labels, a sixth in from each side. */
		margin: 0 calc(100% / 6);
		border-radius: 999px;
		background: rgb(var(--color-text) / 0.12);
	}
	/* Fill spans from the first marker to the furthest one reached. */
	.bk-fill {
		position: absolute;
		inset: 0 auto 0 0;
		width: calc(var(--bk-progress) * 100%);
		border-radius: 999px;
		background: rgb(var(--color-primary));
		transition: width 0.4s cubic-bezier(0.4, 0, 0.2, 1);
	}
	/* Markers straddle the ends of the track, so they sit centred on it. */
	.bk-marks {
		position: absolute;
		inset: 0 -18px;
		display: flex;
		align-items: center;
		justify-content: space-between;
	}
	.bk-dot {
		display: grid;
		place-items: center;
		width: 36px;
		height: 36px;
		border-radius: 999px;
		background: rgb(var(--color-foreground));
		box-shadow: inset 0 0 0 4px rgb(var(--color-text) / 0.18);
		color: rgb(var(--color-text) / 0.5);
		transition:
			background 0.3s,
			box-shadow 0.3s,
			color 0.3s,
			transform 0.3s;
	}
	.bk-dot.is-done {
		background: rgb(var(--color-primary));
		box-shadow: none;
		color: rgb(var(--color-contrast-text));
	}
	.bk-dot.is-active {
		box-shadow: inset 0 0 0 2px rgb(var(--color-primary));
		color: rgb(var(--color-primary));
	}
	/* Scale plain markers inside a fixed footprint so changing steps does not
	   reflow the rail while the active icon appears. */
	.bk-dot:not(.is-active) {
		transform: scale(0.5);
	}
	.bk-labels {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		list-style: none;
		margin: 0;
		padding: 0;
	}
	.bk-head {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 2px;
		width: 100%;
		padding: 0;
		border: 0;
		background: transparent;
		font: inherit;
		color: rgb(var(--color-text) / 0.5);
		text-align: center;
	}
	.bk-head:not(:disabled) {
		cursor: pointer;
	}
	.bk-head:not(:disabled):hover .bk-title {
		text-decoration: underline;
	}
	.bk-title {
		font-size: 15px;
		font-weight: 600;
		line-height: 1.25;
		transition: font-size 0.3s;
	}
	.bk-step.is-done .bk-head {
		color: rgb(var(--color-primary));
	}
	.bk-step.is-active .bk-head {
		color: rgb(var(--color-primary));
	}
	.bk-step.is-active .bk-title {
		font-size: 17px;
	}
	.bk-sub {
		font-size: 12.5px;
		font-weight: 500;
		line-height: 1.35;
		color: rgb(var(--color-text) / 0.5);
	}
	.bk-step.is-active .bk-sub {
		color: rgb(var(--color-text) / 0.7);
	}
	.bk-step.is-upcoming .bk-sub {
		color: rgb(var(--color-text) / 0.35);
	}
	.bk-content {
		flex: 1;
		min-height: 0;
		overflow: hidden;
		margin-top: 52px;
	}

	.booking-step-scroll-shell {
		position: relative;
		flex: 1;
		min-height: 0;
	}

	.booking-step-scroll-shell::before,
	.booking-step-scroll-shell::after {
		content: '';
		position: absolute;
		right: 0.75rem;
		left: 0;
		height: 3.5rem;
		z-index: 1;
		opacity: 0;
		pointer-events: none;
		transition: opacity 160ms ease;
	}

	.booking-step-scroll-shell::before {
		top: 0;
		background: linear-gradient(to bottom, rgb(var(--color-foreground)), transparent);
	}

	.booking-step-scroll-shell::after {
		bottom: 0;
		background: linear-gradient(to bottom, transparent, rgb(var(--color-foreground)));
	}

	.booking-step-scroll-shell.show-fade-top::before,
	.booking-step-scroll-shell.show-fade-bottom::after {
		opacity: 1;
	}

	.booking-step-scroll {
		height: 100%;
		overflow-x: hidden;
		overflow-y: auto;
		padding-bottom: 2.5rem;
		scrollbar-color: rgb(var(--color-text) / 0.28) transparent;
		scrollbar-width: thin;
	}

	.booking-step-scroll::-webkit-scrollbar {
		width: 8px;
	}

	.booking-step-scroll::-webkit-scrollbar-track {
		background: transparent;
	}

	.booking-step-scroll::-webkit-scrollbar-thumb {
		border: 2px solid rgb(var(--color-foreground));
		border-radius: 999px;
		background: rgb(var(--color-text) / 0.28);
	}

	.booking-step-scroll::-webkit-scrollbar-thumb:hover {
		background: rgb(var(--color-text) / 0.45);
	}

	@media (prefers-reduced-motion: reduce) {
		.booking-step-scroll-shell::before,
		.booking-step-scroll-shell::after {
			transition: none;
		}
	}
	/* Too narrow for three subtitles side by side — keep the current one only. */
	@media (max-width: 640px) {
		.booking-step-actions {
			align-items: stretch;
			flex-direction: column;
		}

		.booking-step-actions :global(button) {
			width: 100%;
		}

		.booking-confirmed-actions {
			align-items: stretch;
			flex-direction: column;
		}

		.booking-event-actions {
			flex-direction: column;
		}

		.booking-event-action {
			text-align: center;
		}

		.review-editorial {
			grid-template-columns: minmax(0, 1fr);
		}

		.review-editorial-schedule {
			border-right: 0;
			border-bottom: 1px solid rgb(var(--color-border));
		}

		.review-confirm {
			justify-content: stretch;
			padding-top: 1.5rem;
		}

		.review-confirm :global(button) {
			width: 100%;
			justify-content: center;
		}

		.bk-title,
		.bk-step.is-active .bk-title {
			font-size: 13.5px;
		}
		.bk-step:not(.is-active) .bk-sub {
			display: none;
		}
	}
</style>
