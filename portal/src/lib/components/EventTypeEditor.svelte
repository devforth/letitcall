<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import calendarEventIcon from '@iconify-icons/tabler/calendar-event';
	import calendarPlusIcon from '@iconify-icons/tabler/calendar-plus';
	import checkIcon from '@iconify-icons/tabler/check';
	import clockIcon from '@iconify-icons/tabler/clock';
	import usersIcon from '@iconify-icons/tabler/users';
	import xIcon from '@iconify-icons/tabler/x';
	import refreshIcon from '@iconify-icons/tabler/refresh';
	import { appPath, callApi } from '$lib/api';
	import LeaveGuard from '$lib/components/LeaveGuard.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import NumberInput from '$lib/components/ui/NumberInput.svelte';
	import SearchableSelect from '$lib/components/ui/SearchableSelect.svelte';
	import ScheduleEditor from '$lib/components/ScheduleEditor.svelte';
	import StickyActions from '$lib/components/StickyActions.svelte';
	import HostSelector from '$lib/components/HostSelector.svelte';
	import type { EventType, ManagedUser, ScheduleDay } from '$lib/types';

	let {
		slug = ''
	}: {
		slug?: string;
	} = $props();

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
	const requestBody = $derived({
		name,
		durationMinutes: Number(durationChoice === 'custom' ? customDuration : durationChoice),
		bookingWindowDays: Number(bookingWindowDays),
		inviteeLimit: inviteeLimit === '' ? null : Number(inviteeLimit),
		timezone,
		requiredHostEmails,
		optionalHostEmails,
		schedule: schedule.map((day) =>
			day.enabled ? day : { day: day.day, enabled: false, start: '', end: '', breaks: [] }
		)
	});
	let savedEventType = $state<EventType>();
	let savedBody = $state('');
	const changed = $derived(JSON.stringify(requestBody) !== savedBody);
	const outlinedBlockStyle =
		'background: rgb(var(--color-background)); box-shadow: 0 0 0 1px var(--color-border);';
	let clock: number | undefined;
	let leaveGuard: LeaveGuard;

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
				savedEventType = response.eventType;
				applyEventType(response.eventType);
			} else {
				timezone = Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC';
				requiredHostEmails = [sessionResponse.user.email];
			}
			savedBody = JSON.stringify(requestBody);
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
			if (slug) {
				await callApi(`/api/event-types/${encodeURIComponent(slug)}`, {
					method: 'PUT',
					body: JSON.stringify(requestBody)
				});
				await leaveGuard.leave(appPath('/scheduling'));
			} else {
				await callApi('/api/event-types', {
					method: 'POST',
					body: JSON.stringify(requestBody)
				});
				await leaveGuard.leave(appPath('/scheduling'));
			}
		} catch {
			// callApi reports the error globally.
		} finally {
			saving = false;
		}
	}

	function cancel() {
		void leaveGuard.leave(appPath('/scheduling'));
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
	<div class="rounded-lg p-8" style={outlinedBlockStyle}>
		<p class="text-sm" style="color: rgb(var(--color-text) / 0.65);">Loading event type…</p>
	</div>
{:else}
	<div class="mb-8">
		<PageHeader
			id="event-type-title"
			title={slug ? 'Edit event type' : 'New event type'}
			description="Configure the booking duration, recipients and availability."
			icon={slug ? calendarEventIcon : calendarPlusIcon}
			parent={{ href: appPath('/scheduling'), label: 'Scheduling' }}
		/>
	</div>
	<form onsubmit={save}>
		<div class="flex flex-col rounded-md" style="background: rgb(var(--color-background));">
			<section
				class="grid gap-5 sm:grid-cols-2"
				aria-labelledby="event-details-title"
			>
				<div class="-mb-2 sm:col-span-2">
					<div>
						<h2 id="event-details-title" class="font-semibold" style="color: rgb(var(--color-text));">Event details</h2>
						<p class="mt-1 text-sm" style="color: rgb(var(--color-text) / 0.65);">Set the booking length, availability window and schedule timezone</p>
					</div>
				</div>

				<Input id="event-name" label="Event Name" bind:value={name} required />
				{#if slug}
					<Input id="event-slug" label="Event alias" value={eventSlug} readonly />
				{/if}
				<SearchableSelect
					id="duration"
					label="Duration"
					bind:value={durationChoice}
					icon={clockIcon}
					required
					options={[
						{ value: '15', label: '15 minutes' },
						{ value: '30', label: '30 minutes' },
						{ value: '45', label: '45 minutes' },
						{ value: '60', label: '1 hour' },
						{ value: 'custom', label: 'Custom' }
					]}
				/>
				{#if durationChoice === 'custom'}
					<NumberInput id="custom-duration" label="Custom duration in minutes" bind:value={customDuration} max={1440} icon={clockIcon} required />
				{/if}
				<NumberInput id="booking-window" label="How many calendar days ahead can invitees book" bind:value={bookingWindowDays} icon={calendarEventIcon} required />
				<NumberInput id="invitee-limit" label="Invitees limit (empty means one booking)" bind:value={inviteeLimit} placeholder="One booking" icon={usersIcon} />
				<Input id="timezone" label="Schedule timezone (read-only)" value={`${timezone} — ${currentTime}`} readonly />
			</section>

		<div class="mt-8">
			<HostSelector
				{users}
				bind:required={requiredHostEmails}
				bind:optional={optionalHostEmails}
				error={hostsError}
				onchange={() => (hostsError = '')}
			/>
		</div>
		<ScheduleEditor bind:schedule stepMinutes={requestBody.durationMinutes} />

		{#if !slug || changed}
			<StickyActions class="mt-3">
				<Button
					variant="primary-outline"
					class="outlined-action-button"
					disabled={saving}
					onclick={slug ? () => applyEventType(savedEventType!) : cancel}
				>
					<span class="flex items-center gap-2">
						<Icon icon={slug ? refreshIcon : xIcon} width="20" height="20" />
						{slug ? 'Revert' : 'Cancel'}
					</span>
				</Button>
				<Button
					type="submit"
					rounded
					class="primary-action-button"
					disabled={saving}
				>
					<span class="flex items-center gap-2">
						<Icon icon={checkIcon} width="20" height="20" />
						{saving ? (slug ? 'Saving…' : 'Creating…') : (slug ? 'Save' : 'Create')}
					</span>
				</Button>
			</StickyActions>
		{/if}
		</div>
	</form>
{/if}

<LeaveGuard
	bind:this={leaveGuard}
	{changed}
	title={slug ? 'Leave without saving?' : 'Leave without creating?'}
	description={slug
		? 'Your changes to this event type have not been saved and will be lost.'
		: 'This event type has not been created yet. Everything you entered will be lost.'}
/>
