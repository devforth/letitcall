<script lang="ts">
	import type { AuditLog } from '$lib/types';
	import AuditPayload from '$lib/components/AuditPayload.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import SegmentedControl from '$lib/components/ui/SegmentedControl.svelte';

	let { auditLogs }: { auditLogs: AuditLog[] } = $props();
	let expanded = $state(new Set<string>());
	const detailTabs = [
		{ value: 'overview', label: 'Overview' },
		{ value: 'details', label: 'Details' }
	];

	function setDetails(id: string, visible: boolean) {
		const next = new Set(expanded);
		if (visible) next.add(id);
		else next.delete(id);
		expanded = next;
	}

	function words(value: string): string {
		return value.replaceAll('_', ' ');
	}

	function title(value: string): string {
		const label = words(value);
		return label.charAt(0).toUpperCase() + label.slice(1);
	}

	function resourceLabel(auditLog: AuditLog): string {
		if (auditLog.resource === 'api_token') return auditLog.payload.name as string;
		return auditLog.resourceId;
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
	style="background: rgb(var(--color-background)); box-shadow: 0 0 0 1px var(--color-border);"
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
							<div class="font-medium">{auditLog.actor.fullName}</div>
							<div class="mt-1 text-xs">{auditLog.actor.email}</div>
						</td>
						<td class="px-4 py-3 align-top">
							<div class="font-medium">{title(auditLog.action)}</div>
							<div class="mt-1 text-xs">{title(auditLog.resource)} · {resourceLabel(auditLog)}</div>
						</td>
						<td class="px-4 py-3 align-top">
							<time datetime={auditLog.createdAt}>
								<span class="block font-medium">{localDate(auditLog.createdAt)}</span>
								<span class="mt-1 block text-xs">{localTime(auditLog.createdAt)}</span>
							</time>
						</td>
						<td class="px-4 py-3 text-right align-top">
							<SegmentedControl
								options={detailTabs}
								value={expanded.has(auditLog.id) ? 'details' : 'overview'}
								label="Audit log details"
								onchange={(value) => setDetails(auditLog.id, value === 'details')}
							/>
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
					<tr><td class="empty-table-cell" colspan="5">No backoffice mutations have been logged</td></tr>
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
		border-bottom: 1px solid var(--color-border);
		font-size: 0.75rem;
		letter-spacing: 0.025em;
	}

	.audit-log-table tbody tr {
		border-bottom: 1px solid var(--color-border);
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

	.empty-table-cell {
		padding: 2.5rem 1rem;
		color: rgb(var(--color-text) / 0.65) !important;
		font-size: 0.875rem;
		text-align: center;
	}

</style>
