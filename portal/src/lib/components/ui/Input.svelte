<script lang="ts">
	import type { HTMLInputAttributes } from 'svelte/elements';

	let {
		id,
		label,
		type = 'text',
		value = $bindable(''),
		placeholder = '',
		required = false,
		autocomplete,
		disabled = false,
		readonly = false,
		minlength,
		error = ''
	}: {
		id: string;
		label: string;
		type?: 'text' | 'email' | 'password' | 'search';
		value?: string;
		placeholder?: string;
		required?: boolean;
		autocomplete?: HTMLInputAttributes['autocomplete'];
		disabled?: boolean;
		readonly?: boolean;
		minlength?: number;
		error?: string;
	} = $props();
</script>

<label class="grid gap-2 text-sm" for={id}>
	<span class="font-medium">{label}</span>
	<input
		{id}
		{type}
		bind:value
		{placeholder}
		{required}
		{autocomplete}
		{disabled}
		{readonly}
		{minlength}
		aria-invalid={error ? 'true' : undefined}
		aria-describedby={error ? `${id}-error` : undefined}
		class="min-h-11 w-full border border-black bg-white px-3 py-2 text-black outline-none placeholder:text-black/50 focus:ring-2 focus:ring-black focus:ring-offset-2 disabled:opacity-40"
	/>
	{#if error}
		<span id={`${id}-error`} class="text-xs" role="alert">{error}</span>
	{/if}
</label>
