<script lang="ts">
	import { branding } from '$lib/stores/branding.svelte';

	let { title = '' }: { title?: string } = $props();
</script>

<svelte:head><title>{title ? `${title} · ${branding.name}` : branding.name}</title></svelte:head>
