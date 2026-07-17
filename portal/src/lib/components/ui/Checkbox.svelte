<script lang="ts">
	let {
		id,
		label,
		checked = $bindable(false),
		disabled = false
	}: { id: string; label: string; checked?: boolean; disabled?: boolean } = $props();
</script>

<label class="flex min-h-11 cursor-pointer items-center gap-3 text-sm" for={id}>
	<input
		{id}
		type="checkbox"
		bind:checked
		{disabled}
		class="size-5 appearance-none rounded-md border-2 border-[rgb(var(--color-border))] bg-[rgb(var(--color-foreground))] checked:border-[rgb(var(--color-primary))] checked:bg-[rgb(var(--color-primary))] focus:outline-none focus:ring-2 focus:ring-[rgb(var(--color-primary))]/25 disabled:cursor-not-allowed disabled:opacity-40"
	/>
	<span style="color: rgb(var(--color-text));">{label}</span>
</label>
