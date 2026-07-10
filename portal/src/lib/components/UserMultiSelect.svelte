<script lang="ts">
	import { avatarURL } from '$lib/api';
	import type { ManagedUser } from '$lib/types';

	let {
		users,
		selected = $bindable([]),
		error = '',
		onchange
	}: {
		users: ManagedUser[];
		selected?: string[];
		error?: string;
		onchange?: () => void;
	} = $props();

	function toggle(email: string, checked: boolean) {
		selected = checked ? [...selected, email].sort() : selected.filter((value) => value !== email);
		onchange?.();
	}
</script>

<fieldset
	id="recipients"
	tabindex="-1"
	aria-describedby={error ? 'recipients-error' : 'recipients-description'}
	class="border border-black p-4 outline-none focus:ring-2 focus:ring-black focus:ring-offset-2"
>
	<legend class="px-2 text-sm font-medium">Recipients</legend>
	<p id="recipients-description" class="mb-3 text-xs">Choose at least one user. Each selected user receives booking email.</p>
	<div class="grid gap-2 sm:grid-cols-2">
		{#each users as user (user.email)}
			<label class="flex min-h-14 cursor-pointer items-center gap-3 border border-black p-3">
				<input
					type="checkbox"
					checked={selected.includes(user.email)}
					onchange={(event) => toggle(user.email, event.currentTarget.checked)}
					class="size-5 shrink-0 appearance-none border border-black bg-white checked:bg-black focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2"
				/>
				{#if user.avatarPath}
					<img src={avatarURL(user.avatarPath)} alt="" class="size-9 border border-black object-cover" />
				{:else}
					<span class="grid size-9 place-items-center border border-black" aria-hidden="true">—</span>
				{/if}
				<span class="min-w-0 text-sm">
					<span class="block truncate font-medium">{user.email}</span>
					<span class="block text-xs">{user.googleConnected ? 'Google Calendar connected' : 'Email only'}</span>
				</span>
			</label>
		{/each}
	</div>
	{#if error}
		<p id="recipients-error" class="mt-3 text-xs" role="alert">{error}</p>
	{/if}
</fieldset>
