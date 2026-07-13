<script lang="ts">
	import xIcon from '@iconify-icons/tabler/x';
	import checkIcon from '@iconify-icons/tabler/check';
	import exclamationIcon from '@iconify-icons/tabler/exclamation-mark';
	import Icon from '@iconify/svelte';
	import { dismissNotification, notificationDurationMs, notifications } from '$lib/notifications';

	const variantIcons = {
		success: checkIcon,
		error: xIcon,
		info: exclamationIcon
	};
</script>

<div
	class="pointer-events-none fixed top-4 right-4 z-50 grid w-[min(24rem,calc(100vw-2rem))] gap-2"
	aria-live="assertive"
>
	{#each $notifications as notification (notification.id)}
		<div
			class="notification pointer-events-auto overflow-hidden rounded-lg"
		class:notification-info={notification.variant === 'info'}
		class:notification-error={notification.variant === 'error'}
		class:notification-success={notification.variant === 'success'}
		>
			<button
				type="button"
				class="notification-close absolute top-2 right-2 grid size-6 place-items-center rounded-md transition"
				aria-label="Dismiss notification"
				onclick={() => dismissNotification(notification.id)}
			>
				<Icon icon={xIcon} width="14" height="14" aria-hidden="true" />
			</button>
			<div class="flex items-stretch">
				<span class="notification-badge flex shrink-0 items-center justify-center px-2.5">
					<Icon icon={variantIcons[notification.variant]} width="30" height="30" aria-hidden="true" />
				</span>
				<div class="flex min-w-0 flex-1 flex-col gap-3 p-4">
					<p class="text-sm">{notification.message}</p>
					{#if notification.progress !== undefined}
						<div class="flex items-center justify-between gap-3 text-xs">
							<div class="flex-1">
								<div class="h-1.5 bg-border rounded overflow-hidden">
									<div
										class="h-full bg-current transition-all"
										style={`width: ${notification.progress}%`}
										aria-valuenow={notification.progress}
										aria-valuemin={0}
										aria-valuemax={100}
										role="progressbar"
									></div>
								</div>
							</div>
							<span class="whitespace-nowrap font-medium">{notification.progress}%</span>
						</div>
						{#if notification.subtitle}
							<p class="text-xs opacity-75">{notification.subtitle}</p>
						{/if}
					{/if}
					<div class="notification-timer" style={`--timer-duration: ${notificationDurationMs}ms;`} aria-hidden="true"></div>
				</div>
			</div>
		</div>
	{/each}
</div>

<style>
	.notification {
		position: relative;
		--notification-color: rgb(var(--color-primary));
		background: rgb(var(--color-foreground));
		box-shadow: var(--shadow);
		border: 1px solid rgb(var(--color-border));
	}

	.notification-close {
		color: rgb(var(--color-text) / 0.5);
	}

	.notification-close:hover {
		color: rgb(var(--color-text));
		background: rgb(var(--color-background));
	}

	.notification-badge {
		color: var(--notification-color);
		background: color-mix(in srgb, rgb(var(--color-foreground)), black 14%);
	}

	.notification-info {
		--notification-color: rgb(var(--warning));
	}

	.notification-error {
		--notification-color: rgb(var(--error));
	}

	.notification-success {
		--notification-color: rgb(var(--success));
	}

	.notification-timer {
		position: relative;
		margin-top: 0.25rem;
		height: 4px;
		border-radius: 9999px;
		background: rgb(var(--color-background));
		overflow: hidden;
	}

	.notification-timer::after {
		content: '';
		position: absolute;
		inset: 0;
		border-radius: inherit;
		background: rgb(var(--color-border));
		transform-origin: right;
		animation-name: expire;
		animation-duration: var(--timer-duration);
		animation-timing-function: linear;
		animation-fill-mode: forwards;
	}

	@keyframes expire {
		to {
			transform: scaleX(0);
		}
	}
</style>
