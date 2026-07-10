<script lang="ts">
	import type { ManagedUser } from '$lib/types';
	import Button from '$lib/components/ui/Button.svelte';

	let {
		users,
		currentEmail,
		deletingEmail = '',
		ondelete
	}: {
		users: ManagedUser[];
		currentEmail: string;
		deletingEmail?: string;
		ondelete: (email: string) => void;
	} = $props();
</script>

<div class="overflow-x-auto border border-black">
	<table class="w-full min-w-[44rem] border-collapse text-left text-sm">
		<thead>
			<tr class="border-b border-black">
				<th class="px-4 py-3 font-semibold">Email</th>
				<th class="px-4 py-3 font-semibold">Is Google connected</th>
				<th class="px-4 py-3 font-semibold">Timezone</th>
				<th class="px-4 py-3 text-right font-semibold"><span class="sr-only">Actions</span></th>
			</tr>
		</thead>
		<tbody>
			{#each users as user (user.email)}
				<tr class="border-b border-black last:border-b-0">
					<td class="px-4 py-3 font-medium">{user.email}</td>
					<td class="px-4 py-3">{user.googleConnected ? 'Yes' : 'No'}</td>
					<td class="px-4 py-3">{user.timezone}</td>
					<td class="px-4 py-3 text-right">
						<Button
							variant="danger"
							disabled={user.email === currentEmail || deletingEmail === user.email}
							onclick={() => ondelete(user.email)}
						>
							{deletingEmail === user.email ? 'Deleting…' : 'Delete'}
						</Button>
					</td>
				</tr>
			{:else}
				<tr>
					<td class="px-4 py-8 text-center" colspan="4">No users found.</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>
