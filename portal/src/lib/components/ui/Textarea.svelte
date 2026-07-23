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

<label class="grid gap-1.5 text-sm" for={id}>
	<span class="font-medium" style="color: rgb(var(--color-text));">{label}</span>
	<textarea
		{id}
		bind:value
		{placeholder}
		{required}
		{disabled}
		{maxlength}
		{rows}
		class="w-full resize-y rounded-[10px] border-2 border-[rgb(var(--color-border))] bg-[rgb(var(--color-foreground))] px-3 py-2 text-[rgb(var(--color-text))] outline-none transition placeholder:text-[rgb(var(--color-text))]/40 focus:border-[rgb(var(--color-primary))] focus:ring-2 focus:ring-[rgb(var(--color-primary))]/25 disabled:cursor-not-allowed disabled:opacity-40"
	></textarea>
</label>
