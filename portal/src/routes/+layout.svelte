<script lang="ts">
	import './layout.css';
	import NotificationStack from '$lib/components/NotificationStack.svelte';

	let { children } = $props();
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
