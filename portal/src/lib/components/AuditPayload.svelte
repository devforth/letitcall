<script lang="ts">
	let {
		action,
		payload
	}: {
		action: string;
		payload: Record<string, unknown>;
	} = $props();

	type Change = { before: unknown; after: unknown };

	function isChange(value: unknown): value is Change {
		return typeof value === 'object' && value !== null && 'before' in value && 'after' in value;
	}

	function formatValue(value: unknown): string {
		if (value === null) return 'null';
		if (typeof value === 'string') return value || '""';
		if (typeof value === 'number' || typeof value === 'boolean') return String(value);
		return JSON.stringify(value, null, 2);
	}

	const fields = $derived(Object.entries(payload));
	const showsDiff = $derived(action === 'edited' && fields.every(([, value]) => isChange(value)));
</script>

<div
	class="audit-payload overflow-x-auto rounded-md"
	style="background: rgb(var(--color-background)); box-shadow: 0 0 0 1px var(--color-border);"
>
	<table class="w-full min-w-[38rem] text-left text-sm">
		<thead>
			<tr>
				<th class="px-4 py-3 font-semibold">Field</th>
				{#if showsDiff}
					<th class="px-4 py-3 font-semibold">Previous value</th>
					<th class="px-4 py-3 font-semibold">New value</th>
				{:else}
					<th class="px-4 py-3 font-semibold">Value</th>
				{/if}
			</tr>
		</thead>
		<tbody>
			{#each fields as [field, value] (field)}
				<tr>
					<th class="px-4 py-3 align-top font-medium">{field}</th>
					{#if showsDiff && isChange(value)}
						<td class="px-4 py-3 align-top"><pre class="whitespace-pre-wrap break-words font-mono text-xs">{formatValue(value.before)}</pre></td>
						<td class="px-4 py-3 align-top"><pre class="whitespace-pre-wrap break-words font-mono text-xs">{formatValue(value.after)}</pre></td>
					{:else}
						<td class="px-4 py-3 align-top"><pre class="whitespace-pre-wrap break-words font-mono text-xs">{formatValue(value)}</pre></td>
					{/if}
				</tr>
			{:else}
				<tr><td class="empty-table-cell" colspan={showsDiff ? 3 : 2}>No fields changed</td></tr>
			{/each}
		</tbody>
	</table>
</div>

<style>
	.audit-payload table {
		border-collapse: collapse;
	}

	.audit-payload thead {
		background: rgb(var(--color-text) / 0.06);
	}

	.audit-payload thead th {
		border-bottom: 1px solid var(--color-border);
		color: rgb(var(--color-text));
		font-size: 0.75rem;
		letter-spacing: 0.025em;
	}

	.audit-payload tbody tr {
		border-bottom: 1px solid var(--color-border);
	}

	.audit-payload tbody tr:last-child {
		border-bottom: 0;
	}

	.audit-payload td,
	.audit-payload tbody th {
		color: rgb(var(--color-text));
	}

	.empty-table-cell {
		padding: 2.5rem 1rem;
		color: rgb(var(--color-text) / 0.65) !important;
		font-size: 0.875rem;
		text-align: center;
	}
</style>
