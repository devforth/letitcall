<script lang="ts">
	import type { ManagedUser } from '$lib/types';
	import Avatar from '$lib/components/ui/Avatar.svelte';

	let {
		users,
		required = $bindable([]),
		optional = $bindable([]),
		error = '',
		onchange
	}: {
		users: ManagedUser[];
		required?: string[];
		optional?: string[];
		error?: string;
		onchange?: () => void;
	} = $props();

	function role(email: string) {
		if (required.includes(email)) return 'required';
		if (optional.includes(email)) return 'optional';
		return 'none';
	}

	function setRole(email: string, value: string) {
		required = value === 'required'
			? [...required.filter((candidate) => candidate !== email), email].sort()
			: required.filter((candidate) => candidate !== email);
		optional = value === 'optional'
			? [...optional.filter((candidate) => candidate !== email), email].sort()
			: optional.filter((candidate) => candidate !== email);
		onchange?.();
	}
</script>

<fieldset
	id="hosts"
	tabindex="-1"
	aria-describedby={error ? 'hosts-error' : 'hosts-description'}
	class="rounded-lg border-2 p-4 outline-none focus:ring-2 focus:ring-[rgb(var(--color-primary))]/25 focus:ring-offset-2 sm:p-5"
	style="background: rgb(var(--color-foreground)); border-color: rgb(var(--color-border)); box-shadow: var(--shadow-small);"
>
	<legend class="px-2 text-sm font-semibold" style="color: rgb(var(--color-text));">Hosts</legend>
	<p id="hosts-description" class="mb-4 text-sm" style="color: rgb(var(--color-text) / 0.65);">Required hosts determine availability. Optional hosts receive the booking without blocking a time.</p>
	<div class="grid gap-2">
		{#each users as user (user.email)}
			<div class="grid min-h-14 gap-3 rounded-lg border p-3 sm:grid-cols-[1fr_auto] sm:items-center" style="border-color: rgb(var(--color-border)); background: rgb(var(--color-text) / 0.035);">
				<div class="flex min-w-0 items-center gap-3">
					<Avatar name={user.fullName} email={user.email} avatarPath={user.avatarPath} size={36} rounded="lg" ring />
					<span class="min-w-0 text-sm">
						<span class="block truncate font-medium" style="color: rgb(var(--color-text));">{user.email}</span>
						<span class="block text-xs" style="color: rgb(var(--color-text) / 0.65);">{user.googleConnected ? 'Google Calendar connected' : 'Email only'}</span>
					</span>
				</div>
				<div class="flex flex-wrap gap-3 text-xs" style="color: rgb(var(--color-text) / 0.75);">
					{#each [['required', 'Required'], ['optional', 'Optional'], ['none', 'Not a host']] as option}
						<label class="flex cursor-pointer items-center gap-1.5">
							<input
								type="radio"
								name={'host-' + user.email}
								value={option[0]}
								checked={role(user.email) === option[0]}
								onchange={() => setRole(user.email, option[0])}
								class="size-4 appearance-none rounded-full border-2 border-[rgb(var(--color-border))] bg-[rgb(var(--color-foreground))] checked:border-[rgb(var(--color-primary))] checked:bg-[rgb(var(--color-primary))] checked:ring-2 checked:ring-[rgb(var(--color-foreground))] checked:ring-inset focus:outline-none focus:ring-2 focus:ring-[rgb(var(--color-primary))]/25 focus:ring-offset-2"
							/>
							{option[1]}
						</label>
					{/each}
				</div>
			</div>
		{/each}
	</div>
	{#if error}
		<p id="hosts-error" class="mt-3 text-xs font-semibold" style="color: rgb(var(--error));" role="alert">{error}</p>
	{/if}
</fieldset>
