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

<label class="grid gap-1.5 text-sm" for={id}>
	<span class="font-medium" style="color: rgb(var(--color-text));">{label}</span>
	<select
		{id}
		bind:value
		{disabled}
		class="min-h-11 w-full rounded-[10px] border-2 bg-[rgb(var(--color-foreground))] px-3 py-2 text-[rgb(var(--color-text))] outline-none transition focus:border-[rgb(var(--color-primary))] focus:ring-2 focus:ring-[rgb(var(--color-primary))]/25 disabled:cursor-not-allowed disabled:opacity-40"
	>
		{#each options as option (option.value)}
			<option value={option.value}>{option.label}</option>
		{/each}
	</select>
</label>
