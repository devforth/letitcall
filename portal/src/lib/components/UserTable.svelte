<script lang="ts">
	import Icon from '@iconify/svelte';
	import calendarCheckIcon from '@iconify-icons/tabler/check';
	import calendarXIcon from '@iconify-icons/tabler/x';
	import editIcon from '@iconify-icons/mdi/edit';
	import trashIcon from '@iconify-icons/tabler/trash';
	import worldIcon from '@iconify-icons/tabler/world';
	import calendarOffIcon from '@iconify-icons/tabler/calendar-off';
	import { nextSort } from '$lib/sort';
	import type { ManagedUser } from '$lib/types';
	import IconButton from '$lib/components/ui/IconButton.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import SortButton from '$lib/components/ui/SortButton.svelte';
	import TimedActions from '$lib/components/ui/TimedActions.svelte';

	type SortKey = 'name' | 'calendar' | 'timezone';

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
		sort = $bindable('name-ascending'),
		onedit,
		ondelete
	}: {
		users: ManagedUser[];
		currentEmail: string;
		checkingEmail?: string;
		deletingEmail?: string;
		sort?: string;
		onedit: (email: string) => void;
		ondelete: (email: string) => void;
	} = $props();

	const sortKey = $derived(sort.split('-')[0] as SortKey);
	const sortDirection = $derived(sort.split('-')[1] as 'ascending' | 'descending');

	const sortedUsers = $derived([...users].sort(compareUsers));

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
		sort = nextSort(sort, key);
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
						class={`${column.padding} py-3.5`}
					>
						<SortButton label={column.label} active={sortKey === column.key} direction={sortDirection} onclick={() => toggleSort(column.key)} />
					</th>
				{/each}
				<th aria-label="Actions" class="px-5 py-3.5 text-right"></th>
			</tr>
		</thead>
		<tbody>
			{#each sortedUsers as user (user.email)}
				<tr data-timed-actions-row>
					<td class="px-5 py-4">
						<div class="flex min-w-0 items-center gap-3">
							<span class="avatar-wrap">
								<Avatar name={user.fullName} email={user.email} avatarPath={user.avatarPath} size={40} />
							</span>
							<div class="min-w-0">
								<div class="flex items-center gap-2">
									<p class="truncate font-semibold" style="color: rgb(var(--color-text));">{user.fullName?.trim() || 'Unnamed user'}</p>
									{#if user.email === currentEmail}
										<span class="shrink-0" style="color: rgb(var(--color-text) / 0.65);">· you</span>
									{/if}
								</div>
								<p class="mt-0.5 truncate text-xs" style="color: rgb(var(--color-text) / 0.65);">{user.email}</p>
							</div>
						</div>
					</td>
					<td class="px-4 py-4">
						<div class="calendar-cell" class:connected={user.googleConnected}>
							{#if user.googleConnected}
								<Icon icon={calendarCheckIcon} width="20" height="20" class="calendar-status" aria-label="Calendar connected" style="color: rgb(var(--success));" />
							{:else}
								<Icon icon={calendarXIcon} width="20" height="20" class="calendar-status" aria-label="Calendar not connected" style="color: rgb(var(--color-text) / 0.65);" />
								<span class="calendar-text" aria-hidden="true"><Icon icon={calendarOffIcon} width="15" height="15" />Calendar not connected</span>
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
						<TimedActions label={`Show actions for ${user.email}`} controlsId={`user-actions-${user.email}`} actionsVisibleOnSmallScreens>
							<div class="user-actions flex justify-end gap-2">
								<IconButton filled tone="primary" label={`Edit ${user.email}`} onclick={() => onedit(user.email)}>
									<Icon icon={editIcon} width="20" height="20" />
								</IconButton>
								{#if user.email !== currentEmail}
									<IconButton
										filled
										tone="danger"
										label={checkingEmail === user.email ? 'Checking…' : deletingEmail === user.email ? 'Deleting…' : `Delete ${user.email}`}
										disabled={checkingEmail === user.email || deletingEmail === user.email}
										onclick={() => ondelete(user.email)}
									>
										<Icon icon={trashIcon} width="20" height="20" />
									</IconButton>
								{/if}
							</div>
						</TimedActions>
					</td>
				</tr>
			{:else}
				<tr>
					<td class="empty-table-cell" colspan="4">No users found</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>

<style>
	.user-table {
		border-collapse: collapse;
	}

	.user-table thead th {
		border-bottom: 1px solid var(--color-border);
		color: rgb(var(--color-text));
		font-size: 0.75rem;
		font-weight: 400;
		letter-spacing: 0.025em;
		text-transform: none;
	}

	.user-table tbody td {
		border-bottom: 1px solid var(--color-border);
	}

	.user-table tbody tr:last-child td {
		border-bottom: 0;
	}

	.empty-table-cell {
		padding: 2.5rem 1rem;
		color: rgb(var(--color-text) / 0.65);
		font-size: 0.875rem;
		text-align: center;
	}

	:global(.calendar-status path) {
		stroke-width: 3;
	}

	.calendar-text {
		display: none;
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
		border: 1px solid var(--color-border);
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

	.avatar-wrap {
		display: block;
		position: relative;
		width: 2.5rem;
		height: 2.5rem;
		flex: none;
	}

	/* Below the table's 42rem minimum it would scroll, so each user becomes a card. */
	@container users (max-width: 42rem) {
		.user-table,
		.user-table tbody {
			display: grid;
			gap: 1.25rem;
			min-width: 0;
		}

		.user-table thead {
			display: none;
		}

		.user-table tbody tr {
			display: grid;
			grid-template-columns: auto minmax(0, 1fr);
			grid-template-areas: 'user user' 'calendar timezone' 'actions actions';
			align-items: center;
			gap: 0.75rem 1rem;
			border-radius: 0.5rem;
			padding: 1rem 1.25rem 1.25rem;
			background: rgb(var(--color-background));
			/* Inset: the table's scroll wrapper clips anything drawn outside the row. */
			box-shadow: inset 0 0 0 1px var(--color-border);
		}

		.user-table tbody td {
			border-bottom: 0;
			padding: 0;
		}

		.user-table tbody td:nth-child(1) {
			grid-area: user;
		}

		.user-table tbody td:nth-child(2) {
			grid-area: calendar;
		}

		.user-table tbody td:nth-child(3) {
			grid-area: timezone;
		}

		.user-table tbody td:nth-child(4) {
			grid-area: actions;
		}

		.calendar-cell {
			align-items: center;
			margin-right: 0;
			border: 1px solid var(--color-border);
			border-radius: 999px;
			padding: 0.25rem 0.5rem;
			color: rgb(var(--color-text) / 0.65);
			font-size: 0.75rem;
			font-weight: 600;
			line-height: 1;
			white-space: nowrap;
		}

		.calendar-cell :global(.calendar-status),
		.calendar-cell.connected {
			display: none;
		}

		.calendar-text {
			display: inline-flex;
			align-items: center;
			gap: 0.375rem;
		}
	}
</style>
