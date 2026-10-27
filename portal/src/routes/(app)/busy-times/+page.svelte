<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import clockIcon from '@iconify-icons/tabler/clock';
	import clockPauseIcon from '@iconify-icons/tabler/clock-pause';
	import clockPlusIcon from '@iconify-icons/tabler/clock-plus';
	import historyIcon from '@iconify-icons/tabler/history';
	import trashIcon from '@iconify-icons/tabler/trash';
	import { appPath, callApi } from '$lib/api';
	import { browserTimeFormatter } from '$lib/time-format';
	import type { BusyTime, ManagedUser } from '$lib/types';
	import HostBadges from '$lib/components/HostBadges.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import ConfirmationDialog from '$lib/components/ui/ConfirmationDialog.svelte';
	import IconButton from '$lib/components/ui/IconButton.svelte';
	import SegmentedControl from '$lib/components/ui/SegmentedControl.svelte';
	import TimedActions from '$lib/components/ui/TimedActions.svelte';

	let busyTimes = $state<BusyTime[]>([]);
	let users = $state<ManagedUser[]>([]);
	let loading = $state(true);
	let view = $state('upcoming');
	let busyTimeToDelete = $state<BusyTime | null>(null);
	let deleting = $state(false);
	const now = new Date();

	const blockStyle =
		'background: rgb(var(--color-background)); box-shadow: 0 0 0 1px var(--color-border);';
	const rangeFormat = browserTimeFormatter({ dateStyle: 'medium', timeStyle: 'short' });
	const upcoming = $derived(busyTimes.filter((busyTime) => end(busyTime) > now));
	const past = $derived(busyTimes.filter((busyTime) => end(busyTime) <= now).reverse());
	const views = $derived([
		{ value: 'upcoming', label: 'Upcoming', icon: clockIcon, suffix: `· ${upcoming.length}` },
		{ value: 'past', label: 'Past', icon: historyIcon, suffix: `· ${past.length}` }
	]);
	const visible = $derived(view === 'upcoming' ? upcoming : past);

	onMount(async () => {
		try {
			const [busyTimesResponse, usersResponse] = await Promise.all([
				callApi<{ busyTimes: BusyTime[] }>('/api/busy-times'),
				callApi<{ users: ManagedUser[] }>('/api/users')
			]);
			busyTimes = busyTimesResponse.busyTimes;
			users = usersResponse.users;
		} catch {
			// callApi reports the error globally.
		} finally {
			loading = false;
		}
	});

	function end(busyTime: BusyTime) {
		return new Date(new Date(busyTime.start).getTime() + busyTime.durationMinutes * 60_000);
	}

	function formatDuration(minutes: number) {
		const hours = Math.floor(minutes / 60);
		return [hours && `${hours} h`, minutes % 60 && `${minutes % 60} min`].filter(Boolean).join(' ');
	}

	async function deleteBusyTime() {
		const busyTime = busyTimeToDelete!;
		deleting = true;
		try {
			await callApi(`/api/busy-times/${encodeURIComponent(busyTime.id)}`, { method: 'DELETE' });
			busyTimes = busyTimes.filter((candidate) => candidate.id !== busyTime.id);
			busyTimeToDelete = null;
		} catch {
			// callApi reports the error globally.
		} finally {
			deleting = false;
		}
	}
</script>

<PageTitle title="Busy times" />

<section aria-labelledby="busy-times-title" class="flex flex-col gap-6">
	<div class="mb-2">
		<PageHeader
			id="busy-times-title"
			title="Busy times"
			description="Time when members cannot be booked. Times use your local timezone."
			icon={clockPauseIcon}
		>
			<div class="flex flex-wrap items-center gap-3">
				<SegmentedControl options={views} value={view} label="Busy time view" onchange={(value) => (view = value)} />
				<Button rounded class="primary-action-button" onclick={() => void goto(appPath('/busy-times/new'))}>
					<span class="flex items-center gap-2">
						<Icon icon={clockPlusIcon} width="20" height="20" class="shrink-0" />
						Add busy time
					</span>
				</Button>
			</div>
		</PageHeader>
	</div>

	{#if loading}
		<p class="p-8 text-sm" style="color: rgb(var(--color-text) / 0.65);">Loading busy times…</p>
	{:else}
		<div class="grid gap-5">
			{#each visible as busyTime (busyTime.id)}
				<article data-timed-actions-row class="relative grid gap-3 rounded-lg p-4 sm:p-5" style={blockStyle}>
					<div class="sm:pr-16">
						<h3 class="text-xl leading-tight" style="color: rgb(var(--color-text));">
							{rangeFormat.formatRange(new Date(busyTime.start), end(busyTime))}
						</h3>
						<p class="mt-1 text-sm" style="color: rgb(var(--color-text) / 0.65);">{formatDuration(busyTime.durationMinutes)}</p>
					</div>
					<HostBadges hosts={[{ email: busyTime.userEmail, role: 'Host' }]} {users} />
					{#if busyTime.comment}
						<p class="text-sm whitespace-pre-line" style="color: rgb(var(--color-text));">{busyTime.comment}</p>
					{/if}
					<div class="sm:absolute sm:right-5 sm:top-4">
						<TimedActions label="Show actions for this busy time" controlsId={`busy-time-actions-${busyTime.id}`} actionsVisibleOnSmallScreens>
							<IconButton filled tone="danger" label="Delete busy time" onclick={() => (busyTimeToDelete = busyTime)}>
								<Icon icon={trashIcon} width="20" height="20" />
							</IconButton>
						</TimedActions>
					</div>
				</article>
			{:else}
				<p class="rounded-lg px-4 py-10 text-center text-sm" style="{blockStyle} color: rgb(var(--color-text) / 0.65);">
					{view === 'upcoming' ? 'No upcoming busy times' : 'No past busy times'}
				</p>
			{/each}
		</div>
	{/if}
</section>

<ConfirmationDialog
	open={busyTimeToDelete !== null}
	title="Delete busy time?"
	description="The member becomes bookable again during this time."
	confirmLabel="Delete"
	confirmingLabel="Deleting…"
	confirming={deleting}
	onconfirm={deleteBusyTime}
	oncancel={() => (busyTimeToDelete = null)}
/>
