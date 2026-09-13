<script lang="ts">
	import { onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import historyIcon from '@iconify-icons/tabler/history';
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

<section aria-labelledby="audit-log-title" class="flex flex-col gap-6">
	<div class="mb-2">
		<div class="flex min-w-0 items-center gap-4">
			<div
				class="grid size-12 shrink-0 place-items-center rounded-lg"
				style="background: rgb(var(--color-primary) / 0.12); color: rgb(var(--color-primary));"
			>
				<Icon icon={historyIcon} width="24" height="24" />
			</div>
			<div>
				<h1 id="audit-log-title" class="text-2xl font-semibold tracking-tight" style="color: rgb(var(--color-text));">Audit log</h1>
				<p class="text-sm" style="color: rgb(var(--color-text) / 0.65);">Immutable history of backoffice changes. Dates and times use your local timezone.</p>
			</div>
		</div>
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
		background: rgb(var(--color-foreground));
		box-shadow: 0 0 0 1px rgb(var(--color-border));
	}
</style>
