<script lang="ts">
	import { avatarURL } from '$lib/api';
	import { avatarColorsFromInitials } from '$lib/avatar-colors';
	import { branding } from '$lib/stores/branding.svelte';
	import { theme } from '$lib/stores/theme';

	let {
		name = '',
		email,
		avatarPath = '',
		size = 40,
		rounded = 'full',
		class: klass = ''
	}: {
		/** Full name; used for initials. Falls back to email when empty. */
		name?: string | null;
		email: string;
		avatarPath?: string | null;
		/** Rendered width/height in pixels. */
		size?: number;
		rounded?: 'full' | 'xl' | 'lg' | 'md' | 'sm' | 'none';
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
	const fontSize = $derived(Math.round(size * 0.36));

	const initials = $derived.by(() => {
		const parts = name?.trim().split(/\s+/).filter(Boolean) ?? [];
		if (parts.length >= 2) return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
		if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
		return email.slice(0, 2).toUpperCase();
	});
	const initialsColors = $derived(
		avatarColorsFromInitials(initials, branding.theme[$theme].text, branding.theme[$theme].background)
	);
	const initialsStyle = $derived(
		`background-color: ${initialsColors.background}; color: ${initialsColors.textColor};`
	);
	let avatarFailed = $state(false);
</script>

{#if avatarPath && !avatarFailed}
	<img
		src={avatarURL(avatarPath)}
		alt=""
		onerror={() => (avatarFailed = true)}
		class={klass}
		style="width: {size}px; height: {size}px; border-radius: {radius}; object-fit: cover;"
	/>
{:else}
	<span
		class={klass}
		aria-label="Avatar"
		style="width: {size}px; height: {size}px; border-radius: {radius}; display: inline-flex; align-items: center; justify-content: center; font-weight: 700; line-height: 1; font-size: {fontSize}px; {initialsStyle}"
	>{initials}</span>
{/if}
