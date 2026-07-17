<script lang="ts">
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
		<span class="host-badge inline-flex items-center gap-2 rounded-md px-2 py-1 text-xs">
			<Avatar name={recipient?.fullName} email={host.email} avatarPath={recipient?.avatarPath} size={24} rounded="md" />
			{host.email} · {host.role}
		</span>
	{/each}
</div>

<style>
	.host-badge {
		border: 1px solid rgb(var(--color-border));
		background: rgb(var(--color-text) / 0.06);
		color: rgb(var(--color-text) / 0.75);
	}
</style>
