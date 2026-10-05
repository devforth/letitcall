<script lang="ts">
	import sortIcon from '@iconify-icons/tabler/arrows-sort';
	import { nextSort } from '$lib/sort';
	import type { AuditLog } from '$lib/types';
	import { browserTimeFormatter } from '$lib/time-format';
	import AuditPayload from '$lib/components/AuditPayload.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import SearchableSelect from '$lib/components/ui/SearchableSelect.svelte';
	import SegmentedControl from '$lib/components/ui/SegmentedControl.svelte';
	import SortButton from '$lib/components/ui/SortButton.svelte';

	type SortKey = 'user' | 'action' | 'createdAt';

	let { auditLogs }: { auditLogs: AuditLog[] } = $props();
	let sort = $state('createdAt-descending');
	const sortKey = $derived(sort.split('-')[0] as SortKey);
	const sortDirection = $derived(sort.split('-')[1] as 'ascending' | 'descending');
	const sortedLogs = $derived([...auditLogs].sort(compareLogs));
	const columns: { key: SortKey; label: string }[] = [
		{ key: 'user', label: 'User' },
		{ key: 'action', label: 'Action' },
		{ key: 'createdAt', label: 'Date and time' }
	];
	const sortOptions = [
		{ value: 'createdAt-descending', label: 'Newest first' },
		{ value: 'createdAt-ascending', label: 'Oldest first' },
		{ value: 'user-ascending', label: 'User, A-Z' },
		{ value: 'user-descending', label: 'User, Z-A' },
		{ value: 'action-ascending', label: 'Action, A-Z' },
		{ value: 'action-descending', label: 'Action, Z-A' }
	];
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

	function sortValue(auditLog: AuditLog): string {
		if (sortKey === 'user') return auditLog.actor.fullName || auditLog.actor.email;
		if (sortKey === 'action') return `${auditLog.action} ${auditLog.resource}`;
		return auditLog.createdAt;
	}

	function compareLogs(first: AuditLog, second: AuditLog) {
		const comparison = sortValue(first).localeCompare(sortValue(second)) || second.createdAt.localeCompare(first.createdAt);
		return sortDirection === 'ascending' ? comparison : -comparison;
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
		return browserTimeFormatter({ timeStyle: 'medium' }).format(new Date(value));
	}
</script>

<div class="audit-log">
	<div class="audit-sort">
		<SearchableSelect id="audit-sort" label="Sort by" icon={sortIcon} options={sortOptions} bind:value={sort} clearable={false} required />
	</div>
	<div class="audit-panel">
		<div class="overflow-x-auto">
			<table class="audit-log-table w-full min-w-[56rem] text-left text-sm">
				<thead>
					<tr>
						<th class="px-4 py-3 font-semibold">Avatar</th>
						{#each columns as column (column.key)}
							<th aria-sort={sortKey === column.key ? sortDirection : 'none'} class="px-4 py-3 font-semibold">
								<SortButton label={column.label} active={sortKey === column.key} direction={sortDirection} onclick={() => (sort = nextSort(sort, column.key))} />
							</th>
						{/each}
						<th class="px-4 py-3 text-right font-semibold"><span class="sr-only">Details</span></th>
					</tr>
				</thead>
				<tbody>
					{#each sortedLogs as auditLog (auditLog.id)}
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
							<tr class="details-row">
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
</div>

<style>
	.audit-log-table {
		border-collapse: collapse;
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
	}

	.audit-log-table tbody tr:last-child {
		border-bottom: 0;
	}

	.audit-log-table tbody tr.details-open {
		border-bottom: 0;
	}

	.audit-log-table td {
		color: rgb(var(--color-text));
	}

	.audit-log {
		container: audit / inline-size;
		display: grid;
		gap: 1.25rem;
	}

	.audit-panel {
		overflow: hidden;
		border-radius: 0.5rem;
		background: rgb(var(--color-background));
		box-shadow: 0 0 0 1px var(--color-border);
	}

	.audit-sort {
		display: none;
	}

	.empty-table-cell {
		padding: 2.5rem 1rem;
		color: rgb(var(--color-text) / 0.65) !important;
		font-size: 0.875rem;
		text-align: center;
	}

	/* Below the table's 56rem minimum it would scroll, so every entry becomes a card. */
	@container audit (max-width: 56rem) {
		.audit-sort {
			display: block;
		}

		.audit-panel {
			overflow: visible;
			background: none;
			box-shadow: none;
		}

		.audit-log-table,
		.audit-log-table tbody {
			display: block;
			min-width: 0;
		}

		.audit-log-table thead {
			display: none;
		}

		.audit-log-table tbody tr {
			display: grid;
			grid-template-columns: auto minmax(0, 1fr);
			grid-template-areas: 'avatar user' 'avatar action' 'date date' 'details details';
			gap: 0.5rem 1rem;
			margin-bottom: 1.25rem;
			border: 0;
			border-radius: 0.5rem;
			padding: 1rem 1.25rem 1.25rem;
			background: rgb(var(--color-background));
			/* Inset: the table's scroll wrapper clips anything drawn outside the row. */
			box-shadow: inset 0 0 0 1px var(--color-border);
		}

		.audit-log-table tbody td {
			padding: 0;
		}

		.audit-log-table tbody td:nth-child(1) {
			grid-area: avatar;
		}

		.audit-log-table tbody td:nth-child(2) {
			grid-area: user;
		}

		.audit-log-table tbody td:nth-child(3) {
			grid-area: action;
		}

		.audit-log-table tbody td:nth-child(4) {
			grid-area: date;
		}

		.audit-log-table tbody td:nth-child(4) time span {
			display: inline;
			margin: 0 0.375rem 0 0;
		}

		.audit-log-table tbody td:nth-child(5) {
			grid-area: details;
			text-align: left;
		}

		/* The open entry and its details read as one card. */
		.audit-log-table tbody tr.details-open {
			margin-bottom: 0;
			border-radius: 0.5rem 0.5rem 0 0;
			box-shadow:
				inset 1px 0 0 var(--color-border),
				inset -1px 0 0 var(--color-border),
				inset 0 1px 0 var(--color-border);
		}

		.audit-log-table tbody tr.details-row {
			display: block;
			padding-top: 0;
			border-radius: 0 0 0.5rem 0.5rem;
			box-shadow:
				inset 1px 0 0 var(--color-border),
				inset -1px 0 0 var(--color-border),
				inset 0 -1px 0 var(--color-border);
		}

		.audit-log-table tbody tr.details-row td {
			display: block;
		}
	}
</style>
