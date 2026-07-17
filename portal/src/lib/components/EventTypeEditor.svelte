<script lang="ts">
	import { goto } from '$app/navigation';
	import { onDestroy, onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import calendarEventIcon from '@iconify-icons/tabler/calendar-event';
	import checkIcon from '@iconify-icons/tabler/check';
	import xIcon from '@iconify-icons/tabler/x';
	import { appPath, callApi } from '$lib/api';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import NumberInput from '$lib/components/ui/NumberInput.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import ScheduleEditor from '$lib/components/ScheduleEditor.svelte';
	import HostSelector from '$lib/components/HostSelector.svelte';
	import type { EventType, ManagedUser, ScheduleDay } from '$lib/types';

	let { slug = '' }: { slug?: string } = $props();

	let users = $state<ManagedUser[]>([]);
	let name = $state('');
	let durationChoice = $state('30');
	let customDuration = $state('30');
	let bookingWindowDays = $state('60');
	let inviteeLimit = $state('');
	let timezone = $state('UTC');
	let requiredHostEmails = $state<string[]>([]);
	let optionalHostEmails = $state<string[]>([]);
	let hostsError = $state('');
	let schedule = $state<ScheduleDay[]>(defaultSchedule());
	let currentTime = $state('');
	let loading = $state(true);
	let saving = $state(false);

	const eventSlug = $derived(slug || slugify(name));
	const blockStyle =
		'background: rgb(var(--color-foreground)); border-color: rgb(var(--color-border)); box-shadow: var(--shadow-small);';
	const eventDetailsContainerStyle = `${blockStyle} background: rgb(var(--color-primary));`;
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
				requiredHostEmails = [sessionResponse.user.email];
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
		bookingWindowDays = String(eventType.bookingWindowDays);
		inviteeLimit = eventType.inviteeLimit === null ? '' : String(eventType.inviteeLimit);
		timezone = eventType.timezone;
		requiredHostEmails = [...eventType.requiredHostEmails];
		optionalHostEmails = [...eventType.optionalHostEmails];
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
		if (requiredHostEmails.length === 0) {
			hostsError = 'Select at least one required host.';
			const form = event.currentTarget as HTMLFormElement;
			form.querySelector<HTMLElement>('#hosts')?.focus();
			return;
		}
		saving = true;
		try {
			const body = {
				name,
				durationMinutes: Number(durationChoice === 'custom' ? customDuration : durationChoice),
				bookingWindowDays: Number(bookingWindowDays),
				inviteeLimit: inviteeLimit === '' ? null : Number(inviteeLimit),
				timezone,
				requiredHostEmails,
				optionalHostEmails,
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
	<div class="rounded-lg border-2 p-8" style={blockStyle}>
		<p class="text-sm" style="color: rgb(var(--color-text) / 0.65);">Loading event type…</p>
	</div>
{:else}
	<form class="flex flex-col gap-4" onsubmit={save}>
		<div class="rounded-lg border-2 p-4 sm:p-5" style={blockStyle}>
			<div class="flex min-w-0 items-center gap-4">
				<div
					class="grid size-12 shrink-0 place-items-center rounded-lg"
					style="background: rgb(var(--color-primary) / 0.12); color: rgb(var(--color-primary));"
				>
					<Icon icon={calendarEventIcon} width="24" height="24" />
				</div>
				<div>
					<h1 class="text-2xl font-semibold tracking-tight" style="color: rgb(var(--color-text));">{slug ? 'Edit event type' : 'New event type'}</h1>
					<p class="text-sm" style="color: rgb(var(--color-text) / 0.65);">Configure the booking duration, recipients, and availability.</p>
				</div>
			</div>
		</div>

		<div class="rounded-[0.625rem] border-2" style={eventDetailsContainerStyle}>
			<section class="ml-1 grid gap-5 rounded-md rounded-l-lg p-4 sm:grid-cols-2 sm:p-5" style="background: rgb(var(--color-foreground));" aria-labelledby="event-details-title">
				<div class="sm:col-span-2">
					<h2 id="event-details-title" class="font-semibold" style="color: rgb(var(--color-text));">Event details</h2>
					<p class="mt-1 text-sm" style="color: rgb(var(--color-text) / 0.65);">Set the booking length, availability window, and schedule timezone.</p>
				</div>

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
				<NumberInput id="booking-window" label="How many calendar days ahead can invitees book?" bind:value={bookingWindowDays} required />
				<NumberInput id="invitee-limit" label="Invitees limit (empty means one booking)" bind:value={inviteeLimit} placeholder="One booking" />
				<Input id="timezone" label="Schedule timezone (read-only)" value={`${timezone} — ${currentTime}`} readonly />
			</section>
		</div>

		<HostSelector
			{users}
			bind:required={requiredHostEmails}
			bind:optional={optionalHostEmails}
			error={hostsError}
			onchange={() => (hostsError = '')}
		/>
		<ScheduleEditor bind:schedule />

		<div class="flex flex-wrap items-center gap-3">
			<Button variant="secondary" onclick={() => void goto(appPath('/scheduling'))}>
				<span class="flex items-center gap-2">
					<Icon icon={xIcon} width="18" height="18" class="cancel-event-type-icon shrink-0" />
					Cancel
				</span>
			</Button>
			<Button type="submit" disabled={saving}>
				<span class="flex items-center gap-2">
					<Icon icon={checkIcon} width="18" height="18" class="save-event-type-icon shrink-0" />
					{saving ? 'Saving…' : 'Save event type'}
				</span>
			</Button>
		</div>
	</form>
{/if}

<style>
	:global(.cancel-event-type-icon path),
	:global(.save-event-type-icon path) {
		stroke-width: 3;
	}
</style>
