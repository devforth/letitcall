<script lang="ts">
	import { onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import externalLinkIcon from '@iconify-icons/tabler/external-link';
	import { appPath, avatarURL, callApi } from '$lib/api';
	import EventTypeActionsMenu from '$lib/components/EventTypeActionsMenu.svelte';
	import ConfirmationDialog from '$lib/components/ui/ConfirmationDialog.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import type { EventType, ManagedUser } from '$lib/types';

	let eventTypes = $state<EventType[]>([]);
	let users = $state<ManagedUser[]>([]);
	let loading = $state(true);
	let deletingSlug = $state('');
	let eventTypeToDelete = $state<EventType | null>(null);

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

	function user(email: string) {
		return users.find((candidate) => candidate.email === email);
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

<section aria-labelledby="scheduling-title">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 id="scheduling-title" class="text-2xl font-semibold tracking-tight">Scheduling</h1>
			<p class="mt-2 text-sm">Manage shared event types and their booking availability.</p>
		</div>
		<a class="border border-black bg-black px-4 py-3 text-sm font-medium text-white" href={appPath('/scheduling/new')}
			>Add new event type</a
		>
	</div>

	{#if loading}
		<p class="mt-6 border border-black p-6 text-sm">Loading event types…</p>
	{:else}
		<div class="mt-6 grid gap-3">
			{#each eventTypes as eventType (eventType.eventSlug)}
				<article class="grid gap-4 border border-black p-4 md:grid-cols-[1fr_auto] md:items-center">
					<div class="min-w-0">
						<h2 class="font-semibold">{eventType.name}</h2>
						<p class="mt-1 text-xs">{eventType.durationMinutes} minutes</p>
						<div class="mt-3 flex flex-wrap items-center gap-2">
							{#each eventType.requiredHostEmails as email (email)}
								{@const recipient = user(email)}
								<span class="inline-flex items-center gap-2 border border-black px-2 py-1 text-xs">
									{#if recipient?.avatarPath}
										<img src={avatarURL(recipient.avatarPath)} alt="" class="size-6 border border-black object-cover" />
									{/if}
									{email} · Required
								</span>
							{/each}
							{#each eventType.optionalHostEmails as email (email)}
								{@const recipient = user(email)}
								<span class="inline-flex items-center gap-2 border border-black px-2 py-1 text-xs">
									{#if recipient?.avatarPath}
										<img src={avatarURL(recipient.avatarPath)} alt="" class="size-6 border border-black object-cover" />
									{/if}
									{email} · Optional
								</span>
							{/each}
						</div>
					</div>
					<div class="flex gap-2">
						<a
							class="grid size-11 place-items-center border border-black"
							href={appPath(`/book/${eventType.eventSlug}`)}
							title="Open booking page"
							aria-label={`Open booking page for ${eventType.name}`}
						>
							<Icon icon={externalLinkIcon} width="20" height="20" />
						</a>
						<a class="border border-black px-4 py-3 text-center text-sm font-medium" href={appPath(`/scheduling/${eventType.eventSlug}`)}>Edit</a>
						<EventTypeActionsMenu
							name={eventType.name}
							deleting={deletingSlug === eventType.eventSlug}
							ondelete={() => (eventTypeToDelete = eventType)}
						/>
					</div>
				</article>
			{:else}
				<p class="border border-black p-6 text-sm">No event types yet.</p>
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
