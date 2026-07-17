<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		children,
		type = 'button',
		disabled = false,
		variant = 'primary',
		size = 'default',
		fullWidth = false,
		class: className = '',
		onclick
	}: {
		children: Snippet;
		type?: 'button' | 'submit' | 'reset';
		disabled?: boolean;
		variant?: 'primary' | 'secondary' | 'danger' | 'ghost' | 'soft';
		size?: 'default' | 'small';
		fullWidth?: boolean;
		class?: string;
		onclick?: (event: MouseEvent) => void;
	} = $props();

	const variants = {
		primary: 'button-primary',
		secondary: 'button-secondary',
		danger: 'button-secondary',
		ghost: 'button-ghost',
		soft: 'button-soft'
	};
</script>

<button
	{type}
	{disabled}
	{onclick}
	class={`inline-flex min-h-11 items-center justify-center border px-4 py-2 text-sm font-medium transition disabled:cursor-not-allowed disabled:opacity-40 ${variants[variant]} ${size === 'small' ? 'button-small' : ''} ${fullWidth ? 'w-full' : ''} ${className}`}
>
	{@render children()}
</button>
