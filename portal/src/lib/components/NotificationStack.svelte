<script lang="ts">
	import xIcon from '@iconify-icons/tabler/x';
	import alertTriangleIcon from '@iconify-icons/tabler/alert-triangle';
	import checkIcon from '@iconify-icons/tabler/check';
	import Icon from '@iconify/svelte';
	import { dismissNotification, notificationDurationMs, notifications } from '$lib/notifications';

	const variantIcons = {
		success: checkIcon,
		error: alertTriangleIcon
	};
</script>

<div
	class="pointer-events-none fixed top-4 right-4 z-50 grid w-[min(24rem,calc(100vw-2rem))] gap-2"
	aria-live="assertive"
>
	{#each $notifications as notification (notification.id)}
		<div
			class="notification pointer-events-auto"
			class:notification-error={notification.variant === 'error'}
			class:notification-success={notification.variant === 'success'}
		>
			<span class="notification-badge">
				<Icon icon={variantIcons[notification.variant]} width="20" height="20" aria-hidden="true" />
			</span>
			<p class="notification-title">{notification.message}</p>
			<button
				type="button"
				class="notification-close"
				aria-label="Dismiss notification"
				onclick={() => dismissNotification(notification.id)}
			>
				<Icon icon={xIcon} width="16" height="16" aria-hidden="true" />
			</button>
			<div class="notification-timer" style={`--timer-duration: ${notificationDurationMs}ms;`} aria-hidden="true"></div>
		</div>
	{/each}
</div>

<style>
	.notification {
		position: relative;
		display: flex;
		align-items: center;
		gap: 0.75rem;
		overflow: hidden;
		border-radius: 10px;
		padding: 0.875rem 0.75rem 1rem 0.875rem;
		background: rgb(var(--color-background));
		box-shadow:
			0 0 0 1px var(--color-border),
			0 8px 24px rgb(0 0 0 / 0.1);
		color: rgb(var(--color-text));
	}

	.notification-error {
		--notification-color: rgb(var(--error));
	}

	.notification-success {
		--notification-color: rgb(var(--success));
	}

	.notification-badge {
		display: grid;
		width: 2.25rem;
		height: 2.25rem;
		flex: none;
		place-items: center;
		border-radius: 999px;
		background: color-mix(in srgb, var(--notification-color) 14%, transparent);
		color: var(--notification-color);
	}

	.notification-title {
		flex: 1;
		min-width: 0;
		margin: 0;
		font-size: 0.9375rem;
		font-weight: 600;
		line-height: 1.35;
	}

	.notification-close {
		display: grid;
		width: 1.75rem;
		height: 1.75rem;
		flex: none;
		place-items: center;
		border: 0;
		border-radius: 6px;
		background: transparent;
		color: rgb(var(--color-text) / 0.65);
		cursor: pointer;
	}

	.notification-close:hover {
		background: rgb(var(--color-text) / 0.1);
		color: rgb(var(--color-text));
	}

	.notification-timer {
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0;
		height: 2px;
		background: var(--notification-color);
		transform-origin: left;
		animation: expire var(--timer-duration) linear forwards;
	}

	@keyframes expire {
		to {
			transform: scaleX(0);
		}
	}
</style>
