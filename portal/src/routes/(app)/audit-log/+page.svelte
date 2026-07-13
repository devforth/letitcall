<script lang="ts">
	import { onMount } from 'svelte';
	import { callApi } from '$lib/api';
	import type { AuditLog } from '$lib/types';
	import AuditLogTable from '$lib/components/AuditLogTable.svelte';
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

<section aria-labelledby="audit-log-title">
	<div class="mb-6">
		<h1 id="audit-log-title" class="text-2xl font-semibold tracking-tight">Audit log</h1>
		<p class="mt-2 text-sm">Immutable history of backoffice changes. Dates and times use your local timezone.</p>
	</div>

	{#if error}
		<p class="border border-black p-4 text-sm" role="alert">{error}</p>
	{:else if loading}
		<p class="border border-black p-6 text-sm">Loading audit log…</p>
	{:else}
		<AuditLogTable {auditLogs} />
	{/if}
</section>
