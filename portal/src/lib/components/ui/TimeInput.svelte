<script lang="ts">
	let {
		id,
		label,
		value = $bindable(''),
		disabled = false,
		onchange
	}: {
		id: string;
		label: string;
		value?: string;
		disabled?: boolean;
		onchange?: (value: string) => void;
	} = $props();
</script>

<label class="grid gap-1.5 text-sm" for={id}>
	<span class="font-medium" style="color: rgb(var(--color-text));">{label}</span>
	<input
		{id}
		type="time"
		bind:value
		oninput={(event) => onchange?.(event.currentTarget.value)}
		{disabled}
		step="900"
		class="min-h-11 w-full rounded-[10px] border-2 bg-[rgb(var(--color-foreground))] px-3 py-2 text-[rgb(var(--color-text))] outline-none transition focus:border-[rgb(var(--color-primary))] focus:ring-2 focus:ring-[rgb(var(--color-primary))]/25 disabled:cursor-not-allowed disabled:opacity-40"
	/>
</label>
