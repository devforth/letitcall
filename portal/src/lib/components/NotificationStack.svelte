<script lang="ts">
	import { dismissNotification, notifications } from '$lib/notifications';
</script>

<div
	class="pointer-events-none fixed top-4 right-4 z-50 grid w-[min(24rem,calc(100vw-2rem))] gap-2"
	aria-live="assertive"
>
	{#each $notifications as notification (notification.id)}
		<div class="pointer-events-auto flex items-start gap-3 border border-black bg-white p-4 shadow-lg">
			<p class="min-w-0 flex-1 text-sm">{notification.message}</p>
			<button
				type="button"
				class="grid size-7 shrink-0 place-items-center border border-black text-sm"
				aria-label="Dismiss notification"
				onclick={() => dismissNotification(notification.id)}>×</button
			>
		</div>
	{/each}
</div>
