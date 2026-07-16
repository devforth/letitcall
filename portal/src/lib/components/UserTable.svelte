<script lang="ts">
	import Icon from '@iconify/svelte';
	import calendarCheckIcon from '@iconify-icons/tabler/check';
	import calendarXIcon from '@iconify-icons/tabler/x';
	import checkIcon from '@iconify-icons/tabler/circle-check-filled';
	import editIcon from '@iconify-icons/tabler/edit';
	import trashIcon from '@iconify-icons/tabler/trash';
	import usersIcon from '@iconify-icons/tabler/users';
	import worldIcon from '@iconify-icons/tabler/world';
	import type { ManagedUser } from '$lib/types';
	import IconButton from '$lib/components/ui/IconButton.svelte';
	import { avatarURL } from '$lib/api';

	type SortKey = 'name' | 'calendar' | 'timezone';
	type SortDirection = 'ascending' | 'descending';

	const sortableColumns: { key: SortKey; label: string; padding: string }[] = [
		{ key: 'name', label: 'User', padding: 'px-5' },
		{ key: 'calendar', label: 'Calendar', padding: 'px-4' },
		{ key: 'timezone', label: 'Timezone', padding: 'px-4' }
	];

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

	let sortKey = $state<SortKey>('name');
	let sortDirection = $state<SortDirection>('ascending');

	const sortedUsers = $derived([...users].sort(compareUsers));

	function initialsFor(user: ManagedUser) {
		const parts = user.fullName?.trim().split(/\s+/).filter(Boolean) ?? [];
		if (parts.length >= 2) return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
		if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
		return user.email.slice(0, 2).toUpperCase();
	}

	function compareUsers(first: ManagedUser, second: ManagedUser) {
		const comparison = sortValue(first).localeCompare(sortValue(second)) || first.email.localeCompare(second.email);
		return sortDirection === 'ascending' ? comparison : -comparison;
	}

	function sortValue(user: ManagedUser) {
		if (sortKey === 'calendar') return user.googleConnected ? 'Connected' : 'Not connected';
		if (sortKey === 'timezone') return user.timezone;
		return user.fullName?.trim() || 'Unnamed user';
	}

	function toggleSort(key: SortKey) {
		if (sortKey === key) {
			sortDirection = sortDirection === 'ascending' ? 'descending' : 'ascending';
			return;
		}

		sortKey = key;
		sortDirection = 'ascending';
	}
</script>

<div class="overflow-x-auto">
	<table class="user-table w-full min-w-[42rem] text-left text-sm">
		<thead>
			<tr>
				{#each sortableColumns as column (column.key)}
					<th
						aria-sort={sortKey === column.key ? sortDirection : 'none'}
						class:text-center={column.key === 'calendar'}
						class={`${column.padding} py-3.5 font-semibold`}
					>
						<button
							type="button"
							class:active-sort={sortKey === column.key}
							class="sort-button"
							onclick={() => toggleSort(column.key)}
						>
							{column.label}
							<svg
								class="sort-chevrons"
								viewBox="0 0 24 24"
								width="22"
								height="22"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
								stroke-linecap="round"
								stroke-linejoin="round"
								aria-hidden="true"
							>
								<path
									class="chev"
									class:strong={sortKey === column.key && sortDirection === 'ascending'}
									class:faint={sortKey === column.key && sortDirection === 'descending'}
									d="M8 10l4 -4l4 4"
								/>
								<path
									class="chev"
									class:strong={sortKey === column.key && sortDirection === 'descending'}
									class:faint={sortKey === column.key && sortDirection === 'ascending'}
									d="M8 14l4 4l4 -4"
								/>
							</svg>
						</button>
					</th>
				{/each}
				<th aria-label="Actions" class="px-5 py-3.5 text-right font-semibold"></th>
			</tr>
		</thead>
		<tbody>
			{#each sortedUsers as user (user.email)}
				<tr>
					<td class="px-5 py-4">
						<div class="flex min-w-0 items-center gap-3">
							<span class="avatar-wrap">
								{#if user.avatarPath}
									<img
										src={avatarURL(user.avatarPath)}
										alt=""
										class="size-10 rounded-xl object-cover"
										style="border: 2px solid rgb(var(--color-border));"
									/>
								{:else}
									<span
										class="flex size-10 items-center justify-center rounded-xl text-sm font-bold leading-none"
										style="background: rgb(var(--color-primary) / 0.14); color: rgb(var(--color-primary));"
										aria-label="No avatar"
									>
										{initialsFor(user)}
									</span>
								{/if}
							</span>
							<div class="min-w-0">
								<div class="flex items-center gap-2">
									<p class="truncate font-semibold" style="color: rgb(var(--color-text));">{user.fullName?.trim() || 'Unnamed user'}</p>
									{#if user.email === currentEmail}
										<span class="current-user-marker">
											<Icon icon={checkIcon} width="16" height="16" aria-label="Current user" />
										</span>
									{/if}
								</div>
								<p class="mt-0.5 truncate text-xs" style="color: rgb(var(--color-text) / 0.65);">{user.email}</p>
							</div>
						</div>
					</td>
					<td class="px-4 py-4">
						<div class="calendar-cell">
							{#if user.googleConnected}
								<Icon icon={calendarCheckIcon} width="20" height="20" class="calendar-status" aria-label="Calendar connected" style="color: rgb(var(--color-primary));" />
							{:else}
								<Icon icon={calendarXIcon} width="20" height="20" class="calendar-status" aria-label="Calendar not connected" style="color: rgb(var(--color-text) / 0.65);" />
							{/if}
						</div>
					</td>
					<td class="px-4 py-4">
						<span class="timezone-chip">
							<Icon icon={worldIcon} width="15" height="15" class="shrink-0" />
							{user.timezone}
						</span>
					</td>
					<td class="px-5 py-4">
						<div class="user-actions flex justify-end gap-2">
							<IconButton tone="primary" label={`Edit ${user.email}`} onclick={() => onedit(user.email)}>
								<Icon icon={editIcon} width="20" height="20" />
							</IconButton>
							<IconButton
								tone="danger"
								label={checkingEmail === user.email ? 'Checking…' : deletingEmail === user.email ? 'Deleting…' : `Delete ${user.email}`}
								disabled={user.email === currentEmail || checkingEmail === user.email || deletingEmail === user.email}
								onclick={() => ondelete(user.email)}
							>
								<Icon icon={trashIcon} width="20" height="20" />
							</IconButton>
						</div>
					</td>
				</tr>
			{:else}
				<tr>
					<td class="px-5 py-14 text-center" colspan="4">
						<div class="mx-auto flex max-w-xs flex-col items-center">
							<Icon icon={usersIcon} width="30" height="30" style="color: rgb(var(--color-text) / 0.65);" />
							<p class="mt-3 font-semibold" style="color: rgb(var(--color-text));">No users found</p>
							<p class="mt-1 text-xs" style="color: rgb(var(--color-text) / 0.65);">Try a different search or connection filter.</p>
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
		background: rgb(var(--color-text) / 0.06);
	}

	.user-table thead th {
		border-bottom: 2px solid rgb(var(--color-border));
		color: rgb(var(--color-text));
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.025em;
		text-transform: none;
	}

	.sort-button {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		border: 0;
		padding: 0;
		background: transparent;
		color: inherit;
		font: inherit;
		letter-spacing: inherit;
		text-transform: inherit;
		cursor: pointer;
		transition: color 0.15s ease;
	}

	.sort-button:hover,
	.sort-button.active-sort {
		color: rgb(var(--color-primary));
	}

	.sort-button:focus-visible {
		outline: 2px solid rgb(var(--color-primary));
		outline-offset: 3px;
	}

	.sort-chevrons .chev {
		opacity: 0.5;
		transition: opacity 0.15s ease;
	}

	.sort-chevrons .chev.strong {
		opacity: 1;
	}

	.sort-chevrons .chev.faint {
		opacity: 0.28;
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

	.user-table tbody tr:hover {
		background: rgb(var(--color-primary) / 0.045);
	}

	.user-actions {
		opacity: 0;
		pointer-events: none;
		transform: translateX(0.5rem);
		transition:
			opacity 0.18s ease,
			transform 0.18s ease;
	}

	.user-table tbody tr:hover .user-actions,
	.user-table tbody tr:focus-within .user-actions {
		opacity: 1;
		pointer-events: auto;
		transform: translateX(0);
	}

	@media (prefers-reduced-motion: reduce) {
		.user-actions {
			transition: none;
		}
	}

	:global(.calendar-status path) {
		stroke-width: 3;
	}

	.calendar-cell {
		display: flex;
		justify-content: center;
		margin-right: 20px;
	}

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

	.timezone-chip {
		color: rgb(var(--color-text) / 0.65);
	}

	.current-user-marker {
		display: inline-flex;
		transform: translateY(1px);
		color: rgb(var(--color-primary));
	}

	.avatar-wrap {
		display: block;
		position: relative;
		width: 2.5rem;
		height: 2.5rem;
		flex: none;
	}
</style>
