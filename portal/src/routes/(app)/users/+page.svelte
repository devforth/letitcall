<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import Icon from '@iconify/svelte';
	import userPlusIcon from '@iconify-icons/tabler/user-plus';
	import sortIcon from '@iconify-icons/tabler/arrows-sort';
	import usersIcon from '@iconify-icons/tabler/users';
	import { callApi, appPath, getSession } from '$lib/api';
	import UserDeletionDialog from '$lib/components/UserDeletionDialog.svelte';
	import UserTable from '$lib/components/UserTable.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import ConfirmationDialog from '$lib/components/ui/ConfirmationDialog.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import SegmentedControl from '$lib/components/ui/SegmentedControl.svelte';
	import SearchableSelect from '$lib/components/ui/SearchableSelect.svelte';
	import type { ManagedUser, UserDeletionImpact } from '$lib/types';

	let users = $state<ManagedUser[]>([]);
	let currentEmail = $state('');
	let loading = $state(true);
	let checkingEmail = $state('');
	let deletingEmail = $state('');
	let reassigning = $state(false);
	let userToDelete = $state<ManagedUser | null>(null);
	let deletionImpact = $state<UserDeletionImpact | null>(null);
	let error = $state('');
	let search = $state('');
	let connFilter = $state<'all' | 'connected' | 'notConnected'>('all');

	const connFilters: { value: 'all' | 'connected' | 'notConnected'; label: string }[] = [
		{ value: 'all', label: 'All' },
		{ value: 'connected', label: 'Connected' },
		{ value: 'notConnected', label: 'Not connected' }
	];
	let sort = $state('name-ascending');
	const sortOptions = [
		{ value: 'name-ascending', label: 'Name, A-Z' },
		{ value: 'name-descending', label: 'Name, Z-A' },
		{ value: 'calendar-ascending', label: 'Calendar connected first' },
		{ value: 'calendar-descending', label: 'Calendar not connected first' },
		{ value: 'timezone-ascending', label: 'Timezone, A-Z' },
		{ value: 'timezone-descending', label: 'Timezone, Z-A' }
	];
	const boldUserPlusIcon = {
		...userPlusIcon,
		body: userPlusIcon.body.replace('stroke-width="2"', 'stroke-width="2.25"')
	};

	const searchMatches = $derived(
		users.filter((candidate) => {
			const q = search.trim().toLowerCase();
			return (
				!q ||
				candidate.email.toLowerCase().includes(q) ||
				(candidate.fullName ?? '').toLowerCase().includes(q)
			);
		})
	);

	const connCounts = $derived({
		all: searchMatches.length,
		connected: searchMatches.filter((candidate) => candidate.googleConnected).length,
		notConnected: searchMatches.filter((candidate) => !candidate.googleConnected).length
	});
	const connectionFilterOptions = $derived(
		connFilters.map((filter) => ({ ...filter, suffix: `· ${connCounts[filter.value]}` }))
	);

	const filteredUsers = $derived(
		searchMatches.filter(
			(candidate) =>
				connFilter === 'all' ||
				(connFilter === 'connected' ? candidate.googleConnected : !candidate.googleConnected)
		)
	);

	onMount(async () => {
		try {
			const [session, response] = await Promise.all([
				getSession(),
				callApi<{ users: ManagedUser[] }>('/api/users')
			]);
			currentEmail = session.user.email;
			users = response.users;
		} catch (cause) {
			error = cause instanceof Error ? cause.message : 'Unable to load users';
		} finally {
			loading = false;
		}
	});

	function editUser(emailToEdit: string) {
		void goto(appPath(`/users/${encodeURIComponent(emailToEdit)}`));
	}

	async function prepareDelete(emailToDelete: string) {
		checkingEmail = emailToDelete;
		try {
			deletionImpact = await callApi<UserDeletionImpact>(
				`/api/users/${encodeURIComponent(emailToDelete)}/deletion-impact`
			);
			userToDelete = users.find((user) => user.email === emailToDelete) ?? null;
		} catch {
			// callApi reports the error globally.
		} finally {
			checkingEmail = '';
		}
	}

	async function deleteUser() {
		const user = userToDelete!;
		deletingEmail = user.email;
		try {
			await callApi(`/api/users/${encodeURIComponent(user.email)}`, { method: 'DELETE' });
			users = users.filter((candidate) => candidate.email !== user.email);
			closeDeletionDialog();
		} catch {
			// callApi reports the error globally.
		} finally {
			deletingEmail = '';
		}
	}

	async function reassignAndDelete(newHostEmail: string) {
		const user = userToDelete!;
		reassigning = true;
		try {
			await callApi(`/api/users/${encodeURIComponent(user.email)}/reassign-bookings`, {
				method: 'POST',
				body: JSON.stringify({ newHostEmail })
			});
			await deleteUser();
		} catch {
			// callApi reports the error globally.
		} finally {
			reassigning = false;
		}
	}

	function closeDeletionDialog() {
		userToDelete = null;
		deletionImpact = null;
	}

</script>

<PageTitle title="Users" />

<section aria-labelledby="users-title" class="users-page flex flex-col gap-6">
	<div class="mb-2">
		<PageHeader
			id="users-title"
			title="Users"
			description="Manage who can sign in and host events."
			icon={usersIcon}
		>
			<Button rounded class="primary-action-button self-start" onclick={() => void goto(appPath('/users/new'))}>
				<span class="flex items-center gap-2">
					<Icon icon={boldUserPlusIcon} width="20" height="20" class="shrink-0" />
					Add user
				</span>
			</Button>
		</PageHeader>
	</div>

	<div class="users-panel">
		<div class="users-toolbar flex flex-wrap items-end justify-between gap-4 p-3 sm:p-4">
			<div>
				<h2 class="font-semibold" style="color: rgb(var(--color-text));">People</h2>
				<p class="mt-1 text-sm" style="color: rgb(var(--color-text) / 0.65);">
					{loading ? 'Loading your team…' : `${filteredUsers.length} of ${users.length} shown`}
				</p>
			</div>
			<div class="lg:self-center">
				<SegmentedControl
					options={connectionFilterOptions}
					value={connFilter}
					label="Filter by Google connection"
					onchange={(value) => (connFilter = value as typeof connFilter)}
				/>
			</div>
			<div class="min-w-[220px] flex-1 lg:w-72 lg:flex-none">
				<Input id="user-search" label="Search users" type="search" bind:value={search} />
			</div>
			<div class="users-sort w-full">
				<SearchableSelect id="user-sort" label="Sort by" icon={sortIcon} options={sortOptions} bind:value={sort} clearable={false} required />
			</div>
		</div>

		{#if error}
			<p class="m-3 rounded-md border-2 p-3 text-sm" style="border-color: rgb(var(--error)); color: rgb(var(--error));" role="alert">{error}</p>
		{/if}
		{#if loading}
			<p class="p-8 text-sm" style="color: rgb(var(--color-text) / 0.65);">Loading users…</p>
		{:else}
			<UserTable users={filteredUsers} bind:sort {currentEmail} {checkingEmail} {deletingEmail} onedit={editUser} ondelete={prepareDelete} />
		{/if}
	</div>
</section>

{#if userToDelete && deletionImpact?.requiresReassignment}
	<UserDeletionDialog
		open
		user={userToDelete}
		impact={deletionImpact}
		candidates={users.filter((user) => user.email !== userToDelete?.email)}
		confirming={reassigning}
		onconfirm={reassignAndDelete}
		oncancel={closeDeletionDialog}
	/>
{:else if userToDelete && deletionImpact}
	<ConfirmationDialog
		open
		title="Delete user?"
		description={`Delete ${userToDelete.email}? This action cannot be undone.`}
		confirmLabel="Delete user"
		confirmingLabel="Deleting…"
		confirming={deletingEmail === userToDelete.email}
		onconfirm={deleteUser}
		oncancel={closeDeletionDialog}
	/>
{/if}

<style>
	.users-page {
		container: users / inline-size;
	}

	.users-panel {
		overflow: hidden;
		border-radius: 0.5rem;
		background: rgb(var(--color-background));
		box-shadow: 0 0 0 1px var(--color-border);
	}

	.users-toolbar {
		border-bottom: 1px solid var(--color-border);
		background: linear-gradient(to bottom, transparent 75%, color-mix(in srgb, var(--color-border) 30%, transparent));
	}

	.users-sort {
		display: none;
	}

	/* Matches UserTable: below 42rem the toolbar and every user become separate cards. */
	@container users (max-width: 42rem) {
		.users-panel {
			display: grid;
			gap: 1.25rem;
			overflow: visible;
			background: none;
			box-shadow: none;
		}

		.users-toolbar {
			border-bottom: 0;
			border-radius: 0.5rem;
			background:
				linear-gradient(to bottom, transparent 75%, color-mix(in srgb, var(--color-border) 30%, transparent)),
				rgb(var(--color-background));
			box-shadow: 0 0 0 1px var(--color-border);
		}

		.users-sort {
			display: block;
		}
	}
</style>
