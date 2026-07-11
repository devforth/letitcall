<script lang="ts">
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';

	let {
		idPrefix,
		emails = $bindable(),
		limit = null
	}: {
		idPrefix: string;
		emails: string[];
		limit?: number | null;
	} = $props();

	const canAdd = $derived(limit === null || emails.length < limit);

	function addGuest() {
		emails.push('');
	}

	function removeGuest(index: number) {
		emails.splice(index, 1);
	}
</script>

<fieldset class="grid gap-4">
	<legend class="text-sm font-medium">Additional guests</legend>
	{#each emails as _, index (`${idPrefix}-${index}`)}
		<div class="grid gap-2 sm:grid-cols-[1fr_auto] sm:items-end">
			<Input id={`${idPrefix}-${index}`} label={`Guest ${index + 1} email`} type="email" bind:value={emails[index]} required autocomplete="off" />
			<Button variant="secondary" onclick={() => removeGuest(index)}>Remove</Button>
		</div>
	{/each}
	{#if canAdd}
		<div><Button variant="secondary" onclick={addGuest}>Add guest</Button></div>
	{/if}
</fieldset>
