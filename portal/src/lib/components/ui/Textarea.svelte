<script lang="ts">
	let {
		id,
		label,
		value = $bindable(''),
		placeholder = '',
		required = false,
		disabled = false,
		maxlength,
		rows = 5
	}: {
		id: string;
		label: string;
		value?: string;
		placeholder?: string;
		required?: boolean;
		disabled?: boolean;
		maxlength?: number;
		rows?: number;
	} = $props();
</script>

<label class="grid gap-2 text-sm" for={id}>
	<span class="font-medium">{label}</span>
	<textarea
		{id}
		bind:value
		{placeholder}
		{required}
		{disabled}
		{maxlength}
		{rows}
		class="w-full resize-y border border-black bg-white px-3 py-2 text-black outline-none placeholder:text-black/50 focus:ring-2 focus:ring-black focus:ring-offset-2 disabled:opacity-40"
	></textarea>
</label>
