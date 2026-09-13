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

<div
	class="overflow-hidden rounded-lg"
	style="background: rgb(var(--color-foreground)); box-shadow: 0 0 0 1px rgb(var(--color-border));"
>
	<div class="overflow-x-auto">
		<table class="audit-log-table w-full min-w-[56rem] text-left text-sm">
			<thead>
				<tr>
					<th class="px-4 py-3 font-semibold">Avatar</th>
					<th class="px-4 py-3 font-semibold">User</th>
					<th class="px-4 py-3 font-semibold">Action</th>
					<th class="px-4 py-3 font-semibold">Date and time</th>
					<th class="px-4 py-3 text-right font-semibold"><span class="sr-only">Details</span></th>
				</tr>
			</thead>
			<tbody>
				{#each auditLogs as auditLog (auditLog.id)}
					<tr class:details-open={expanded.has(auditLog.id)}>
						<td class="px-4 py-3 align-top">
							<Avatar name={auditLog.actor.fullName} email={auditLog.actor.email} avatarPath={auditLog.actor.avatarPath} size={44} />
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
							<Button variant="ghost" size="small" onclick={() => toggle(auditLog.id)}>
								<Icon icon={expanded.has(auditLog.id) ? chevronUpIcon : chevronDownIcon} class="mr-2 size-4" />
								{expanded.has(auditLog.id) ? 'Hide details' : 'View details'}
							</Button>
						</td>
					</tr>
					{#if expanded.has(auditLog.id)}
						<tr>
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
</div>

<style>
	.audit-log-table {
		border-collapse: collapse;
	}

	.audit-log-table thead {
		background: rgb(var(--color-text) / 0.06);
	}

	.audit-log-table th {
		color: rgb(var(--color-text));
	}

	.audit-log-table thead th {
		border-bottom: 1px solid rgb(var(--color-border));
		font-size: 0.75rem;
		letter-spacing: 0.025em;
	}

	.audit-log-table tbody tr {
		border-bottom: 1px solid rgb(var(--color-border));
		transition: background 0.15s ease;
	}

	.audit-log-table tbody tr:last-child {
		border-bottom: 0;
	}

	.audit-log-table tbody tr.details-open {
		border-bottom: 0;
	}

	.audit-log-table tbody tr:not(.details-open):hover,
	.audit-log-table tbody tr.details-open:hover,
	.audit-log-table tbody tr.details-open:hover + tr,
	.audit-log-table tbody tr.details-open:has(+ tr:hover),
	.audit-log-table tbody tr.details-open + tr:hover {
		background: rgb(var(--color-primary) / 0.045);
	}

	.audit-log-table td {
		color: rgb(var(--color-text));
	}
</style>
