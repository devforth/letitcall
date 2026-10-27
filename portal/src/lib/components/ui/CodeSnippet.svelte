<script lang="ts">
	import CopyButton from '$lib/components/ui/CopyButton.svelte';

	let { value, label }: { value: string; label: string } = $props();
</script>

<div class="code-snippet">
	<code>{value}</code>
	<CopyButton {value} {label} />
</div>

<style>
	.code-snippet {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	code {
		min-width: 0;
		flex: 1;
		overflow-x: auto;
		border-radius: 0.5rem;
		padding: 0.5rem 0.75rem;
		background: rgb(var(--color-text) / 0.08);
		color: rgb(var(--color-text) / 0.85);
		font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
		font-size: 0.875rem;
		white-space: pre;
	}
</style>
