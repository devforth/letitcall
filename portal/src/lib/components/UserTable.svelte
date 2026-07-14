<script lang="ts">
	import Icon from '@iconify/svelte';
	import checkIcon from '@iconify-icons/tabler/circle-check-filled';
	import editIcon from '@iconify-icons/tabler/edit';
	import trashIcon from '@iconify-icons/tabler/trash';
	import usersIcon from '@iconify-icons/tabler/users';
	import worldIcon from '@iconify-icons/tabler/world';
	import xIcon from '@iconify-icons/tabler/circle-x-filled';
	import type { ManagedUser } from '$lib/types';
	import Button from '$lib/components/ui/Button.svelte';
	import { avatarURL } from '$lib/api';

	let {
		users,
		currentEmail,
		checkingEmail = '',
		deletingEmail = '',
		onedit,
		ondelete
	}: {
		users: ManagedUser[];
		currentEmail: string;
		checkingEmail?: string;
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

<div class="overflow-x-auto">
	<table class="user-table w-full min-w-[42rem] text-left text-sm">
		<thead>
			<tr>
				<th class="px-5 py-3.5 font-semibold">User</th>
				<th class="px-4 py-3.5 font-semibold">Calendar</th>
				<th class="px-4 py-3.5 font-semibold">Timezone</th>
				<th aria-label="Actions" class="px-5 py-3.5 text-right font-semibold"></th>
			</tr>
		</thead>
		<tbody>
			{#each users as user (user.email)}
				<tr class:current-user={user.email === currentEmail}>
					<td class="px-5 py-4">
						<div class="flex min-w-0 items-center gap-3">
							{#if user.avatarPath}
								<img
									src={avatarURL(user.avatarPath)}
									alt=""
									class="size-10 shrink-0 rounded-xl object-cover"
									style="border: 2px solid rgb(var(--color-border));"
								/>
							{:else}
								<span
									class="flex size-10 shrink-0 items-center justify-center rounded-xl text-sm font-bold leading-none"
									style="background: rgb(var(--color-primary) / 0.14); color: rgb(var(--color-primary));"
									aria-label="No avatar"
								>
									{initialsFor(user)}
								</span>
							{/if}
							<div class="min-w-0">
								<div class="flex items-center gap-2">
									<p class="truncate font-semibold" style="color: rgb(var(--color-text));">{user.fullName?.trim() || 'Unnamed user'}</p>
									{#if user.email === currentEmail}
										<span class="current-badge">You</span>
									{/if}
								</div>
								<p class="mt-0.5 truncate text-xs" style="color: rgb(var(--color-muted-foreground));">{user.email}</p>
							</div>
						</div>
					</td>
					<td class="px-4 py-4">
						{#if user.googleConnected}
							<span class="connection-state is-connected">
								<Icon icon={checkIcon} width="16" height="16" class="shrink-0" />
								Connected
							</span>
						{:else}
							<span class="connection-state">
								<Icon icon={xIcon} width="16" height="16" class="shrink-0" />
								Not connected
							</span>
						{/if}
					</td>
					<td class="px-4 py-4">
						<span class="timezone-chip">
							<Icon icon={worldIcon} width="15" height="15" class="shrink-0" />
							{user.timezone}
						</span>
					</td>
					<td class="px-5 py-4">
						<div class="flex justify-end gap-2">
							<Button class="size-9 !min-h-0 !p-0" variant="primary" onclick={() => onedit(user.email)}>
								<Icon icon={editIcon} width="17" height="17" />
								<span class="sr-only">Edit {user.email}</span>
							</Button>
							<Button
								class="size-9 !min-h-0 !p-0"
								variant="secondary"
								disabled={user.email === currentEmail || checkingEmail === user.email || deletingEmail === user.email}
								onclick={() => ondelete(user.email)}
							>
								<Icon icon={trashIcon} width="17" height="17" style="color: rgb(var(--color-primary));" />
								<span class="sr-only">{checkingEmail === user.email ? 'Checking…' : deletingEmail === user.email ? 'Deleting…' : `Delete ${user.email}`}</span>
							</Button>
						</div>
					</td>
				</tr>
			{:else}
				<tr>
					<td class="px-5 py-14 text-center" colspan="4">
						<div class="mx-auto flex max-w-xs flex-col items-center">
							<Icon icon={usersIcon} width="30" height="30" style="color: rgb(var(--color-muted-foreground));" />
							<p class="mt-3 font-semibold" style="color: rgb(var(--color-text));">No users found</p>
							<p class="mt-1 text-xs" style="color: rgb(var(--color-muted-foreground));">Try a different search or connection filter.</p>
						</div>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>

<style>
	.user-table {
		border-collapse: collapse;
	}

	.user-table thead {
		background: rgb(var(--color-muted-background));
	}

	.user-table thead th {
		border-bottom: 2px solid rgb(var(--color-border));
		color: rgb(var(--color-muted-foreground));
		font-size: 0.6875rem;
		letter-spacing: 0.1em;
		text-transform: uppercase;
	}

	.user-table tbody td {
		border-bottom: 1px solid rgb(var(--color-border));
	}

	.user-table tbody tr:last-child td {
		border-bottom: 0;
	}

	.user-table tbody tr {
		transition: background 0.15s ease;
	}

	.user-table tbody tr:hover,
	.user-table tbody tr.current-user {
		background: rgb(var(--color-primary) / 0.045);
	}

	.current-badge,
	.connection-state,
	.timezone-chip {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		border: 1px solid rgb(var(--color-border));
		border-radius: 999px;
		padding: 0.25rem 0.5rem;
		font-size: 0.75rem;
		font-weight: 600;
		line-height: 1;
		white-space: nowrap;
	}

	.current-badge,
	.connection-state.is-connected {
		border-color: rgb(var(--color-primary) / 0.35);
		background: rgb(var(--color-primary) / 0.1);
		color: rgb(var(--color-primary));
	}

	.connection-state,
	.timezone-chip {
		color: rgb(var(--color-muted-foreground));
	}
</style>
