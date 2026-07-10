<script lang="ts">
	import { onMount } from 'svelte';
	import { appPath, avatarURL, callApi } from '$lib/api';
	import type { EventType, ManagedUser } from '$lib/types';

	let eventTypes = $state<EventType[]>([]);
	let users = $state<ManagedUser[]>([]);
	let loading = $state(true);

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
</script>

<svelte:head><title>Scheduling · Let It Call</title></svelte:head>

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
						<p class="mt-1 text-xs">/book/{eventType.eventSlug} · {eventType.durationMinutes} minutes</p>
						<div class="mt-3 flex flex-wrap items-center gap-2">
							{#each eventType.recipientEmails as email (email)}
								{@const recipient = user(email)}
								<span class="inline-flex items-center gap-2 border border-black px-2 py-1 text-xs">
									{#if recipient?.avatarPath}
										<img src={avatarURL(recipient.avatarPath)} alt="" class="size-6 border border-black object-cover" />
									{/if}
									{email}
								</span>
							{/each}
						</div>
					</div>
					<a class="border border-black px-4 py-3 text-center text-sm font-medium" href={appPath(`/scheduling/${eventType.eventSlug}`)}>Edit</a>
				</article>
			{:else}
				<p class="border border-black p-6 text-sm">No event types yet.</p>
			{/each}
		</div>
	{/if}
</section>
