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

	function initialsFor(user: ManagedUser) {
		const parts = user.fullName?.trim().split(/\s+/).filter(Boolean) ?? [];
		if (parts.length >= 2) return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
		if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
		return user.email.slice(0, 2).toUpperCase();
	}
</script>

<div class="overflow-x-auto border border-black">
	<table class="w-full min-w-[50rem] border-collapse text-left text-sm">
		<thead>
			<tr class="border-b border-black">
				<th class="px-4 py-3 font-semibold">User</th>
				<th class="px-4 py-3 font-semibold">Email</th>
				<th class="px-4 py-3 font-semibold">
					<span class="flex items-center gap-2">
						Google
						<svg class="size-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.07 0l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" /><path d="M14 11a5 5 0 0 0-7.07 0l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" /></svg>
					</span>
				</th>
				<th class="px-4 py-3 font-semibold">Timezone</th>
				<th aria-label="Actions" class="px-4 py-3 text-right font-semibold"></th>
			</tr>
		</thead>
		<tbody>
			{#each users as user (user.email)}
				<tr class="border-b border-black last:border-b-0">
					<td class="px-4 py-3">
						<div class="flex items-center gap-3">
							{#if user.avatarPath}
								<img
									src={avatarURL(user.avatarPath)}
									alt=""
									class="size-11 shrink-0 rounded-full object-cover"
									style="border: 2px solid rgb(var(--color-border));"
								/>
							{:else}
								<span
									class="flex size-11 shrink-0 items-center justify-center rounded-full text-sm font-bold leading-none"
									style="background: rgb(var(--color-primary)); color: rgb(var(--color-contrast-text));"
									aria-label="No avatar"
								>
									{initialsFor(user)}
								</span>
							{/if}
							{#if user.fullName?.trim()}
								<span class="flex flex-col font-medium leading-tight">
									{#each user.fullName.trim().split(/\s+/) as part, i (i)}
										<span>{part}</span>
									{/each}
								</span>
							{:else}
								<span class="font-medium">—</span>
							{/if}
						</div>
					</td>
					<td class="px-4 py-3 font-medium">{user.email}</td>
					<td class="px-4 py-3">
						{#if user.googleConnected}
							<svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" style="color: rgb(var(--success));" aria-label="Connected"><path d="M5 13l4 4L19 7" /></svg>
						{:else}
							<svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" style="color: rgb(var(--error));" aria-label="Not connected"><path d="M6 6l12 12M18 6L6 18" /></svg>
						{/if}
					</td>
					<td class="px-4 py-3">{user.timezone}</td>
					<td class="px-4 py-3">
						<div class="flex justify-end gap-2">
							<Button variant="primary" onclick={() => onedit(user.email)}>
								<svg class="size-[18px] shrink-0" viewBox="0 0 24 24" fill="currentColor"><path d="M13.5 6.5 17.5 10.5 8 20H4v-4l9.5-9.5z" /><path d="M15 5 19 9l1.6-1.6a2 2 0 0 0 0-2.8l-1.2-1.2a2 2 0 0 0-2.8 0L15 5z" /></svg>
								<span class="sr-only">Edit {user.email}</span>
							</Button>
							<Button
								variant="danger"
								disabled={user.email === currentEmail || deletingEmail === user.email}
								onclick={() => ondelete(user.email)}
							>
								<svg class="size-[18px] shrink-0" viewBox="0 0 24 24" fill="currentColor"><path d="M9 3a1 1 0 0 0-1 1v1H4.5a1 1 0 1 0 0 2H5v12a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V7h.5a1 1 0 1 0 0-2H16V4a1 1 0 0 0-1-1H9zm1 6a1 1 0 0 1 1 1v7a1 1 0 1 1-2 0v-7a1 1 0 0 1 1-1zm4 0a1 1 0 0 1 1 1v7a1 1 0 1 1-2 0v-7a1 1 0 0 1 1-1z" /></svg>
								<span class="sr-only">{deletingEmail === user.email ? 'Deleting…' : `Delete ${user.email}`}</span>
							</Button>
						</div>
					</td>
				</tr>
			{:else}
				<tr>
					<td class="px-4 py-8 text-center" colspan="5">No users found.</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>
