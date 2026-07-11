<script lang="ts">
	import { page } from '$app/state';
	import EventTypeEditor from '$lib/components/EventTypeEditor.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';
</script>

<PageTitle title="Edit event type" />

<EventTypeEditor slug={page.params.slug} />
