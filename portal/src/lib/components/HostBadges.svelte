<script lang="ts">
	import starIcon from '@iconify-icons/tabler/star-filled';
	import Icon from '@iconify/svelte';
	import type { ManagedUser } from '$lib/types';
	import Avatar from '$lib/components/ui/Avatar.svelte';

	let {
		hosts,
		users
	}: {
		hosts: { email: string; role: 'Required' | 'Optional' | 'Host' }[];
		users: ManagedUser[];
	} = $props();

	function user(email: string) {
		return users.find((candidate) => candidate.email === email);
	}
</script>

<div class="flex flex-wrap items-center gap-2">
	{#each hosts as host (host.email)}
		{@const recipient = user(host.email)}
		{@const hostName = recipient?.fullName || host.email}
		<span class="host-badge inline-flex min-w-0 items-center gap-2 rounded-full py-1 pl-1 pr-3 text-xs">
			<Avatar name={recipient?.fullName} email={host.email} avatarPath={recipient?.avatarPath} size={28} />
			<span class="min-w-0 leading-tight">
				<span class="flex min-w-0 items-baseline gap-1">
					<span class="truncate font-semibold" style="color: rgb(var(--color-text));">{hostName}</span>
					{#if host.role === 'Required'}
						<span class="host-role" aria-label="Required host" title="Required host">
							<Icon icon={starIcon} width="11" height="11" aria-hidden="true" />
						</span>
					{/if}
				</span>
				{#if recipient?.fullName}
					<span class="block truncate">{host.email}</span>
				{/if}
			</span>
		</span>
	{/each}
</div>

<style>
	.host-badge {
		background: rgb(var(--color-background));
		box-shadow: 0 0 0 1px var(--color-border);
		color: rgb(var(--color-text) / 0.75);
	}

	.host-role {
		display: inline-flex;
		align-items: center;
		flex-shrink: 0;
		color: rgb(var(--warning));
		font-weight: 600;
	}
</style>
