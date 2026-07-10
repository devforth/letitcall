<script lang="ts">
	import { goto } from '$app/navigation';
	import { onDestroy, onMount } from 'svelte';
	import { appPath, callApi } from '$lib/api';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import NumberInput from '$lib/components/ui/NumberInput.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import ScheduleEditor from '$lib/components/ScheduleEditor.svelte';
	import UserMultiSelect from '$lib/components/UserMultiSelect.svelte';
	import type { EventType, ManagedUser, ScheduleDay } from '$lib/types';

	let { slug = '' }: { slug?: string } = $props();

	let users = $state<ManagedUser[]>([]);
	let name = $state('');
	let durationChoice = $state('30');
	let customDuration = $state('30');
	let bookingWindowMode = $state('limited');
	let bookingWindowDays = $state('60');
	let inviteeLimit = $state('');
	let timezone = $state('UTC');
	let recipientEmails = $state<string[]>([]);
	let recipientsError = $state('');
	let schedule = $state<ScheduleDay[]>(defaultSchedule());
	let currentTime = $state('');
	let loading = $state(true);
	let saving = $state(false);

	const eventSlug = $derived(slug || slugify(name));
	let clock: number | undefined;

	onMount(async () => {
		try {
			const [usersResponse, sessionResponse] = await Promise.all([
				callApi<{ users: ManagedUser[] }>('/api/users'),
				callApi<{ user: ManagedUser }>('/api/auth/session')
			]);
			users = usersResponse.users;
			if (slug) {
				const response = await callApi<{ eventType: EventType }>(
					`/api/event-types/${encodeURIComponent(slug)}`
				);
				applyEventType(response.eventType);
			} else {
				timezone = Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC';
				recipientEmails = [sessionResponse.user.email];
			}
			updateCurrentTime();
			clock = window.setInterval(updateCurrentTime, 1000);
		} catch {
			// callApi reports the error globally.
		} finally {
			loading = false;
		}
	});

	onDestroy(() => {
		if (clock !== undefined) window.clearInterval(clock);
	});

	function applyEventType(eventType: EventType) {
		name = eventType.name;
		durationChoice = [15, 30, 45, 60].includes(eventType.durationMinutes)
			? String(eventType.durationMinutes)
			: 'custom';
		customDuration = String(eventType.durationMinutes);
		bookingWindowMode = eventType.bookingWindowDays === null ? 'unlimited' : 'limited';
		bookingWindowDays = String(eventType.bookingWindowDays ?? 60);
		inviteeLimit = eventType.inviteeLimit === null ? '' : String(eventType.inviteeLimit);
		timezone = eventType.timezone;
		recipientEmails = [...eventType.recipientEmails];
		schedule = eventType.schedule.map((day) => ({
			...day,
			start: day.start ?? '',
			end: day.end ?? '',
			breaks: day.breaks ?? []
		}));
	}

	function updateCurrentTime() {
		currentTime = new Intl.DateTimeFormat(undefined, {
			timeZone: timezone,
			dateStyle: 'medium',
			timeStyle: 'medium'
		}).format(new Date());
	}

	async function save(event: SubmitEvent) {
		event.preventDefault();
		if (recipientEmails.length === 0) {
			recipientsError = 'Select at least one recipient.';
			const form = event.currentTarget as HTMLFormElement;
			form.querySelector<HTMLElement>('#recipients')?.focus();
			return;
		}
		saving = true;
		try {
			const body = {
				name,
				durationMinutes: Number(durationChoice === 'custom' ? customDuration : durationChoice),
				bookingWindowDays:
					bookingWindowMode === 'unlimited' ? null : Number(bookingWindowDays),
				inviteeLimit: inviteeLimit === '' ? null : Number(inviteeLimit),
				timezone,
				recipientEmails,
				schedule: schedule.map((day) =>
					day.enabled
						? day
						: { day: day.day, enabled: false, start: '', end: '', breaks: [] }
				)
			};
			if (slug) {
				await callApi(`/api/event-types/${encodeURIComponent(slug)}`, {
					method: 'PUT',
					body: JSON.stringify(body)
				});
			} else {
				await callApi('/api/event-types', {
					method: 'POST',
					body: JSON.stringify(body)
				});
			}
			await goto(appPath('/scheduling'));
		} catch {
			// callApi reports the error globally.
		} finally {
			saving = false;
		}
	}

	function slugify(value: string) {
		return (value.toLocaleLowerCase().match(/[\p{L}\p{N}]+/gu) ?? []).join('-');
	}

	function defaultSchedule(): ScheduleDay[] {
		return ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'].map(
			(day, index) => ({
				day,
				enabled: index < 5,
				start: index < 5 ? '10:00' : '',
				end: index < 5 ? '16:00' : '',
				breaks: []
			})
		);
	}
</script>

{#if loading}
	<p class="border border-black p-6 text-sm">Loading event type…</p>
{:else}
	<form class="grid gap-5" onsubmit={save}>
		<div class="flex flex-wrap items-start justify-between gap-4">
			<div>
				<h1 class="text-2xl font-semibold tracking-tight">{slug ? 'Edit event type' : 'New event type'}</h1>
				<p class="mt-2 text-sm">Configure the booking duration, recipients, and availability.</p>
			</div>
			<div class="flex gap-2">
				<Button variant="secondary" onclick={() => goto(appPath('/scheduling'))}>Cancel</Button>
				<Button type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save event type'}</Button>
			</div>
		</div>

		<section class="grid gap-4 border border-black p-4 sm:grid-cols-2">
			<Input id="event-name" label="Event Name" bind:value={name} required />
			{#if slug}
				<Input id="event-slug" label="Event alias" value={eventSlug} readonly />
			{/if}
			<Select
				id="duration"
				label="Duration"
				bind:value={durationChoice}
				options={[
					{ value: '15', label: '15 minutes' },
					{ value: '30', label: '30 minutes' },
					{ value: '45', label: '45 minutes' },
					{ value: '60', label: '1 hour' },
					{ value: 'custom', label: 'Custom' }
				]}
			/>
			{#if durationChoice === 'custom'}
				<NumberInput id="custom-duration" label="Custom duration in minutes" bind:value={customDuration} max={1440} required />
			{/if}
			<Select
				id="booking-window-mode"
				label="How far ahead can invitees book?"
				bind:value={bookingWindowMode}
				options={[
					{ value: 'limited', label: 'Limit calendar days' },
					{ value: 'unlimited', label: 'No limit' }
				]}
			/>
			{#if bookingWindowMode === 'limited'}
				<NumberInput id="booking-window" label="Calendar days" bind:value={bookingWindowDays} required />
			{/if}
			<NumberInput id="invitee-limit" label="Invitees limit (empty means no limit)" bind:value={inviteeLimit} placeholder="No limit" />
			<Input id="timezone" label="Schedule timezone (read-only)" value={`${timezone} — ${currentTime}`} readonly />
		</section>

		<UserMultiSelect
			{users}
			bind:selected={recipientEmails}
			error={recipientsError}
			onchange={() => (recipientsError = '')}
		/>
		<ScheduleEditor bind:schedule />
	</form>
{/if}
