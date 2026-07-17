<script lang="ts">
	import SearchableSelect from '$lib/components/ui/SearchableSelect.svelte';
	let value = $state('Europe/Kiev');
	const options = ['Europe/Kiev', 'Europe/London', 'America/New_York'];
</script>

<div style="padding:32px; display:grid; grid-template-columns:1fr 1fr; gap:16px; align-items:stretch; max-width:640px;">
	<SearchableSelect id="tz" label="Timezone" {options} bind:value />
	<div style="background:#555; border-radius:10px; padding:16px; min-height:110px;">taller sibling forces the row to stretch</div>
</div>
