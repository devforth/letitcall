<script lang="ts">
	import { avatarURL } from '$lib/api';

	let {
		name = '',
		email,
		avatarPath = '',
		size = 40,
		rounded = 'full',
		variant = 'gradient',
		ring = false,
		class: klass = ''
	}: {
		/** Full name; used for initials. Falls back to email when empty. */
		name?: string | null;
		email: string;
		avatarPath?: string | null;
		/** Rendered width/height in pixels. */
		size?: number;
		rounded?: 'full' | 'xl' | 'lg' | 'md' | 'sm' | 'none';
		/** Initials background: soft brand gradient, or solid primary fill. */
		variant?: 'gradient' | 'solid';
		/** 1px border ring in the theme border color. */
		ring?: boolean;
		class?: string;
	} = $props();

	const radii = {
		full: '9999px',
		xl: '0.75rem',
		lg: '0.5rem',
		md: '0.375rem',
		sm: '0.25rem',
		none: '0'
	};

	const radius = $derived(radii[rounded]);
	const ringStyle = $derived(ring ? 'box-shadow: 0 0 0 1px rgb(var(--color-border));' : '');
	const fontSize = $derived(Math.round(size * 0.36));

	const initialsBg = $derived(
		variant === 'solid'
			? 'background: rgb(var(--color-primary)); color: rgb(var(--color-contrast-text));'
			: 'background: linear-gradient(135deg, rgb(var(--color-primary) / 0.35), rgb(var(--color-primary) / 0.2) 50%, rgb(var(--color-primary) / 0.08)); color: rgb(var(--color-primary));'
	);

	const initials = $derived.by(() => {
		const parts = name?.trim().split(/\s+/).filter(Boolean) ?? [];
		if (parts.length >= 2) return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
		if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
		return email.slice(0, 2).toUpperCase();
	});
</script>

{#if avatarPath}
	<img
		src={avatarURL(avatarPath)}
		alt=""
		class={klass}
		style="width: {size}px; height: {size}px; border-radius: {radius}; object-fit: cover; {ringStyle}"
	/>
{:else}
	<span
		class={klass}
		aria-label="Avatar"
		style="width: {size}px; height: {size}px; border-radius: {radius}; display: inline-flex; align-items: center; justify-content: center; font-weight: 700; line-height: 1; font-size: {fontSize}px; {initialsBg} {ringStyle}"
	>{initials}</span>
{/if}
