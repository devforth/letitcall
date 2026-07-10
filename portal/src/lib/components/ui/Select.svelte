<script lang="ts">
	let {
		id,
		label,
		options,
		value = $bindable(''),
		disabled = false
	}: {
		id: string;
		label: string;
		options: { value: string; label: string }[];
		value?: string;
		disabled?: boolean;
	} = $props();
</script>

<label class="grid gap-2 text-sm" for={id}>
	<span class="font-medium">{label}</span>
	<select
		{id}
		bind:value
		{disabled}
		class="min-h-11 w-full border border-black bg-white px-3 py-2 text-black outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 disabled:opacity-40"
	>
		{#each options as option (option.value)}
			<option value={option.value}>{option.label}</option>
		{/each}
	</select>
</label>
