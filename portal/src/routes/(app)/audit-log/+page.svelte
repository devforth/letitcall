<script lang="ts">
	import { onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import historyIcon from '@iconify-icons/tabler/history';
	import { callApi } from '$lib/api';
	import type { AuditLog } from '$lib/types';
	import AuditLogTable from '$lib/components/AuditLogTable.svelte';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';

	let auditLogs = $state<AuditLog[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			auditLogs = (await callApi<{ auditLogs: AuditLog[] }>('/api/audit-logs')).auditLogs;
		} catch (cause) {
			error = cause instanceof Error ? cause.message : 'Unable to load audit logs';
		} finally {
			loading = false;
		}
	});
</script>

<PageTitle title="Audit log" />

<section aria-labelledby="audit-log-title" class="flex flex-col gap-6">
	<div class="mb-2">
		<PageHeader
			id="audit-log-title"
			title="Audit log"
			description="Immutable history of backoffice changes. Dates and times use your local timezone."
			icon={historyIcon}
		/>
	</div>

	{#if error}
		<p class="outlined-block p-4 text-sm" role="alert">{error}</p>
	{:else if loading}
		<p class="outlined-block p-6 text-sm">Loading audit log…</p>
	{:else}
		<AuditLogTable {auditLogs} />
	{/if}
</section>

<style>
	.outlined-block {
		border: 0;
		border-radius: 8px;
		background: rgb(var(--color-background));
		box-shadow: 0 0 0 1px var(--color-border);
	}
</style>
