<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import Icon from '@iconify/svelte';
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
		'background: rgb(var(--color-foreground)); border-color: rgb(var(--color-border)); box-shadow: var(--shadow-small);';

	const filteredUsers = $derived(
		users.filter((candidate) => {
			const q = search.trim().toLowerCase();
			const matchesQuery =
				!q ||
				candidate.email.toLowerCase().includes(q) ||
				(candidate.fullName ?? '').toLowerCase().includes(q);
			const matchesConnection =
				connFilter === 'all' ||
				(connFilter === 'connected' ? candidate.googleConnected : !candidate.googleConnected);
			return matchesQuery && matchesConnection;
		})
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

<section aria-labelledby="users-title" class="flex flex-col gap-5">
	<div class="rounded-lg border-2 p-4 sm:p-5" style={blockStyle}>
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
					<p class="text-sm" style="color: rgb(var(--color-muted-foreground));">Manage who can sign in and host events.</p>
				</div>
			</div>
			<div>
				{#if showForm}
					<Button variant="ghost" class="size-10 !min-h-0 !p-0" onclick={() => (showForm = false)}>
						<Icon icon={xIcon} width="20" height="20" />
						<span class="sr-only">Close new user form</span>
					</Button>
				{:else}
					<Button onclick={() => (showForm = true)}>
						<span class="flex items-center gap-2">
							<Icon icon={addUserIcon} width="18" height="18" class="shrink-0" />
							Add user
						</span>
					</Button>
				{/if}
			</div>
		</div>

		{#if showForm}
			<form
				class="mt-6 grid gap-5 border-t-2 pt-5 lg:grid-cols-3"
				style="border-color: rgb(var(--color-border));"
				onsubmit={createUser}
			>
				<div class="lg:col-span-3">
					<h2 class="font-semibold" style="color: rgb(var(--color-text));">New user</h2>
					<p class="mt-1 text-sm" style="color: rgb(var(--color-muted-foreground));">Add their details and optional sign-in password.</p>
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
					options={timezones}
					bind:value={timezone}
					required
				/>
				<div class="lg:col-span-3">
					<ImageSelector id="new-avatar" legend="Avatar" bind:this={avatarSelector} />
				</div>
				<div class="flex items-end lg:col-span-3">
					<Button type="submit" disabled={saving}>{saving ? 'Creating…' : 'Create user'}</Button>
				</div>
			</form>
		{/if}
	</div>

	<div class="overflow-hidden rounded-lg border-2" style={blockStyle}>
		<div class="flex flex-wrap items-end justify-between gap-4 border-b-2 p-3 sm:p-4" style="border-color: rgb(var(--color-border));">
			<div>
				<h2 class="font-semibold" style="color: rgb(var(--color-text));">People</h2>
				<p class="mt-1 text-sm" style="color: rgb(var(--color-muted-foreground));">
					{loading ? 'Loading your team…' : `${filteredUsers.length} of ${users.length} shown`}
				</p>
			</div>
			<div class="flex w-full flex-wrap gap-3 lg:w-auto lg:flex-nowrap">
				<div class="min-w-[220px] flex-1 lg:w-72 lg:flex-none">
					<Input id="user-search" label="Search users" type="search" bind:value={search} />
				</div>
				<div
					class="flex shrink-0 overflow-hidden rounded-md border-2"
					style="border-color: rgb(var(--color-border));"
					role="group"
					aria-label="Filter by Google connection"
				>
					{#each connFilters as filter (filter.value)}
						<button
							type="button"
							class="filter-seg px-3 py-2 text-sm font-semibold"
							class:on={connFilter === filter.value}
							aria-pressed={connFilter === filter.value}
							onclick={() => (connFilter = filter.value)}
						>
							{filter.label}
						</button>
					{/each}
				</div>
			</div>
		</div>

		{#if error}
			<p class="m-3 rounded-md border-2 p-3 text-sm" style="border-color: rgb(var(--error)); color: rgb(var(--error));" role="alert">{error}</p>
		{/if}
		{#if loading}
			<p class="p-8 text-sm" style="color: rgb(var(--color-muted-foreground));">Loading users…</p>
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
	.filter-seg {
		background: transparent;
		color: rgb(var(--color-muted-foreground));
		border-right: 2px solid rgb(var(--color-border));
		cursor: pointer;
		transition:
			background 0.15s,
			color 0.15s;
	}

	.filter-seg:last-child {
		border-right: 0;
	}

	.filter-seg:hover:not(.on) {
		background: rgb(var(--color-primary) / 0.08);
		color: rgb(var(--color-primary));
	}

	.filter-seg.on {
		background: rgb(var(--color-primary));
		color: rgb(var(--color-contrast-text));
	}
</style>
