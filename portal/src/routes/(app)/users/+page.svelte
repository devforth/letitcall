<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { callApi, appPath, getSession } from '$lib/api';
	import AvatarSelector from '$lib/components/AvatarSelector.svelte';
	import UserTable from '$lib/components/UserTable.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import SearchableSelect from '$lib/components/ui/SearchableSelect.svelte';
	import type { ManagedUser } from '$lib/types';
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
	let deletingEmail = $state('');
	let error = $state('');
	let avatarSelector = $state<AvatarSelector | null>(null);

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
			const avatar = (await avatarSelector?.exportAvatar()) ?? '';
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

	async function deleteUser(emailToDelete: string) {
		if (!window.confirm(`Delete ${emailToDelete}?`)) return;
		deletingEmail = emailToDelete;
		error = '';
		try {
			await callApi(`/api/users/${encodeURIComponent(emailToDelete)}`, { method: 'DELETE' });
			users = users.filter((user) => user.email !== emailToDelete);
		} catch (cause) {
			error = cause instanceof Error ? cause.message : 'Unable to delete user';
		} finally {
			deletingEmail = '';
		}
	}
</script>

<svelte:head><title>Users · Let It Call</title></svelte:head>

<section aria-labelledby="users-title">
	<div class="mb-6 flex flex-wrap items-center justify-between gap-4">
		<div>
			<h1 id="users-title" class="text-2xl font-semibold tracking-tight">Users</h1>
			<p class="mt-2 text-sm">Manage who can sign in and schedule events.</p>
		</div>
		<Button onclick={() => (showForm = !showForm)}>{showForm ? 'Cancel' : 'Add user'}</Button>
	</div>

	{#if error}
		<p class="mb-5 border border-black p-3 text-sm" role="alert">{error}</p>
	{/if}

	{#if showForm}
		<form class="mb-8 grid gap-5 border border-black p-5 lg:grid-cols-3" onsubmit={createUser}>
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
				<AvatarSelector id="new-avatar" bind:this={avatarSelector} />
			</div>
			<div class="flex items-end lg:col-span-3">
				<Button type="submit" disabled={saving}>{saving ? 'Creating…' : 'Create user'}</Button>
			</div>
		</form>
	{/if}

	{#if loading}
		<p class="border border-black p-6 text-sm">Loading users…</p>
	{:else}
		<UserTable {users} {currentEmail} {deletingEmail} onedit={editUser} ondelete={deleteUser} />
	{/if}
</section>
