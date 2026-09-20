<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		children,
		type = 'button',
		disabled = false,
		variant = 'primary',
		size = 'default',
		rounded = false,
		fullWidth = false,
		class: className = '',
		style: styleValue = '',
		onclick
	}: {
		children: Snippet;
		type?: 'button' | 'submit' | 'reset';
		disabled?: boolean;
		variant?: 'primary' | 'primary-outline' | 'outline' | 'danger' | 'ghost' | 'soft' | 'success' | 'danger-soft';
		size?: 'default' | 'small';
		rounded?: boolean;
		fullWidth?: boolean;
		class?: string;
		style?: string;
		onclick?: (event: MouseEvent) => void;
	} = $props();

	const variants = {
		primary: 'button-primary',
		'primary-outline': 'button-primary-outline',
		outline: 'button-outline',
		danger: 'button-outline',
		ghost: 'button-ghost',
		soft: 'button-soft',
		success: 'button-success',
		'danger-soft': 'button-danger-soft'
	};
</script>

<button
	{type}
	{disabled}
	{onclick}
	style={`${rounded ? 'border-radius: 9999px !important;' : ''}${styleValue}`}
	class={`inline-flex min-h-11 items-center justify-center border px-4 py-2 text-sm font-medium transition disabled:cursor-not-allowed disabled:opacity-40 ${variants[variant]} ${size === 'small' ? 'button-small' : ''} ${fullWidth ? 'w-full' : ''} ${className}`}
>
	{@render children()}
</button>
