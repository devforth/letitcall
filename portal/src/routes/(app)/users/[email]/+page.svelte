<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { callApi, appPath, avatarURL } from '$lib/api';
	import ImageSelector from '$lib/components/ImageSelector.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import SearchableSelect from '$lib/components/ui/SearchableSelect.svelte';
	import { getLocalTimezones } from '$lib/timezones';
	import type { ManagedUser } from '$lib/types';

	let email = $state('');
	let fullName = $state('');
	let password = $state('');
	let timezone = $state('UTC');
	let timezones = $state<string[]>(['UTC']);
	let avatarPath = $state('');
	let avatarSelector = $state<ImageSelector | null>(null);
	let loading = $state(true);
	let saving = $state(false);
	let error = $state('');

	onMount(async () => {
		const localTimezones = getLocalTimezones();
		timezones = localTimezones.options;
		try {
			const response = await callApi<{ users: ManagedUser[] }>('/api/users');
			const user = response.users.find((candidate) => candidate.email === page.params.email);
			if (!user) throw new Error('User not found');
			email = user.email;
			fullName = user.fullName;
			timezone = user.timezone;
			avatarPath = user.avatarPath ?? '';
			if (!timezones.includes(timezone)) timezones = [timezone, ...timezones];
		} catch (cause) {
			error = cause instanceof Error ? cause.message : 'Unable to load user';
		} finally {
			loading = false;
		}
	});

	async function saveUser(event: SubmitEvent) {
		event.preventDefault();
		saving = true;
		error = '';
		try {
			const update: { fullName: string; timezone: string; password?: string; avatar?: string } = {
				fullName,
				timezone
			};
			if (password) update.password = password;
			const avatar = (await avatarSelector?.exportImage()) ?? '';
			if (avatar) update.avatar = avatar;
			await callApi(`/api/users/${encodeURIComponent(email)}`, {
				method: 'PATCH',
				body: JSON.stringify(update)
			});
			await goto(appPath('/users'));
		} catch (cause) {
			error = cause instanceof Error ? cause.message : 'Unable to update user';
		} finally {
			saving = false;
		}
	}
</script>

<PageTitle title="Edit user" />

<section aria-labelledby="edit-user-title">
	<div class="mb-6">
		<h1 id="edit-user-title" class="text-2xl font-semibold tracking-tight">Edit user</h1>
		<p class="mt-2 text-sm">Update account settings without changing the sign-in email.</p>
	</div>

	{#if error}
		<p class="mb-5 border border-black p-3 text-sm" role="alert">{error}</p>
	{/if}

	{#if loading}
		<p class="border border-black p-6 text-sm">Loading user…</p>
	{:else if email}
		<form class="grid gap-5 lg:grid-cols-2" onsubmit={saveUser}>
			<Input id="edit-email" label="Email" type="email" bind:value={email} readonly autocomplete="email" />
			<Input id="edit-full-name" label="Full name" bind:value={fullName} autocomplete="name" />
			<SearchableSelect
				id="edit-timezone"
				label="Timezone"
				options={timezones}
				bind:value={timezone}
				required
			/>
			<Input
				id="edit-password"
				label="New password"
				hint="Leave blank to keep current"
				type="password"
				bind:value={password}
				minlength={12}
				autocomplete="new-password"
			/>
			<div class="lg:col-span-2">
				<ImageSelector
					id="edit-avatar"
					legend="Avatar"
					current={avatarPath ? avatarURL(avatarPath) : ''}
					ondelete={() => (avatarPath = '')}
					bind:this={avatarSelector}
				/>
			</div>
			<div class="flex flex-wrap gap-2 lg:col-span-2">
				<Button type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save changes'}</Button>
				<Button variant="secondary" onclick={() => goto(appPath('/users'))}>Cancel</Button>
			</div>
		</form>
	{/if}
</section>
