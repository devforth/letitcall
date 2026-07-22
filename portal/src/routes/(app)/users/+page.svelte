<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import Icon from '@iconify/svelte';
	import checkIcon from '@iconify-icons/tabler/check';
	import plusIcon from '@iconify-icons/tabler/plus';
	import worldIcon from '@iconify-icons/tabler/world';
	import xIcon from '@iconify-icons/tabler/x';
	import { callApi, appPath, getSession } from '$lib/api';
	import addUserIcon from '$lib/icons/add-user';
	import ImageSelector from '$lib/components/ImageSelector.svelte';
	import UserDeletionDialog from '$lib/components/UserDeletionDialog.svelte';
	import UserTable from '$lib/components/UserTable.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import ConfirmationDialog from '$lib/components/ui/ConfirmationDialog.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import SearchableSelect from '$lib/components/ui/SearchableSelect.svelte';
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

	const blockStyle =
		'background: rgb(var(--color-foreground)); box-shadow: var(--shadow-small);';
	const newUserContainerStyle = `${blockStyle} background: rgb(var(--color-primary)); border-color: rgb(var(--color-primary));`;

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

<section aria-labelledby="users-title" class="flex flex-col gap-4">
	<div class="rounded-lg p-4 sm:p-5" style={blockStyle}>
		<div class="flex flex-wrap items-center justify-between gap-5">
			<div class="flex min-w-0 items-center gap-4">
				<div
					class="grid size-12 shrink-0 place-items-center rounded-lg"
					style="background: rgb(var(--color-primary) / 0.12); color: rgb(var(--color-primary));"
				>
					<Icon icon={addUserIcon} width="24" height="24" />
				</div>
				<div>
					<div class="flex items-center gap-3">
						<h1 id="users-title" class="text-2xl font-semibold tracking-tight" style="color: rgb(var(--color-text));">Users</h1>
						<span class="inline-flex items-center gap-1.5 rounded-md px-2 py-1 text-xs font-semibold" style="background: rgb(var(--color-primary) / 0.1); color: rgb(var(--color-primary));">
							{loading ? 'Loading…' : `${users.length} ${users.length === 1 ? 'member' : 'members'}`}
						</span>
					</div>
					<p class="text-sm" style="color: rgb(var(--color-text) / 0.65);">Manage who can sign in and host events.</p>
				</div>
			</div>
			{#if !showForm}
				<Button onclick={() => (showForm = true)}>
					<span class="flex items-center gap-2">
						<Icon icon={plusIcon} width="18" height="18" class="add-user-plus shrink-0" />
						Add user
					</span>
				</Button>
			{/if}
		</div>
	</div>

	{#if showForm}
		<div class="rounded-[0.625rem] border-2" style={newUserContainerStyle}>
			<form
				class="ml-1 grid gap-5 rounded-md rounded-l-lg p-4 sm:p-5 lg:grid-cols-3"
				style="background: rgb(var(--color-foreground));"
				onsubmit={createUser}
			>
				<div class="lg:col-span-3">
					<h2 class="font-semibold" style="color: rgb(var(--color-text));">New user</h2>
					<p class="mt-1 text-sm" style="color: rgb(var(--color-text) / 0.65);">Add their details and optional sign-in password.</p>
				</div>
				<Input id="new-email" label="Email" type="email" bind:value={email} required autocomplete="off" />
				<Input id="new-full-name" label="Full name (optional)" bind:value={fullName} autocomplete="name" />
				<Input
					id="new-password"
					label="Temporary password (optional)"
					type="password"
					bind:value={password}
					minlength={12}
					autocomplete="new-password"
				/>
				<SearchableSelect
					id="new-timezone"
					label="Timezone"
					icon={worldIcon}
					options={timezones}
					bind:value={timezone}
					required
				/>
				<div class="lg:col-span-3">
					<ImageSelector id="new-avatar" legend="Avatar" bind:this={avatarSelector} />
				</div>
				<div class="flex items-end gap-3 lg:col-span-3">
					<Button variant="secondary" onclick={() => (showForm = false)}>
						<span class="flex items-center gap-2">
							<Icon icon={xIcon} width="18" height="18" class="cancel-icon shrink-0" />
							Cancel
						</span>
					</Button>
					<Button type="submit" disabled={saving}>
						<span class="flex items-center gap-2">
							<Icon icon={checkIcon} width="18" height="18" class="create-user-icon shrink-0" />
							{saving ? 'Creating…' : 'Create user'}
						</span>
					</Button>
				</div>
			</form>
		</div>
	{/if}

	<div class="overflow-hidden rounded-lg" style={blockStyle}>
		<div class="flex flex-wrap items-end justify-between gap-4 border-b-2 p-3 sm:p-4" style="border-color: rgb(var(--color-border));">
			<div>
				<h2 class="font-semibold" style="color: rgb(var(--color-text));">People</h2>
				<p class="mt-1 text-sm" style="color: rgb(var(--color-text) / 0.65);">
					{loading ? 'Loading your team…' : `${filteredUsers.length} of ${users.length} shown`}
				</p>
			</div>
			<div
				class="connection-filter lg:self-center"
				role="group"
				aria-label="Filter by Google connection"
			>
				{#each connFilters as filter (filter.value)}
					<button
						type="button"
						class="filter-seg"
						class:on={connFilter === filter.value}
						aria-pressed={connFilter === filter.value}
						onclick={() => (connFilter = filter.value)}
					>
						{filter.label}
						<span class="filter-count">· {connCounts[filter.value]}</span>
					</button>
				{/each}
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
	.connection-filter {
		display: inline-flex;
		flex-shrink: 0;
		align-items: center;
		gap: 0.25rem;
		border: 1px solid rgb(var(--color-border));
		border-radius: 999px;
		padding: 0.25rem;
		background: rgb(var(--color-text) / 0.06);
	}

	:global(.cancel-icon path) {
		stroke-width: 3;
	}

	:global(.create-user-icon path) {
		stroke-width: 3;
	}

	:global(.add-user-plus path) {
		stroke-width: 3;
	}

	.filter-seg {
		border: 0;
		border-radius: 999px;
		padding: 0.4rem 0.9rem;
		background: transparent;
		color: rgb(var(--color-text) / 0.65);
		font-size: 0.8125rem;
		font-weight: 600;
		cursor: pointer;
		transition: color 0.15s, background 0.15s;
	}

	.filter-count {
		opacity: 0.45;
		font-variant-numeric: tabular-nums;
	}

	.filter-seg:hover:not(.on) {
		color: rgb(var(--color-text));
	}

	.filter-seg.on {
		background: rgb(var(--color-primary));
		color: rgb(var(--color-contrast-text));
		box-shadow: 0 1px 3px rgb(0 0 0 / 0.12);
	}
</style>
