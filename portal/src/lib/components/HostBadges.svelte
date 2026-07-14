<script lang="ts">
	import { avatarURL } from '$lib/api';
	import type { ManagedUser } from '$lib/types';

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
		<span class="inline-flex items-center gap-2 border border-black px-2 py-1 text-xs">
			{#if recipient?.avatarPath}
				<img src={avatarURL(recipient.avatarPath)} alt="" class="size-6 border border-black object-cover" />
			{/if}
			{host.email} · {host.role}
		</span>
	{/each}
</div>
