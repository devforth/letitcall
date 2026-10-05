<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { Booking, PublicEventType } from '$lib/types';
	import EventTypeAside from '$lib/components/EventTypeAside.svelte';

	let {
		eventType,
		booking,
		bookingDetails = 'all',
		children
	}: {
		eventType: PublicEventType;
		booking?: Booking;
		bookingDetails?: 'fixed' | 'all';
		children: Snippet;
	} = $props();

	const blockStyle =
		'background: rgb(var(--color-background)); border: 1px solid var(--color-border); box-shadow: 0 0 3vw color-mix(in srgb, rgb(var(--color-background)) var(--booking-shadow-background-weight, 80%), black);';
</script>

<main class="min-h-screen sm:p-8 lg:p-10">
	<div
		class="custom__general-container mx-auto grid min-h-screen max-w-7xl overflow-hidden sm:min-h-[calc(100vh-5rem)] sm:rounded-2xl lg:h-[calc(100vh-5rem)] lg:min-h-0 lg:grid-cols-[21rem_1fr]"
		style={blockStyle}
	>
		<EventTypeAside {eventType} {booking} {bookingDetails} />
		{@render children()}
	</div>
</main>

<style>
	:global(html.dark) .custom__general-container {
		--booking-shadow-background-weight: 40%;
	}
</style>
