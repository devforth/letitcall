<script lang="ts">
	import type { ManagedUser } from '$lib/types';
	import Button from '$lib/components/ui/Button.svelte';
	import { avatarURL } from '$lib/api';

	let {
		users,
		currentEmail,
		deletingEmail = '',
		onedit,
		ondelete
	}: {
		users: ManagedUser[];
		currentEmail: string;
		deletingEmail?: string;
		onedit: (email: string) => void;
		ondelete: (email: string) => void;
	} = $props();
</script>

<div class="overflow-x-auto border border-black">
	<table class="w-full min-w-[50rem] border-collapse text-left text-sm">
		<thead>
			<tr class="border-b border-black">
				<th class="px-4 py-3 font-semibold">Avatar</th>
				<th class="px-4 py-3 font-semibold">Email</th>
				<th class="px-4 py-3 font-semibold">Full name</th>
				<th class="px-4 py-3 font-semibold">Is Google connected</th>
				<th class="px-4 py-3 font-semibold">Timezone</th>
				<th aria-label="Actions" class="px-4 py-3 text-right font-semibold"></th>
			</tr>
		</thead>
		<tbody>
			{#each users as user (user.email)}
				<tr class="border-b border-black last:border-b-0">
					<td class="px-4 py-3">
						{#if user.avatarPath}
							<img
								src={avatarURL(user.avatarPath)}
								alt=""
								class="size-11 border border-black object-cover"
							/>
						{:else}
							<span aria-label="No avatar">—</span>
						{/if}
					</td>
					<td class="px-4 py-3 font-medium">{user.email}</td>
					<td class="px-4 py-3">{user.fullName || '—'}</td>
					<td class="px-4 py-3">{user.googleConnected ? 'Yes' : 'No'}</td>
					<td class="px-4 py-3">{user.timezone}</td>
					<td class="px-4 py-3">
						<div class="flex justify-end gap-2">
							<Button variant="secondary" onclick={() => onedit(user.email)}>Edit</Button>
							<Button
								variant="danger"
								disabled={user.email === currentEmail || deletingEmail === user.email}
								onclick={() => ondelete(user.email)}
							>
								{deletingEmail === user.email ? 'Deleting…' : 'Delete'}
							</Button>
						</div>
					</td>
				</tr>
			{:else}
				<tr>
					<td class="px-4 py-8 text-center" colspan="6">No users found.</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>
