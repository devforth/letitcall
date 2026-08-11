<script lang="ts">
	import { avatarURL } from '$lib/api';

	let {
		name = '',
		email,
		avatarPath = '',
		size = 40,
		rounded = 'full',
		ring = false,
		onBrand = false,
		class: klass = ''
	}: {
		/** Full name; used for initials. Falls back to email when empty. */
		name?: string | null;
		email: string;
		avatarPath?: string | null;
		/** Rendered width/height in pixels. */
		size?: number;
		rounded?: 'full' | 'xl' | 'lg' | 'md' | 'sm' | 'none';
		/** 1px border ring in the theme border color. */
		ring?: boolean;
		/** Set when the avatar sits on a --color-primary surface; see initialsBg. */
		onBrand?: boolean;
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

	// Brand mesh — the one fallback treatment, for every size and call site. Two soft
	// blobs of --color-primary, the stronger at 22%/20% and a quieter one at 80%/78%,
	// over a flat 9% wash, with the initials in full-strength primary. Two light sources
	// on a diagonal give the disc more shape than a single spotlight can without pushing
	// any one stop high enough to fight the letters.
	//
	// Both blobs fade to `/ 0` — the same brand colour at zero alpha, not `transparent`,
	// so no engine interpolates the midpoint through a grey. Every stop is alpha, so the
	// disc picks up whatever it sits on and follows the tenant's branding hue without a
	// per-hue table.
	//
	// On a --color-primary surface those alpha stops have nothing to sit on: primary
	// tinted with primary, holding primary letters, disappears. `onBrand` keeps the three
	// layers byte-for-byte identical and slides an opaque --color-contrast-text base
	// underneath, so the disc reads the same there as anywhere else. Same treatment,
	// not a second one.
	const mesh = [
		'radial-gradient(circle at 22% 20%, rgb(var(--color-primary) / 0.4), rgb(var(--color-primary) / 0) 58%)',
		'radial-gradient(circle at 80% 78%, rgb(var(--color-primary) / 0.3), rgb(var(--color-primary) / 0) 60%)',
		'linear-gradient(rgb(var(--color-primary) / 0.09), rgb(var(--color-primary) / 0.09))'
	].join(', ');
	const initialsBg = $derived(
		`background-color: ${onBrand ? 'rgb(var(--color-contrast-text))' : 'transparent'}; background-image: ${mesh}; color: rgb(var(--color-primary));`
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
