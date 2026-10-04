<script lang="ts">
	import Icon from '@iconify/svelte';
	import checkIcon from '@iconify-icons/tabler/check';
	import userPlusIcon from '@iconify-icons/tabler/user-plus';
	import worldIcon from '@iconify-icons/tabler/world';
	import xIcon from '@iconify-icons/tabler/x';
	import { callApi, appPath } from '$lib/api';
	import ImageSelector from '$lib/components/ImageSelector.svelte';
	import LeaveGuard from '$lib/components/LeaveGuard.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import StickyActions from '$lib/components/StickyActions.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import SearchableSelect from '$lib/components/ui/SearchableSelect.svelte';
	import { getLocalTimezones } from '$lib/timezones';

	const localTimezones = getLocalTimezones();
	let email = $state('');
	let fullName = $state('');
	let password = $state('');
	let timezone = $state(localTimezones.current);
	let avatarSelector = $state<ImageSelector | null>(null);
	let avatarChanged = $state(false);
	let leaveGuard: LeaveGuard;
	const changed = $derived(
		!!(email || fullName || password) || timezone !== localTimezones.current || avatarChanged
	);
	let saving = $state(false);
	let error = $state('');

	async function createUser(event: SubmitEvent) {
		event.preventDefault();
		saving = true;
		error = '';
		try {
			const avatar = await avatarSelector?.exportImage();
			await callApi('/api/users', {
				method: 'POST',
				body: JSON.stringify({ email, fullName, password, timezone, avatar })
			});
			await leaveGuard.leave(appPath('/users'));
		} catch (cause) {
			error = cause instanceof Error ? cause.message : 'Unable to create user';
		} finally {
			saving = false;
		}
	}
</script>

<PageTitle title="New user" />

<section aria-labelledby="new-user-title">
	<div class="mb-8">
		<PageHeader
			id="new-user-title"
			title="New user"
			description="Invite someone to sign in and host events."
			icon={userPlusIcon}
			parent={{ href: appPath('/users'), label: 'Users' }}
		/>
	</div>

	{#if error}
		<p class="mb-5 border border-black p-3 text-sm" role="alert">{error}</p>
	{/if}

	<form class="grid gap-5 lg:grid-cols-2" onsubmit={createUser}>
		<Input id="new-email" label="Email" type="email" bind:value={email} required autocomplete="off" />
		<SearchableSelect
			id="new-timezone"
			emptyText="No matching timezones"
			label="Timezone"
			icon={worldIcon}
			options={localTimezones.options}
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
			<ImageSelector id="new-avatar" legend="Avatar (optional)" onchange={() => (avatarChanged = true)} bind:this={avatarSelector} />
		</div>
		<StickyActions class="lg:col-span-2">
			<Button variant="primary-outline" class="outlined-action-button" onclick={() => leaveGuard.leave(appPath('/users'))}>
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
		</StickyActions>
	</form>
</section>

<LeaveGuard
	bind:this={leaveGuard}
	{changed}
	title="Leave without creating?"
	description="This user has not been created yet. Everything you entered will be lost."
/>
