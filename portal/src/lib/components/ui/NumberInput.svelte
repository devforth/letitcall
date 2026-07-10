<script lang="ts">
	let {
		id,
		label,
		value = $bindable(''),
		min = 1,
		max,
		placeholder = '',
		disabled = false,
		required = false
	}: {
		id: string;
		label: string;
		value?: string;
		min?: number;
		max?: number;
		placeholder?: string;
		disabled?: boolean;
		required?: boolean;
	} = $props();
</script>

<label class="grid gap-2 text-sm" for={id}>
	<span class="font-medium">{label}</span>
	<input
		{id}
		type="number"
		bind:value
		{min}
		{max}
		{placeholder}
		{disabled}
		{required}
		step="1"
		class="min-h-11 w-full border border-black bg-white px-3 py-2 text-black outline-none placeholder:text-black/50 focus:ring-2 focus:ring-black focus:ring-offset-2 disabled:opacity-40"
	/>
</label>
