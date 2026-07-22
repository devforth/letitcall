<script lang="ts">
	import Icon from '@iconify/svelte';
	import chevronDownIcon from '@iconify-icons/tabler/chevron-down';
	import chevronUpIcon from '@iconify-icons/tabler/chevron-up';
	import type { AuditLog } from '$lib/types';
	import AuditPayload from '$lib/components/AuditPayload.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';

	let { auditLogs }: { auditLogs: AuditLog[] } = $props();
	let expanded = $state(new Set<string>());

	function toggle(id: string) {
		const next = new Set(expanded);
		if (next.has(id)) next.delete(id);
		else next.add(id);
		expanded = next;
	}

	function words(value: string): string {
		return value.replaceAll('_', ' ');
	}

	function title(value: string): string {
		const label = words(value);
		return label.charAt(0).toUpperCase() + label.slice(1);
	}

	function localDate(value: string): string {
		return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(new Date(value));
	}

	function localTime(value: string): string {
		return new Intl.DateTimeFormat(undefined, { timeStyle: 'medium' }).format(new Date(value));
	}
</script>

<div class="overflow-x-auto border border-black">
	<table class="w-full min-w-[56rem] border-collapse text-left text-sm">
		<thead>
			<tr class="border-b border-black">
				<th class="px-4 py-3 font-semibold">Avatar</th>
				<th class="px-4 py-3 font-semibold">User</th>
				<th class="px-4 py-3 font-semibold">Action</th>
				<th class="px-4 py-3 font-semibold">Date and time</th>
				<th class="px-4 py-3 text-right font-semibold"><span class="sr-only">Details</span></th>
			</tr>
		</thead>
		<tbody>
			{#each auditLogs as auditLog (auditLog.id)}
				<tr class="border-b border-black">
					<td class="px-4 py-3 align-top">
						<Avatar name={auditLog.actor.fullName} email={auditLog.actor.email} avatarPath={auditLog.actor.avatarPath} size={44} rounded="none" class="border border-black" />
					</td>
					<td class="px-4 py-3 align-top">
						<div class="font-medium">{auditLog.actor.fullName || '—'}</div>
						<div class="mt-1 text-xs">{auditLog.actor.email}</div>
					</td>
					<td class="px-4 py-3 align-top">
						<div class="font-medium">{title(auditLog.action)}</div>
						<div class="mt-1 text-xs">{title(auditLog.resource)} · {auditLog.resourceId}</div>
					</td>
					<td class="px-4 py-3 align-top">
						<time datetime={auditLog.createdAt}>
							<span class="block font-medium">{localDate(auditLog.createdAt)}</span>
							<span class="mt-1 block text-xs">{localTime(auditLog.createdAt)}</span>
						</time>
					</td>
					<td class="px-4 py-3 text-right align-top">
						<Button variant="secondary" onclick={() => toggle(auditLog.id)}>
							<Icon icon={expanded.has(auditLog.id) ? chevronUpIcon : chevronDownIcon} class="mr-2 size-4" />
							{expanded.has(auditLog.id) ? 'Hide details' : 'View details'}
						</Button>
					</td>
				</tr>
				{#if expanded.has(auditLog.id)}
					<tr class="border-b border-black">
						<td class="p-4" colspan="5">
							<AuditPayload action={auditLog.action} payload={auditLog.payload} />
						</td>
					</tr>
				{/if}
			{:else}
				<tr><td class="px-4 py-8 text-center" colspan="5">No backoffice mutations have been logged.</td></tr>
			{/each}
		</tbody>
	</table>
</div>
