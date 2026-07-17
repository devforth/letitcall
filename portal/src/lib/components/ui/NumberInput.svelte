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

<label class="grid gap-1.5 text-sm" for={id}>
	<span class="font-medium" style="color: rgb(var(--color-text));">{label}</span>
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
		class="min-h-11 w-full rounded-[10px] border-2 bg-[rgb(var(--color-foreground))] px-3 py-2 text-[rgb(var(--color-text))] outline-none transition placeholder:text-[rgb(var(--color-text))]/40 focus:border-[rgb(var(--color-primary))] focus:ring-2 focus:ring-[rgb(var(--color-primary))]/25 disabled:cursor-not-allowed disabled:opacity-40"
	/>
</label>
