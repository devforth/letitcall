<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import Icon from '@iconify/svelte';
	import checkIcon from '@iconify-icons/tabler/check';
	import userPlusIcon from '@iconify-icons/tabler/user-plus';
	import usersIcon from '@iconify-icons/tabler/users';
	import worldIcon from '@iconify-icons/tabler/world';
	import xIcon from '@iconify-icons/tabler/x';
	import { callApi, appPath, getSession } from '$lib/api';
	import ImageSelector from '$lib/components/ImageSelector.svelte';
	import UserDeletionDialog from '$lib/components/UserDeletionDialog.svelte';
	import UserTable from '$lib/components/UserTable.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import ConfirmationDialog from '$lib/components/ui/ConfirmationDialog.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import SearchableSelect from '$lib/components/ui/SearchableSelect.svelte';
	import SegmentedControl from '$lib/components/ui/SegmentedControl.svelte';
	import type { ManagedUser, UserDeletionImpact } from '$lib/types';
	import { getLocalTimezones } from '$lib/timezones';

	let users = $state<ManagedUser[]>([]);
	let currentEmail = $state('');
	let email = $state('');
	let fullName = $state('');
	let password = $state('');
	let timezone = $state('UTC');
	let timezones = $state<string[]>(['UTC']);
	let showForm = $state(false);
	let loading = $state(true);
	let saving = $state(false);
	let checkingEmail = $state('');
	let deletingEmail = $state('');
	let reassigning = $state(false);
	let userToDelete = $state<ManagedUser | null>(null);
	let deletionImpact = $state<UserDeletionImpact | null>(null);
	let error = $state('');
	let avatarSelector = $state<ImageSelector | null>(null);
	let search = $state('');
	let connFilter = $state<'all' | 'connected' | 'notConnected'>('all');

	const connFilters: { value: 'all' | 'connected' | 'notConnected'; label: string }[] = [
		{ value: 'all', label: 'All' },
		{ value: 'connected', label: 'Connected' },
		{ value: 'notConnected', label: 'Not connected' }
	];
	const newUserContainerStyle =
		'background: rgb(var(--color-primary)); box-shadow: 0 0 0 1px var(--color-border);';
	const tableBlockStyle =
		'background: rgb(var(--color-background)); box-shadow: 0 0 0 1px var(--color-border);';
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
		const localTimezones = getLocalTimezones();
		timezone = localTimezones.current;
		timezones = localTimezones.options;

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

	async function createUser(event: SubmitEvent) {
		event.preventDefault();
		saving = true;
		error = '';
		try {
			const avatar = (await avatarSelector?.exportImage()) ?? '';
			const response = await callApi<{ user: ManagedUser }>('/api/users', {
				method: 'POST',
				body: JSON.stringify({ email, fullName, password, timezone, avatar })
			});
			users = [...users, response.user].sort((a, b) => a.email.localeCompare(b.email));
			email = '';
			fullName = '';
			password = '';
			timezone = getLocalTimezones().current;
			showForm = false;
		} catch (cause) {
			error = cause instanceof Error ? cause.message : 'Unable to create user';
		} finally {
			saving = false;
		}
	}

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

<section aria-labelledby="users-title" class="flex flex-col gap-6">
	<div class="mb-2">
		<PageHeader
			id="users-title"
			title="Users"
			description="Manage who can sign in and host events."
			icon={usersIcon}
		>
			{#if !showForm}
				<Button
					rounded
					style="font-weight: 500 !important; padding-right: 1rem !important; padding-bottom: 0.5rem !important;"
					class="add-user-button self-start"
					onclick={() => (showForm = true)}
				>
					<span class="flex items-center gap-2">
						<Icon icon={boldUserPlusIcon} width="18" height="18" class="shrink-0" />
						Add user
					</span>
				</Button>
			{/if}
		</PageHeader>
	</div>

	{#if showForm}
		<div class="mb-2 overflow-hidden rounded-[0.625rem]" style={newUserContainerStyle}>
			<form
				class="ml-1 flex flex-col rounded-md rounded-l-lg"
				style="background: rgb(var(--color-background));"
				onsubmit={createUser}
			>
				<div
					class="flex min-w-0 items-center gap-2 rounded-t-md p-3 sm:p-4"
					style="background: linear-gradient(110deg, rgb(var(--color-primary) / 0.12), rgb(var(--color-background)) 42%); box-shadow: inset 0 -1px 0 var(--color-border);"
				>
					<span class="grid size-8 shrink-0 place-items-center" style="color: rgb(var(--color-primary));">
						<Icon icon={userPlusIcon} width="26" height="26" />
					</span>
					<h2 class="text-xl font-semibold" style="color: rgb(var(--color-primary));">New user</h2>
				</div>
				<div class="grid gap-5 p-4 sm:p-5 lg:grid-cols-2">
					<Input id="new-email" label="Email" type="email" bind:value={email} required autocomplete="off" />
					<SearchableSelect
						id="new-timezone"
						emptyText="No matching timezones"
						label="Timezone"
						icon={worldIcon}
						options={timezones}
						bind:value={timezone}
						required
					/>
					<Input id="new-full-name" label="Full name (optional)" bind:value={fullName} autocomplete="name" />
					<Input
						id="new-password"
						label="Temporary password (optional)"
						type="password"
						bind:value={password}
						minlength={12}
						autocomplete="new-password"
					/>
					<div class="lg:col-span-2">
						<ImageSelector id="new-avatar" legend="Avatar (optional)" bind:this={avatarSelector} />
					</div>
					<div class="flex items-end justify-end gap-3 lg:col-span-2">
						<Button variant="primary-outline" class="outlined-action-button" onclick={() => (showForm = false)}>
							<span class="flex items-center gap-2">
								<Icon icon={xIcon} width="20" height="20" />
								Cancel
							</span>
						</Button>
						<Button type="submit" rounded class="primary-action-button" disabled={saving}>
							<span class="flex items-center gap-2">
								<Icon icon={checkIcon} width="20" height="20" />
								{saving ? 'Creating…' : 'Create'}
							</span>
						</Button>
					</div>
				</div>
			</form>
		</div>
	{/if}

	<div class="overflow-hidden rounded-lg" style={tableBlockStyle}>
		<div class="flex flex-wrap items-end justify-between gap-4 border-b p-3 sm:p-4" style="border-color: var(--color-border);">
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
		</div>

		{#if error}
			<p class="m-3 rounded-md border-2 p-3 text-sm" style="border-color: rgb(var(--error)); color: rgb(var(--error));" role="alert">{error}</p>
		{/if}
		{#if loading}
			<p class="p-8 text-sm" style="color: rgb(var(--color-text) / 0.65);">Loading users…</p>
		{:else}
			<UserTable users={filteredUsers} {currentEmail} {checkingEmail} {deletingEmail} onedit={editUser} ondelete={prepareDelete} />
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
	:global(.add-user-button) {
		min-height: 4rem !important;
		border-radius: 9999px !important;
		font-size: 1.25rem !important;
	}
</style>
