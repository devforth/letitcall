<script lang="ts">
	import './layout.css';
	import NotificationStack from '$lib/components/NotificationStack.svelte';
	import { theme } from '$lib/stores/theme';
	import { branding } from '$lib/stores/branding.svelte';
	import { getPublicConfig } from '$lib/api';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import { onMount } from 'svelte';

	let { children } = $props();

	onMount(() => {
		// Apply theme to DOM on mount
		const unsubscribe = theme.subscribe((currentTheme) => {
			if (currentTheme === 'dark') {
				document.documentElement.classList.add('dark');
			} else {
				document.documentElement.classList.remove('dark');
			}
		});
		void getPublicConfig(false).then((config) => {
			branding.name = config.brandName;
			branding.logoPath = config.logoPath;
		}).catch(() => {});

		return unsubscribe;
	});
</script>

<svelte:head>
	<meta
		name="description"
		content="A focused scheduling application for teams and their calendars."
	/>
</svelte:head>

<PageTitle />

{@render children()}
<NotificationStack />
