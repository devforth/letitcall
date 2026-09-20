<script lang="ts">
	import { goto } from '$app/navigation';
	import { onDestroy, onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import calendarEventIcon from '@iconify-icons/tabler/calendar-event';
	import calendarPlusIcon from '@iconify-icons/tabler/calendar-plus';
	import checkIcon from '@iconify-icons/tabler/check';
	import clockIcon from '@iconify-icons/tabler/clock';
	import usersIcon from '@iconify-icons/tabler/users';
	import xIcon from '@iconify-icons/tabler/x';
	import { appPath, callApi } from '$lib/api';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import NumberInput from '$lib/components/ui/NumberInput.svelte';
	import SearchableSelect from '$lib/components/ui/SearchableSelect.svelte';
	import ScheduleEditor from '$lib/components/ScheduleEditor.svelte';
	import HostSelector from '$lib/components/HostSelector.svelte';
	import type { EventType, ManagedUser, ScheduleDay } from '$lib/types';

	let {
		slug = '',
		embedded = false,
		oncancel,
		oncreate
	}: {
		slug?: string;
		embedded?: boolean;
		oncancel?: () => void;
		oncreate?: (eventType: EventType) => void;
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
	const blockStyle =
		'background: rgb(var(--color-background)); border-color: var(--color-border);';
	const outlinedBlockStyle =
		'background: rgb(var(--color-background)); box-shadow: 0 0 0 1px var(--color-border);';
	const eventDetailsContainerStyle = `${blockStyle} background: rgb(var(--color-primary));`;
	const newEventContainerStyle =
		'background: rgb(var(--color-primary)); box-shadow: 0 0 0 1px var(--color-border);';
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
				await goto(appPath('/scheduling'));
			} else {
				const response = await callApi<{ eventType: EventType }>('/api/event-types', {
					method: 'POST',
					body: JSON.stringify(body)
				});
				if (oncreate) oncreate(response.eventType);
				else await goto(appPath('/scheduling'));
			}
		} catch {
			// callApi reports the error globally.
		} finally {
			saving = false;
		}
	}

	function cancel() {
		if (oncancel) oncancel();
		else void goto(appPath('/scheduling'));
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
	{#if embedded && slug}
		<div class="mb-8">
			<PageHeader
				id="edit-event-type-title"
				title="Edit event type"
				description="Configure the booking duration, recipients and availability."
				icon={calendarEventIcon}
			/>
		</div>
	{/if}
	<form
		class={embedded ? (slug ? '' : 'overflow-hidden rounded-[0.625rem]') : 'flex flex-col gap-4'}
		style={embedded && !slug ? newEventContainerStyle : undefined}
		onsubmit={save}
	>
		<div
			class={embedded ? `${slug ? '' : 'ml-1 rounded-l-lg'} flex flex-col rounded-md` : 'contents'}
			style={embedded ? 'background: rgb(var(--color-background));' : undefined}
		>
		{#if !embedded}
			<PageHeader
				id="event-type-title"
				title={slug ? 'Edit event type' : 'New event type'}
				description="Configure the booking duration, recipients and availability."
				icon={slug ? calendarEventIcon : calendarPlusIcon}
			/>
		{/if}
		{#if embedded && !slug}
			<div class="flex min-w-0 items-center gap-2 rounded-t-md p-3 sm:p-4" style="background: linear-gradient(110deg, rgb(var(--color-primary) / 0.12), rgb(var(--color-background)) 42%); box-shadow: inset 0 -1px 0 var(--color-border);">
				<span class="grid size-8 shrink-0 place-items-center" style="color: rgb(var(--color-primary));">
					<Icon icon={slug ? calendarEventIcon : calendarPlusIcon} width="26" height="26" aria-hidden="true" />
				</span>
				<h1 class="text-xl font-semibold" style="color: rgb(var(--color-primary));">New event type</h1>
			</div>
		{/if}

		<div class={embedded ? '' : 'rounded-[0.625rem] border-2'} style={embedded ? undefined : eventDetailsContainerStyle}>
			<section
				class={`grid gap-5 sm:grid-cols-2 ${embedded && slug ? '' : 'p-4 sm:p-5'} ${embedded ? '' : 'ml-1 rounded-md rounded-l-lg'}`}
				style={`border-color: var(--color-border); ${embedded ? '' : 'background: rgb(var(--color-background));'}`}
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
				<NumberInput id="booking-window" label="How many calendar days ahead can invitees book?" bind:value={bookingWindowDays} icon={calendarEventIcon} required />
				<NumberInput id="invitee-limit" label="Invitees limit (empty means one booking)" bind:value={inviteeLimit} placeholder="One booking" icon={usersIcon} />
				<Input id="timezone" label="Schedule timezone (read-only)" value={`${timezone} — ${currentTime}`} readonly />
			</section>
		</div>

		<div class={embedded ? (slug ? 'mt-8' : 'mt-4') : ''}>
			<HostSelector
				{users}
				{embedded}
				flush={embedded && Boolean(slug)}
				bind:required={requiredHostEmails}
				bind:optional={optionalHostEmails}
				error={hostsError}
				onchange={() => (hostsError = '')}
			/>
		</div>
		<ScheduleEditor bind:schedule {embedded} flush={embedded && Boolean(slug)} />

		<div class={`flex flex-wrap items-center gap-3 ${embedded ? `justify-end ${slug ? 'mt-8' : 'p-4 sm:p-5'}` : ''}`}>
			<Button
				variant="primary-outline"
				class="outlined-action-button"
				onclick={cancel}
			>
				<span class="flex items-center gap-2">
					<Icon icon={xIcon} width="20" height="20" />
					Cancel
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
		</div>
		</div>
	</form>
{/if}
