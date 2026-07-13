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

<div class="overflow-x-auto border border-black bg-white text-black">
	<table class="w-full min-w-[38rem] border-collapse text-left text-sm">
		<thead>
			<tr class="border-b border-black">
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
				<tr class="border-b border-black last:border-b-0">
					<th class="px-4 py-3 align-top font-medium">{field}</th>
					{#if showsDiff && isChange(value)}
						<td class="px-4 py-3 align-top"><pre class="whitespace-pre-wrap break-words font-mono text-xs">{formatValue(value.before)}</pre></td>
						<td class="px-4 py-3 align-top"><pre class="whitespace-pre-wrap break-words font-mono text-xs">{formatValue(value.after)}</pre></td>
					{:else}
						<td class="px-4 py-3 align-top"><pre class="whitespace-pre-wrap break-words font-mono text-xs">{formatValue(value)}</pre></td>
					{/if}
				</tr>
			{:else}
				<tr><td class="px-4 py-5 text-center" colspan={showsDiff ? 3 : 2}>No fields changed.</td></tr>
			{/each}
		</tbody>
	</table>
</div>
