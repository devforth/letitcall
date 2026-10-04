<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import calendarCogIcon from '@iconify-icons/tabler/calendar-cog';
	import calendarPlusIcon from '@iconify-icons/tabler/calendar-plus';
	import editIcon from '@iconify-icons/mdi/edit';
	import externalLinkIcon from '@iconify-icons/charm/link-external';
	import durationIcon from '@iconify-icons/cuida/clock-outline';
	import calendarEventIcon from '@iconify-icons/tabler/calendar-event';
	import trashIcon from '@iconify-icons/tabler/trash';
	import { appPath, callApi } from '$lib/api';
	import { availabilityRanges, formatWallTime } from '$lib/schedule';
	import HostBadges from '$lib/components/HostBadges.svelte';
	import ConfirmationDialog from '$lib/components/ui/ConfirmationDialog.svelte';
	import IconButton from '$lib/components/ui/IconButton.svelte';
	import TimedActions from '$lib/components/ui/TimedActions.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import type { EventType, ManagedUser } from '$lib/types';

	let eventTypes = $state<EventType[]>([]);
	let users = $state<ManagedUser[]>([]);
	let loading = $state(true);
	let deletingSlug = $state('');
	let eventTypeToDelete = $state<EventType | null>(null);

	const blockStyle =
		'background: rgb(var(--color-background)); box-shadow: 0 0 0 1px var(--color-border);';
	const boldCalendarPlusIcon = {
		...calendarPlusIcon,
		body: calendarPlusIcon.body.replace('stroke-width="2"', 'stroke-width="2.25"')
	};

	onMount(async () => {
		try {
			const [eventTypesResponse, usersResponse] = await Promise.all([
				callApi<{ eventTypes: EventType[] }>('/api/event-types'),
				callApi<{ users: ManagedUser[] }>('/api/users')
			]);
			eventTypes = eventTypesResponse.eventTypes;
			users = usersResponse.users;
		} catch {
			// callApi reports the error globally.
		} finally {
			loading = false;
		}
	});

	function hosts(eventType: EventType) {
		return [
			...eventType.requiredHostEmails.map((email) => ({ email, role: 'Required' as const })),
			...eventType.optionalHostEmails.map((email) => ({ email, role: 'Optional' as const }))
		];
	}

	async function deleteEventType() {
		const eventType = eventTypeToDelete!;
		deletingSlug = eventType.eventSlug;
		try {
			await callApi(`/api/event-types/${encodeURIComponent(eventType.eventSlug)}`, { method: 'DELETE' });
			eventTypes = eventTypes.filter((candidate) => candidate.eventSlug !== eventType.eventSlug);
			eventTypeToDelete = null;
		} catch {
			// callApi reports the error globally.
		} finally {
			deletingSlug = '';
		}
	}
</script>

<PageTitle title="Scheduling" />

<section aria-labelledby="scheduling-title" class="flex flex-col gap-6">
	<div class="mb-2">
		<PageHeader
			id="scheduling-title"
			title="Scheduling"
			description="Manage shared event types and their booking availability"
			icon={calendarCogIcon}
			count={eventTypes.length}
		>
			<Button rounded class="primary-action-button self-start" onclick={() => void goto(appPath('/scheduling/new'))}>
				<span class="flex items-center gap-2">
					<Icon icon={boldCalendarPlusIcon} width="20" height="20" class="shrink-0" />
					Add event type
				</span>
			</Button>
		</PageHeader>
	</div>

	{#if loading}
		<p class="p-8 text-sm" style="color: rgb(var(--color-text) / 0.65);">Loading event types…</p>
	{:else}
		<div class="grid gap-5">
			{#each eventTypes as eventType (eventType.eventSlug)}
				<article data-timed-actions-row class="relative grid gap-4 rounded-lg p-4 sm:p-5" style={blockStyle}>
					<div class="min-w-0">
						<div class="sm:pr-36">
							<p class="truncate text-xs leading-none" style="color: rgb(var(--color-text));">/{eventType.eventSlug}</p>
							<h3 class="text-xl leading-tight break-words" style="color: rgb(var(--color-text));">{eventType.name}</h3>
						</div>
						<div class="mt-3">
							<HostBadges hosts={hosts(eventType)} {users} />
						</div>
						<div class="mt-3 flex flex-wrap items-center gap-3">
							<span class="event-meta"><Icon icon={durationIcon} width="16" height="16" />{eventType.durationMinutes}:00</span>
							<span class="event-meta"><Icon icon={calendarEventIcon} width="16" height="16" />{eventType.bookingWindowDays} days ahead</span>
						</div>
						<div class="day-list mt-3">
							{#each eventType.schedule.filter((day) => day.enabled) as day (day.day)}
								<div>
									<span class="capitalize">{day.day.slice(0, 3)}</span>
									{#each availabilityRanges(day) as range, index (index)}
										<span class="day-time">{formatWallTime(range.start)} - {formatWallTime(range.end)}</span>
									{/each}
								</div>
							{/each}
						</div>
					</div>
					<div class="sm:absolute sm:right-5 sm:top-4">
						<TimedActions label={`Show actions for ${eventType.name}`} controlsId={`event-actions-${eventType.eventSlug}`} actionsVisibleOnSmallScreens>
							<div class="event-actions">
								<a
									class="event-icon-link"
									href={appPath(`/book/${eventType.eventSlug}`)}
									title="Open booking page"
									aria-label={`Open booking page for ${eventType.name}`}
								>
									<Icon icon={externalLinkIcon} width="20" height="20" />
								</a>
								<IconButton filled tone="primary" label={`Edit ${eventType.name}`} onclick={() => void goto(appPath(`/scheduling/${eventType.eventSlug}`))}>
									<Icon icon={editIcon} width="20" height="20" />
								</IconButton>
								<IconButton
									filled
									tone="danger"
									label={`Delete ${eventType.name}`}
									disabled={deletingSlug === eventType.eventSlug}
									onclick={() => (eventTypeToDelete = eventType)}
								>
									<Icon icon={trashIcon} width="20" height="20" />
								</IconButton>
							</div>
						</TimedActions>
					</div>
				</article>
			{:else}
				<p class="empty-state rounded-lg" style={blockStyle}>No event types yet</p>
			{/each}
		</div>
	{/if}
</section>

{#if eventTypeToDelete}
	<ConfirmationDialog
		open
		title="Delete event type?"
		description={`This will completely delete ${eventTypeToDelete.name}. This action cannot be undone.`}
		confirmLabel="Delete event type"
		confirmingLabel="Deleting…"
		confirming={deletingSlug === eventTypeToDelete.eventSlug}
		onconfirm={deleteEventType}
		oncancel={() => (eventTypeToDelete = null)}
	/>
{/if}

<style>
	.event-meta {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		color: rgb(var(--color-text));
		font-size: 0.875rem;
		font-weight: 500;
		white-space: nowrap;
	}

	.day-list {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(8.5rem, 1fr));
		grid-auto-rows: 1fr;
		gap: 0.5rem;
		font-size: 0.75rem;
	}

	.day-list div {
		display: flex;
		flex-direction: column;
		gap: 2px;
		border-radius: 8px;
		padding: 0.375rem 0.625rem;
		box-shadow: 0 0 0 1px var(--color-border);
		color: rgb(var(--color-text));
	}

	.day-time {
		color: rgb(var(--color-text) / 0.65);
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
	}

	.event-icon-link {
		display: grid;
		width: 2.5rem;
		height: 2.5rem;
		flex-shrink: 0;
		place-items: center;
		border-radius: 10px;
		background: rgb(var(--color-primary) / 0.12);
		color: rgb(var(--color-primary));
		transition: background 0.15s ease, color 0.15s ease;
	}

	.event-icon-link:hover {
		background: rgb(var(--color-primary));
		color: rgb(var(--color-background));
	}

	.event-icon-link:focus-visible {
		outline: 2px solid rgb(var(--color-primary));
		outline-offset: 2px;
	}

	.event-actions {
		display: flex;
		gap: 0.5rem;
	}

	.empty-state {
		padding: 2.5rem 1rem;
		color: rgb(var(--color-text) / 0.65);
		font-size: 0.875rem;
		text-align: center;
	}

	:global(.add-event-type-plus path) {
		stroke-width: 3;
	}

</style>
