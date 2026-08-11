<script lang="ts">
	import Icon from '@iconify/svelte';
	import clockIcon from '@iconify-icons/tabler/clock';
	import notesIcon from '@iconify-icons/tabler/align-left';
	import usersIcon from '@iconify-icons/tabler/users';
	import type { Booking, PublicEventType } from '$lib/types';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import { branding } from '$lib/stores/branding.svelte';

	let {
		eventType,
		booking,
		bookingDetails = 'all'
	}: {
		eventType: PublicEventType;
		booking?: Booking;
		bookingDetails?: 'fixed' | 'all';
	} = $props();

	const productName = 'Let It Call';
	let highlightedHost = $state<string | null>(null);

	const hosts = $derived([
		...eventType.requiredHosts.map((host) => ({ ...host, avatarSize: 76 })),
		...eventType.optionalHosts.map((host) => ({ ...host, avatarSize: 52 }))
	]);
	const hostHighlightEnabled = $derived(hosts.length > 1);
	const asideStyle =
		'background: rgb(var(--color-primary)); color: rgb(var(--color-contrast-text)); box-shadow: 0 0 0 1px rgb(var(--color-border)), var(--shadow-small);';

	function bookingDate(): string {
		return new Intl.DateTimeFormat(undefined, {
			dateStyle: 'full',
			timeZone: booking?.attendeeTimezone
		}).format(new Date(booking!.time));
	}

	function bookingTimeRange(): string {
		const formatter = new Intl.DateTimeFormat(undefined, {
			hour: 'numeric',
			minute: '2-digit',
			timeZone: booking?.attendeeTimezone
		});
		return `${formatter.format(new Date(booking!.time))} – ${formatter.format(new Date(booking!.endTime))}`;
	}

	function highlightHost(email: string | null) {
		highlightedHost = hostHighlightEnabled ? email : null;
	}
</script>

{#snippet hostNames(hostList: PublicEventType['requiredHosts'])}
	{#each hostList as host, index (host.email)}
		{#if index > 0}, {/if}<span
			class="host-name"
			class:highlight-enabled={hostHighlightEnabled}
			class:highlighted={hostHighlightEnabled && highlightedHost === host.email}
			role="presentation"
			onmouseenter={() => highlightHost(host.email)}
			onmouseleave={() => highlightHost(null)}
		>{host.fullName || host.email}</span>
	{/each}
{/snippet}

<aside class="relative flex flex-col p-6 pb-10 sm:pb-6 sm:rounded-b-2xl lg:rounded-bl-none lg:rounded-tr-2xl lg:rounded-br-2xl lg:p-8" style={asideStyle}>
	<h1 class="text-3xl font-semibold tracking-tight">{eventType.name}</h1>
	<div class="mt-8 flex items-end -space-x-4">
		{#each hosts as host (host.email)}
			<span
				class="host-avatar"
				class:highlighted={hostHighlightEnabled && highlightedHost === host.email}
				role="presentation"
				onmouseenter={() => highlightHost(host.email)}
				onmouseleave={() => highlightHost(null)}
			>
				<Avatar
					name={host.fullName}
					email={host.email}
					avatarPath={host.avatarPath}
					size={host.avatarSize}
					rounded="full"
					onBrand
					class="shadow-[0_0_0_4px_rgb(var(--color-primary))]"
				/>
			</span>
		{/each}
	</div>
	<p class="mt-6 text-sm font-medium">{@render hostNames(eventType.requiredHosts)}</p>
	{#if eventType.optionalHosts.length > 0}
		<p class="mt-1 text-xs">Optional: {@render hostNames(eventType.optionalHosts)}</p>
	{/if}
	{#if !booking}
		<p class="mt-7 flex items-center gap-2 text-sm font-medium">
			<Icon icon={clockIcon} width="22" height="22" />
			{eventType.durationMinutes} min
		</p>
	{/if}
	{#if booking}
		<div class="booking-details">
			<section class="booking-detail">
				<Icon icon={clockIcon} width="22" height="22" class="mt-0.5 shrink-0" />
				<div>
					<p>{bookingDate()}</p>
					<p>{bookingTimeRange()}</p>
					<p class="booking-detail-muted">{booking.attendeeTimezone}</p>
				</div>
			</section>
			<section class="booking-detail">
				<Icon icon={usersIcon} width="22" height="22" class="mt-0.5 shrink-0" />
				<div>
					<p>{booking.attendeeName} · {booking.attendeeEmail}</p>
			{#if bookingDetails === 'all' && booking.guestEmails.length > 0}
						<ul class="booking-detail-list">
							{#each booking.guestEmails as email (email)}
								<li>{email}</li>
							{/each}
						</ul>
					{/if}
				</div>
			</section>
			{#if bookingDetails === 'all' && booking.notes}
				<section class="booking-detail">
					<Icon icon={notesIcon} width="22" height="22" class="mt-0.5 shrink-0" />
					<div>
						<p class="booking-detail-notes">{booking.notes}</p>
					</div>
				</section>
			{/if}
		</div>
	{/if}
	<!-- Mobile puts the brand row above the event name; from sm up it drops back to the
	     foot of the aside where the "Powered by" line sits beneath it. -->
	<div class="order-first flex items-center justify-between gap-4 pb-1 sm:order-none sm:mt-auto sm:pb-0 sm:pt-12">
		<p class="text-xl font-semibold">{branding.name}</p>
		<div class="theme-toggle-contrast shrink-0">
			<ThemeToggle />
		</div>
	</div>
	{#if branding.name !== productName}
		<p
			class="absolute bottom-3 right-5 whitespace-nowrap text-xs font-normal sm:left-1/2 sm:right-auto sm:-translate-x-1/2"
			style="color: rgb(var(--color-contrast-text) / 0.72);"
		>
			Powered by <span class="font-medium" style="color: rgb(var(--color-contrast-text));">{productName}</span>
		</p>
	{/if}
</aside>

<style>
	.host-avatar {
		display: inline-flex;
		border-radius: 9999px;
		transition: transform 0.15s ease;
	}

	.host-avatar.highlighted {
		z-index: 1;
		transform: scale(1.08);
	}

	.host-name {
		text-decoration-color: transparent;
		text-underline-offset: 0.2rem;
		transition: text-decoration-color 0.15s ease;
	}

	.host-name.highlight-enabled {
		cursor: default;
	}

	.host-name.highlighted {
		font-weight: 700;
		text-decoration-line: underline;
		text-decoration-color: currentColor;
		text-decoration-thickness: 2px;
	}

	.booking-details {
		display: grid;
		gap: 1.25rem;
		margin-top: 2rem;
		border-top: 1px solid rgb(var(--color-contrast-text) / 0.3);
		padding-top: 2rem;
	}

	.booking-detail {
		display: flex;
		align-items: start;
		gap: 0.75rem;
		font-size: 0.875rem;
		line-height: 1.5rem;
	}

	.booking-detail-muted {
		color: rgb(var(--color-contrast-text) / 0.75);
	}

	.booking-detail-list {
		display: grid;
		gap: 0.125rem;
		margin: 0;
		padding: 0;
		list-style: none;
		overflow-wrap: anywhere;
	}

	.booking-detail-notes {
		white-space: pre-wrap;
	}

	.theme-toggle-contrast :global(.toggle-switch) {
		background: transparent !important;
		border-color: rgb(var(--color-contrast-text)) !important;
		color: rgb(var(--color-contrast-text)) !important;
		box-shadow: none !important;
	}

	/* Pulled up to sit level with the brand name in the mobile top row, where a
	   hairline drawn as an inset shadow replaces the 2px border. */
	@media (max-width: 640px) {
		.theme-toggle-contrast {
			margin-top: -8px;
		}

		.theme-toggle-contrast :global(.toggle-switch) {
			border: 0 !important;
			box-shadow: inset 0 0 0 1px rgb(var(--color-contrast-text)) !important;
		}
	}
</style>
