<script lang="ts">
	let {
		id,
		label,
		options,
		value = $bindable(''),
		placeholder = 'Search…',
		required = false,
		disabled = false
	}: {
		id: string;
		label: string;
		options: string[];
		value?: string;
		placeholder?: string;
		required?: boolean;
		disabled?: boolean;
	} = $props();
</script>

<label class="grid gap-2 text-sm" for={id}>
	<span class="font-medium">{label}</span>
	<input
		{id}
		list={`${id}-options`}
		type="search"
		bind:value
		{placeholder}
		{required}
		{disabled}
		autocomplete="off"
		class="min-h-11 w-full border border-black bg-white px-3 py-2 text-black outline-none placeholder:text-black/50 focus:ring-2 focus:ring-black focus:ring-offset-2 disabled:opacity-40"
	/>
	<datalist id={`${id}-options`}>
		{#each options as option (option)}
			<option value={option}></option>
		{/each}
	</datalist>
</label>
