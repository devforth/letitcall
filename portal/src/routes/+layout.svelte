<script lang="ts">
	import './layout.css';
	import defaultFavicon from '$lib/assets/favicon.svg';
	import NotificationStack from '$lib/components/NotificationStack.svelte';
	import { theme } from '$lib/stores/theme';
	import { branding, loadBranding } from '$lib/stores/branding.svelte';
	import { logoURL } from '$lib/api';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import { onMount } from 'svelte';

	let { children } = $props();
	let favicon = $derived(branding.logoPath ? logoURL(branding.logoPath) : defaultFavicon);

	onMount(() => {
		// Apply theme to DOM on mount
		const unsubscribe = theme.subscribe((currentTheme) => {
			if (currentTheme === 'dark') {
				document.documentElement.classList.add('dark');
			} else {
				document.documentElement.classList.remove('dark');
			}
		});
		void loadBranding(false).catch(() => {});

		return unsubscribe;
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
