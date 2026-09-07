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
		class="m-[3px] min-h-11 w-[calc(100%_-_6px)] rounded-[10px] border-0 bg-[rgb(var(--color-foreground))] px-3 py-2 text-[rgb(var(--color-text))] shadow-[0_0_0_1px_rgb(var(--color-border))] outline-none transition focus:shadow-[0_0_0_1px_rgb(var(--color-primary)),0_0_0_3px_rgb(var(--color-primary)/0.25)] disabled:cursor-not-allowed disabled:opacity-40"
	>
		{#each options as option (option.value)}
			<option value={option.value}>{option.label}</option>
		{/each}
	</select>
</label>
