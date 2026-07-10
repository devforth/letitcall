<script lang="ts">
	import './layout.css';
	import { theme } from '$lib/stores/theme';
	import { onMount } from 'svelte';
	import NotificationStack from '$lib/components/NotificationStack.svelte';

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

		return unsubscribe;
	});
</script>

<svelte:head>
	<title>Let It Call</title>
	<meta
		name="description"
		content="A focused scheduling application for teams and their calendars."
	/>
</svelte:head>

{@render children()}
<NotificationStack />
