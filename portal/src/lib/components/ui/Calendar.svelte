<script lang="ts">
	let {
		id,
		label,
		value = $bindable(''),
		min,
		max,
		required = false,
		disabled = false
	}: {
		id: string;
		label: string;
		value?: string;
		min?: string;
		max?: string;
		required?: boolean;
		disabled?: boolean;
	} = $props();
</script>

<label class="grid gap-2 text-sm" for={id}>
	<span class="font-medium">{label}</span>
	<input
		{id}
		type="date"
		bind:value
		{min}
		{max}
		{required}
		{disabled}
		class="min-h-11 w-full border border-black bg-white px-3 py-2 text-black outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 disabled:opacity-40"
	/>
</label>
