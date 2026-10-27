<script lang="ts">
	import { onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import checkIcon from '@iconify-icons/tabler/check';
	import clockIcon from '@iconify-icons/tabler/clock';
	import clockPlusIcon from '@iconify-icons/tabler/clock-plus';
	import userIcon from '@iconify-icons/tabler/user';
	import xIcon from '@iconify-icons/tabler/x';
	import { appPath, callApi } from '$lib/api';
	import { timezoneDateKey } from '$lib/booking';
	import type { ManagedUser } from '$lib/types';
	import LeaveGuard from '$lib/components/LeaveGuard.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import StickyActions from '$lib/components/StickyActions.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import MonthCalendar from '$lib/components/ui/MonthCalendar.svelte';
	import NumberInput from '$lib/components/ui/NumberInput.svelte';
	import SearchableSelect from '$lib/components/ui/SearchableSelect.svelte';
	import Textarea from '$lib/components/ui/Textarea.svelte';
	import TimeInput from '$lib/components/ui/TimeInput.svelte';

	const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
	const today = timezoneDateKey(new Date(), timezone);
	let users = $state<ManagedUser[]>([]);
	let userEmail = $state('');
	let month = $state(today.slice(0, 7));
	let date = $state(today);
	let time = $state('09:00');
	let durationMinutes = $state('60');
	let comment = $state('');
	let savedBody = $state('');
	let saving = $state(false);
	let leaveGuard: LeaveGuard;

	const availableDates = $derived(
		Array.from(
			{ length: new Date(Date.UTC(Number(month.slice(0, 4)), Number(month.slice(5, 7)), 0)).getUTCDate() },
			(_, index) => `${month}-${String(index + 1).padStart(2, '0')}`
		).filter((candidate) => candidate >= today)
	);
	const fields = $derived({ userEmail, date, time, durationMinutes, comment });
	const changed = $derived(JSON.stringify(fields) !== savedBody);

	onMount(async () => {
		try {
			const [usersResponse, sessionResponse] = await Promise.all([
				callApi<{ users: ManagedUser[] }>('/api/users'),
				callApi<{ user: ManagedUser }>('/api/auth/session')
			]);
			users = usersResponse.users;
			userEmail = sessionResponse.user.email;
			savedBody = JSON.stringify(fields);
		} catch {
			// callApi reports the error globally.
		}
	});

	async function create(event: SubmitEvent) {
		event.preventDefault();
		saving = true;
		try {
			await callApi('/api/busy-times', {
				method: 'POST',
				body: JSON.stringify({
					userEmail,
					// Wall time in the browser timezone, sent as a UTC instant with second precision.
					start: new Date(`${date}T${time}`).toISOString().replace('.000Z', 'Z'),
					durationMinutes: Number(durationMinutes),
					comment
				})
			});
			await leaveGuard.leave(appPath('/busy-times'));
		} catch {
			// callApi reports the error globally.
		} finally {
			saving = false;
		}
	}
</script>

<PageTitle title="New busy time" />

<section aria-labelledby="new-busy-time-title">
	<div class="mb-8">
		<PageHeader
			id="new-busy-time-title"
			title="New busy time"
			description={`Block a member's bookings. Times use your timezone, ${timezone}.`}
			icon={clockPlusIcon}
			parent={{ href: appPath('/busy-times'), label: 'Busy times' }}
		/>
	</div>

	<form class="grid grid-cols-1 gap-5 lg:grid-cols-2" onsubmit={create}>
		<div class="max-w-[500px] lg:row-span-4">
			<MonthCalendar bind:month bind:selected={date} {availableDates} minimumMonth={today.slice(0, 7)} {today} />
		</div>
		<SearchableSelect
			id="busy-time-user"
			label="Member"
			icon={userIcon}
			options={users.map((user) => ({ value: user.email, label: user.fullName ? `${user.fullName} · ${user.email}` : user.email }))}
			bind:value={userEmail}
			clearable={false}
			required
		/>
		<TimeInput id="busy-time-start" label="Start time" bind:value={time} step={60} required />
		<NumberInput id="busy-time-duration" label="Duration in minutes" bind:value={durationMinutes} icon={clockIcon} required />
		<Textarea id="busy-time-comment" label="Comment (optional)" bind:value={comment} maxlength={2000} rows={3} />
		<StickyActions class="lg:col-span-2">
			<Button variant="primary-outline" class="outlined-action-button" onclick={() => leaveGuard.leave(appPath('/busy-times'))}>
				<span class="flex items-center gap-2">
					<Icon icon={xIcon} width="20" height="20" />
					Cancel
				</span>
			</Button>
			<Button type="submit" rounded class="primary-action-button" disabled={saving}>
				<span class="flex items-center gap-2">
					<Icon icon={checkIcon} width="20" height="20" />
					{saving ? 'Creating…' : 'Create'}
				</span>
			</Button>
		</StickyActions>
	</form>
</section>

<LeaveGuard
	bind:this={leaveGuard}
	{changed}
	title="Leave without creating?"
	description="This busy time has not been created yet. Everything you entered will be lost."
/>
