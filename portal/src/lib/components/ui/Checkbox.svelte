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
		class="size-5 appearance-none border border-black bg-white checked:bg-black focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 disabled:opacity-40"
	/>
	<span>{label}</span>
</label>
