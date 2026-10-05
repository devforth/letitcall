<script lang="ts">
	import './layout.css';
	import defaultFavicon from '$lib/assets/favicon.svg';
	import NotificationStack from '$lib/components/NotificationStack.svelte';
	import { theme } from '$lib/stores/theme';
	import { branding, loadBranding } from '$lib/stores/branding.svelte';
	import { pageTheme } from '$lib/public-theme';
	import { page } from '$app/state';
	import { logoURL } from '$lib/api';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import { onMount } from 'svelte';

	let { children } = $props();
	let favicon = $derived(branding.logoPath ? logoURL(branding.logoPath) : defaultFavicon);

	$effect(() => {
		document.documentElement.classList.toggle(
			'dark',
			pageTheme(page.url.pathname, $theme, branding.publicTheme) === 'dark'
		);
	});

	onMount(() => {
		void loadBranding(false).catch(() => {});
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<meta
		name="description"
		content="A focused scheduling application for teams and their calendars."
	/>
</svelte:head>

<PageTitle />

{@render children()}
<NotificationStack />
