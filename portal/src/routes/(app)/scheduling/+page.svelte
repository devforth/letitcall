<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import calendarEventIcon from '@iconify-icons/tabler/calendar-event';
	import dotsVerticalIcon from '@iconify-icons/tabler/dots-vertical';
	import editIcon from '@iconify-icons/mdi/edit';
	import externalLinkIcon from '@iconify-icons/charm/link-external';
	import plusIcon from '@iconify-icons/tabler/plus';
	import trashIcon from '@iconify-icons/tabler/trash';
	import { appPath, callApi } from '$lib/api';
	import EventTypeEditor from '$lib/components/EventTypeEditor.svelte';
	import HostBadges from '$lib/components/HostBadges.svelte';
	import ConfirmationDialog from '$lib/components/ui/ConfirmationDialog.svelte';
	import IconButton from '$lib/components/ui/IconButton.svelte';
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
		'background: rgb(var(--color-foreground)); box-shadow: 0 0 0 1px rgb(var(--color-border)), var(--shadow-small);';

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

<section aria-labelledby="scheduling-title" class="flex flex-col gap-4">
	<div class="rounded-lg p-4 sm:p-5" style={blockStyle}>
		<div class="flex flex-wrap items-center justify-between gap-5">
			<div class="flex min-w-0 items-center gap-4">
				<div
					class="grid size-12 shrink-0 place-items-center rounded-lg"
					style="background: rgb(var(--color-primary) / 0.12); color: rgb(var(--color-primary));"
				>
					<Icon icon={calendarEventIcon} width="24" height="24" />
				</div>
				<div>
					<div class="flex items-center gap-3">
						<h1 id="scheduling-title" class="text-2xl font-semibold tracking-tight" style="color: rgb(var(--color-text));">Scheduling</h1>
						<span class="inline-flex items-center gap-1.5 rounded-md px-2 py-1 text-xs font-semibold" style="background: rgb(var(--color-primary) / 0.1); color: rgb(var(--color-primary));">
							{loading ? 'Loading…' : `${eventTypes.length} ${eventTypes.length === 1 ? 'event type' : 'event types'}`}
						</span>
					</div>
					<p class="text-sm" style="color: rgb(var(--color-text) / 0.65);">Manage shared event types and their booking availability.</p>
				</div>
			</div>
			{#if !showForm}
				<Button onclick={() => (showForm = true)}>
					<span class="flex items-center gap-2">
						<Icon icon={plusIcon} width="18" height="18" class="add-event-type-plus shrink-0" />
						Add event type
					</span>
				</Button>
			{/if}
		</div>
	</div>

	{#if showForm}
		<EventTypeEditor embedded oncancel={() => (showForm = false)} oncreate={addEventType} />
	{/if}

	<div class="overflow-hidden rounded-lg" style={blockStyle}>
		<div
			class="p-3 sm:p-4"
			style="background: rgb(var(--color-text) / 0.06); box-shadow: inset 0 -1px rgb(var(--color-border));"
		>
			<h2 class="font-semibold" style="color: rgb(var(--color-text));">Event types</h2>
		</div>

		{#if loading}
			<p class="p-8 text-sm" style="color: rgb(var(--color-text) / 0.65);">Loading event types…</p>
		{:else}
			<div class="event-type-list">
				{#each eventTypes as eventType (eventType.eventSlug)}
					<article class="event-type-row grid gap-4 p-4 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center sm:p-5">
						<div class="flex min-w-0 items-start gap-3">
							<div class="event-type-icon grid size-10 shrink-0 place-items-center rounded-xl" aria-hidden="true">
								<Icon icon={calendarEventIcon} width="20" height="20" />
							</div>
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
						</div>
						<div class="action-slot">
							<span class="event-action-hint" aria-hidden="true">
								<Icon icon={dotsVerticalIcon} width="22" height="22" />
							</span>
							<div class="event-actions flex gap-2">
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
						</div>
					</article>
				{:else}
					<div class="px-5 py-14 text-center">
						<div class="mx-auto flex max-w-xs flex-col items-center">
							<Icon icon={calendarEventIcon} width="30" height="30" style="color: rgb(var(--color-text) / 0.65);" />
							<p class="mt-3 font-semibold" style="color: rgb(var(--color-text));">No event types yet</p>
							<p class="mt-1 text-xs" style="color: rgb(var(--color-text) / 0.65);">Create an event type to start accepting bookings.</p>
						</div>
					</div>
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
	.event-type-row {
		border-bottom: 1px solid rgb(var(--color-border));
		transition: background 0.15s ease;
	}

	.event-type-row:last-child {
		border-bottom: 0;
	}

	.event-type-row:hover {
		background: rgb(var(--color-primary) / 0.045);
	}

	.event-type-icon {
		background: rgb(var(--color-primary) / 0.14);
		color: rgb(var(--color-primary));
	}

	.duration-chip {
		border: 1px solid rgb(var(--color-border));
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
		background: rgb(var(--color-text) / 0.08);
		color: rgb(var(--color-text) / 0.65);
		transition: background 0.15s ease, color 0.15s ease;
	}

	.event-icon-link:hover {
		background: rgb(var(--color-text) / 0.14);
		color: rgb(var(--color-text));
	}

	.event-icon-link:focus-visible {
		outline: 2px solid rgb(var(--color-text) / 0.65);
		outline-offset: 2px;
	}

	.action-slot {
		position: relative;
		display: flex;
		min-height: 2.5rem;
		align-items: center;
		justify-content: flex-end;
	}

	.event-action-hint {
		position: absolute;
		right: 0;
		display: grid;
		width: 2.5rem;
		height: 2.5rem;
		place-items: center;
		color: rgb(var(--color-text) / 0.6);
		pointer-events: none;
		transition: opacity 0.18s ease, transform 0.18s ease;
	}

	.event-actions {
		opacity: 0;
		pointer-events: none;
		transform: translateX(0.5rem);
		transition: opacity 0.18s ease, transform 0.18s ease;
	}

	.event-type-row:hover .event-actions,
	.event-type-row:focus-within .event-actions {
		opacity: 1;
		pointer-events: auto;
		transform: translateX(0);
	}

	.event-type-row:hover .event-action-hint,
	.event-type-row:focus-within .event-action-hint {
		opacity: 0;
		transform: translateX(-0.5rem) scale(0.85);
	}

	:global(.add-event-type-plus path) {
		stroke-width: 3;
	}

	@media (hover: none) {
		.event-actions {
			opacity: 1;
			pointer-events: auto;
			transform: none;
		}

		.event-action-hint {
			display: none;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.event-type-row,
		.event-action-hint,
		.event-actions {
			transition: none;
		}
	}
</style>
