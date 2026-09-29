<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import checkIcon from '@iconify-icons/tabler/check';
	import userEditIcon from '@iconify-icons/tabler/user-edit';
	import worldIcon from '@iconify-icons/tabler/world';
	import xIcon from '@iconify-icons/tabler/x';
	import { callApi, appPath, avatarURL } from '$lib/api';
	import ImageSelector from '$lib/components/ImageSelector.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import SearchableSelect from '$lib/components/ui/SearchableSelect.svelte';
	import { getLocalTimezones } from '$lib/timezones';
	import type { ImageSource, ImageUpload, ManagedUser } from '$lib/types';

	let email = $state('');
	let fullName = $state('');
	let password = $state('');
	let timezone = $state('UTC');
	let timezones = $state<string[]>(['UTC']);
	let avatarPath = $state('');
	let avatarSource = $state<ImageSource>();
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
			avatarSource = user.avatarSource;
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
			const update: { fullName: string; timezone: string; password?: string; avatar?: ImageUpload } = {
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
	<div class="mb-8">
		<PageHeader
			id="edit-user-title"
			title="Edit user"
			description="Update account settings without changing the sign-in email."
			icon={userEditIcon}
		/>
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
				emptyText="No matching timezones"
				label="Timezone"
				icon={worldIcon}
				options={timezones}
				bind:value={timezone}
				required
			/>
			<Input
				id="edit-password"
				label="New password (leave blank to keep current)"
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
					original={avatarSource ? avatarURL(avatarSource.path) : ''}
					editor={avatarSource?.editor}
					showCurrentCopy={false}
					ondelete={() => (avatarPath = '')}
					bind:this={avatarSelector}
				/>
			</div>
			<div class="mt-3 flex flex-wrap justify-end gap-3 lg:col-span-2">
				<Button variant="primary-outline" class="outlined-action-button" onclick={() => goto(appPath('/users'))}>
					<span class="flex items-center gap-2">
						<Icon icon={xIcon} width="20" height="20" />
						Cancel
					</span>
				</Button>
				<Button type="submit" rounded class="primary-action-button" disabled={saving}>
					<span class="flex items-center gap-2">
						<Icon icon={checkIcon} width="20" height="20" />
						{saving ? 'Saving…' : 'Save changes'}
					</span>
				</Button>
			</div>
		</form>
	{/if}
</section>
