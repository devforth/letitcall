<script lang="ts">
	import { page } from '$app/state';
	import EventTypeEditor from '$lib/components/EventTypeEditor.svelte';
</script>

<svelte:head><title>Edit event type · Let It Call</title></svelte:head>

<EventTypeEditor slug={page.params.slug} />
