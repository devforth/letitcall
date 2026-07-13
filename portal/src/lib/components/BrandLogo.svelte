<script lang="ts">
	import { logoURL } from '$lib/api';
	import { branding } from '$lib/stores/branding.svelte';

	let { class: className = '' }: { class?: string } = $props();
</script>

{#if branding.logoPath}
	<img src={logoURL(branding.logoPath)} alt={`${branding.name} logo`} class={className} />
{/if}
