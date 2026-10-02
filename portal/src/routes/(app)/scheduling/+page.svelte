<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import calendarCogIcon from '@iconify-icons/tabler/calendar-cog';
	import calendarPlusIcon from '@iconify-icons/tabler/calendar-plus';
	import editIcon from '@iconify-icons/mdi/edit';
	import externalLinkIcon from '@iconify-icons/charm/link-external';
	import listDetailsIcon from '@iconify-icons/tabler/list-details';
	import trashIcon from '@iconify-icons/tabler/trash';
	import { appPath, callApi } from '$lib/api';
	import EventTypeEditor from '$lib/components/EventTypeEditor.svelte';
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
	let showForm = $state(false);
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

	function addEventType(eventType: EventType) {
		eventTypes = [...eventTypes, eventType].sort((a, b) => a.eventSlug.localeCompare(b.eventSlug));
		showForm = false;
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
			description="Manage shared event types and their booking availability."
			icon={calendarCogIcon}
		>
			{#if !showForm}
				<Button rounded style="font-weight: 500 !important; padding-right: 1rem !important; padding-bottom: 0.5rem !important;" class="add-event-type-button self-start" onclick={() => (showForm = true)}>
					<span class="flex items-center gap-2">
						<Icon icon={boldCalendarPlusIcon} width="18" height="18" class="shrink-0" />
						Add event type
					</span>
				</Button>
			{/if}
		</PageHeader>
	</div>

	{#if showForm}
		<EventTypeEditor embedded oncancel={() => (showForm = false)} oncreate={addEventType} />
	{/if}

	<div class="overflow-hidden rounded-lg" style={blockStyle}>
		<div
			class="flex items-center gap-2 border-b px-4 py-3"
			style="border-color: var(--color-border); background: rgb(var(--color-text) / 0.06); color: rgb(var(--color-text));"
		>
			<Icon icon={listDetailsIcon} width="18" height="18" />
			<h2 class="text-sm font-medium">
				Event types
				{#if eventTypes.length > 0}
					<span style="color: rgb(var(--color-text) / 0.65);">· {eventTypes.length}</span>
				{/if}
			</h2>
		</div>

		{#if loading}
			<p class="p-8 text-sm" style="color: rgb(var(--color-text) / 0.65);">Loading event types…</p>
		{:else}
			<div class="event-type-list">
				{#each eventTypes as eventType (eventType.eventSlug)}
					<article data-timed-actions-row class="event-type-row grid gap-4 p-4 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-start sm:p-5">
						<div class="min-w-0">
							<div class="flex flex-wrap items-center gap-2">
								<h3 class="truncate font-semibold" style="color: rgb(var(--color-text));">{eventType.name}</h3>
								<span class="duration-chip">{eventType.durationMinutes} minutes</span>
							</div>
							<p class="mt-0.5 truncate text-xs" style="color: rgb(var(--color-text) / 0.65);">/{eventType.eventSlug}</p>
							<div class="mt-3">
								<HostBadges hosts={hosts(eventType)} {users} />
							</div>
						</div>
						<TimedActions label={`Show actions for ${eventType.name}`} controlsId={`event-actions-${eventType.eventSlug}`}>
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
					</article>
				{:else}
					<p class="empty-state">No event types yet</p>
				{/each}
			</div>
		{/if}
	</div>
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
	:global(.add-event-type-button) {
		min-height: 4rem !important;
		border-radius: 9999px !important;
		font-size: 1.25rem !important;
	}

	.event-type-row {
		border-bottom: 1px solid var(--color-border);
		transition: background 0.15s ease;
	}

	.event-type-row:last-child {
		border-bottom: 0;
	}

	.event-type-row:hover {
		background: color-mix(in srgb, var(--color-border) 10%, transparent);
	}

	.duration-chip {
		border: 1px solid var(--color-border);
		border-radius: 999px;
		padding: 0.2rem 0.5rem;
		background: rgb(var(--color-text) / 0.06);
		color: rgb(var(--color-text) / 0.65);
		font-size: 0.75rem;
		font-weight: 600;
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

	@media (prefers-reduced-motion: reduce) {
		.event-type-row {
			transition: none;
		}
	}
</style>
