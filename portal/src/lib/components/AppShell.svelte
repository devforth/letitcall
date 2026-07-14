<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { Snippet } from 'svelte';
	import { callApi, appPath, avatarURL } from '$lib/api';
	import type { SessionUser } from '$lib/types';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import BrandLogo from '$lib/components/BrandLogo.svelte';
	import { branding } from '$lib/stores/branding.svelte';

	let {
		user,
		children
	}: {
		user: SessionUser;
		children: Snippet;
	} = $props();

	let loggingOut = $state(false);
	let menuOpen = $state(false);

	const initials = $derived.by(() => {
		const parts = user.fullName?.trim().split(/\s+/).filter(Boolean) ?? [];
		if (parts.length >= 2) return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
		if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
		return user.email.slice(0, 2).toUpperCase();
	});

	const navItems = [
		{ label: 'Bookings', href: '/', exact: true },
		{ label: 'Scheduling', href: '/scheduling', exact: false },
		{ label: 'Users', href: '/users', exact: false },
		{ label: 'Branding', href: '/branding', exact: false },
		{ label: 'API Integration', href: '/api-integration', exact: false },
		{ label: 'Audit log', href: '/audit-log', exact: false }
	];

	function isActive(item: { href: string; exact: boolean }) {
		const target = appPath(item.href);
		return item.exact ? page.url.pathname === target : page.url.pathname.startsWith(target);
	}

	async function logout() {
		loggingOut = true;
		try {
			await callApi('/api/auth/logout', { method: 'POST' });
		} finally {
			await goto(appPath('/auth/login'), { replaceState: true });
		}
	}

	function toggleMenu(event: MouseEvent) {
		event.stopPropagation();
		menuOpen = !menuOpen;
	}

	function closeMenu() {
		menuOpen = false;
	}
</script>

<svelte:window
	onclick={closeMenu}
	onkeydown={(e) => {
		if (e.key === 'Escape') closeMenu();
	}}
/>

<div class="min-h-screen">
	<header
		class="border-b-2"
		style="background: rgb(var(--color-foreground)); border-color: rgb(var(--color-border)); box-shadow: var(--shadow);"
	>
		<div class="mx-auto flex max-w-7xl flex-wrap items-center justify-between gap-4 px-4 py-2 sm:px-6 lg:px-8">
			<a class="flex items-center gap-3 tracking-tight" href={appPath('/')}>
				<BrandLogo class="size-10 rounded-lg border-2 border-black object-cover" />
				<span class="flex flex-col leading-tight">
					<span class="text-lg font-bold" style="color: rgb(var(--color-primary));">{branding.name}</span>
					<span class="text-sm font-bold">Scheduling Admin Panel</span>
				</span>
			</a>
			<div class="flex items-center gap-3">
				<div class="relative">
					<button
						type="button"
						class="flex items-center gap-2 rounded-full py-1 pl-1 pr-3 transition-colors"
						style="background: rgb(var(--color-foreground)); border: 2px solid rgb(var(--color-border)); box-shadow: var(--shadow-small); outline: none;"
						aria-haspopup="menu"
						aria-expanded={menuOpen}
						aria-label="Account menu"
						onclick={toggleMenu}
						onmouseenter={(e) => (e.currentTarget.style.background = 'rgb(var(--color-primary) / 0.1)')}
						onmouseleave={(e) => (e.currentTarget.style.background = 'rgb(var(--color-foreground))')}
					>
						{#if user.avatarPath}
							<img
								src={avatarURL(user.avatarPath)}
								alt=""
								class="size-9 rounded-full object-cover"
								style="border: 2px solid rgb(var(--color-border));"
							/>
						{:else}
							<span
								class="flex size-9 items-center justify-center rounded-full text-sm font-bold leading-none"
								style="background: rgb(var(--color-primary)); color: rgb(var(--color-contrast-text)); border: 2px solid rgb(var(--color-border));"
							>
								{initials}
							</span>
						{/if}
						<span class="hidden max-w-[10rem] truncate text-sm font-bold sm:inline" style="color: rgb(var(--color-text));">
							{user.fullName || user.email}
						</span>
						<svg
							class="size-4 shrink-0 transition-transform"
							style="color: rgb(var(--color-muted-foreground)); {menuOpen ? 'transform: rotate(180deg);' : ''}"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2.2"
							stroke-linecap="round"
							stroke-linejoin="round"
						>
							<path d="M6 9l6 6 6-6" />
						</svg>
					</button>

					{#if menuOpen}
						<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
						<div
							class="absolute right-0 z-40 mt-2 w-64 overflow-hidden rounded-xl p-2"
							style="background: rgb(var(--color-foreground)); border: 2px solid rgb(var(--color-border)); box-shadow: var(--shadow);"
							role="menu"
							tabindex="-1"
							onclick={(e) => e.stopPropagation()}
						>
							<div class="flex items-center gap-3 px-2 py-2">
								{#if user.avatarPath}
									<img
										src={avatarURL(user.avatarPath)}
										alt=""
										class="size-11 rounded-full object-cover"
										style="border: 2px solid rgb(var(--color-border));"
									/>
								{:else}
									<span
										class="flex size-11 items-center justify-center rounded-full text-base font-bold leading-none"
										style="background: rgb(var(--color-primary)); color: rgb(var(--color-contrast-text));"
									>
										{initials}
									</span>
								{/if}
								<span class="flex min-w-0 flex-col leading-tight">
									<span class="truncate text-sm font-bold" style="color: rgb(var(--color-text));">
										{user.fullName || user.email}
									</span>
									{#if user.fullName}
										<span class="truncate text-xs" style="color: rgb(var(--color-muted-foreground));">
											{user.email}
										</span>
									{/if}
								</span>
							</div>

							<div class="my-1 h-px" style="background: rgb(var(--color-border));"></div>

							<a
								class="flex w-full items-center gap-3 rounded-lg px-2 py-2 text-sm font-medium transition-colors"
								style="color: rgb(var(--color-text)); background: transparent;"
								role="menuitem"
								href={appPath(`/users/${encodeURIComponent(user.email)}`)}
								onclick={closeMenu}
								onmouseenter={(e) => (e.currentTarget.style.background = 'rgb(var(--color-primary) / 0.1)')}
								onmouseleave={(e) => (e.currentTarget.style.background = 'transparent')}
							>
								<svg class="size-4 shrink-0" style="color: rgb(var(--color-muted-foreground));" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H5a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2h13a2 2 0 0 0 2-2v-6" /><path d="M18.5 2.5a2.1 2.1 0 0 1 3 3L12 15l-4 1 1-4z" /></svg>
								Edit profile
							</a>

							<div class="my-1 h-px" style="background: rgb(var(--color-border));"></div>

							<div class="flex items-center gap-3 rounded-lg px-2 py-2">
								<svg class="size-4 shrink-0" style="color: rgb(var(--color-muted-foreground));" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="9" /><path d="M3 12h18M12 3a15 15 0 0 1 0 18 15 15 0 0 1 0-18z" /></svg>
								<span class="text-sm font-medium" style="color: rgb(var(--color-text));">Timezone</span>
								<span class="ml-auto truncate text-xs" style="color: rgb(var(--color-muted-foreground));">{user.timezone}</span>
							</div>

							<div class="flex items-center gap-3 rounded-lg px-2 py-2">
								<svg class="size-4 shrink-0" style="color: rgb(var(--color-muted-foreground));" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="9" /><path d="M12 3a9 9 0 0 0 0 18z" fill="currentColor" stroke="none" /></svg>
								<span class="text-sm font-medium" style="color: rgb(var(--color-text));">Theme</span>
								<div class="ml-auto">
									<ThemeToggle />
								</div>
							</div>

							<div class="my-1 h-px" style="background: rgb(var(--color-border));"></div>

							<button
								type="button"
								class="flex w-full items-center gap-3 rounded-lg px-2 py-2 text-left text-sm font-semibold transition-colors disabled:opacity-60"
								style="color: rgb(var(--error)); background: transparent;"
								role="menuitem"
								disabled={loggingOut}
								onclick={logout}
								onmouseenter={(e) => (e.currentTarget.style.background = 'rgb(var(--error) / 0.1)')}
								onmouseleave={(e) => (e.currentTarget.style.background = 'transparent')}
							>
								<svg class="size-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9" /></svg>
								{loggingOut ? 'Signing out…' : 'Sign out'}
							</button>
						</div>
					{/if}
				</div>
			</div>
		</div>
	</header>

	<div class="mx-auto grid max-w-7xl md:grid-cols-[14rem_minmax(0,1fr)]">
		<nav class="pt-4 pb-0 pl-4 pr-4 sm:pt-6 sm:pl-6 sm:pr-6 md:min-h-[calc(100vh-77px)] md:pb-6 md:pr-0">
			<ul
				class="inline-flex gap-1 rounded-xl p-2 md:grid"
				style="background: rgb(var(--color-foreground)); border: 2px solid rgb(var(--color-border)); box-shadow: var(--shadow-small);"
			>
				{#each navItems as item (item.href)}
					{@const active = isActive(item)}
					<li>
						<a
							class="flex items-center gap-3 rounded-lg px-4 py-2 text-sm font-medium transition-colors"
							style={active
								? 'background: rgb(var(--color-primary) / 0.1); color: rgb(var(--color-primary)); font-weight: 700;'
								: 'color: rgb(var(--color-text));'}
							href={appPath(item.href)}
						>
							{#if item.label === 'Bookings'}
								<svg class="size-[18px] shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="17" rx="2" /><path d="M3 9h18M8 2v4M16 2v4" /></svg>
							{:else if item.label === 'Scheduling'}
								<svg class="size-[18px] shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="9" /><path d="M12 7v5l3 2" /></svg>
							{:else if item.label === 'Users'}
								<svg class="size-[18px] shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="9" cy="8" r="3.2" /><path d="M3.5 20a5.5 5.5 0 0 1 11 0M16 5.2a3.2 3.2 0 0 1 0 6M18 20a5.5 5.5 0 0 0-3-4.9" /></svg>
							{:else if item.label === 'API Integration'}
								<svg class="size-[18px] shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M8 3H5a2 2 0 0 0-2 2v3M16 3h3a2 2 0 0 1 2 2v3M8 21H5a2 2 0 0 1-2-2v-3M16 21h3a2 2 0 0 0 2-2v-3" /><circle cx="12" cy="12" r="2.5" /></svg>
							{:else if item.label === 'Audit log'}
								<svg class="size-[18px] shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8z" /><path d="M14 3v5h5M9 13h6M9 17h4" /></svg>
							{:else}
								<svg class="size-[18px] shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3a9 9 0 1 0 0 18c1.2 0 2-.9 2-2 0-.6-.3-1-.6-1.4-.3-.4-.5-.8-.5-1.3 0-1 .8-1.8 1.8-1.8H16a5 5 0 0 0 5-5c0-3.9-4-6.5-9-6.5z" /><circle cx="7.5" cy="10.5" r="1" /><circle cx="12" cy="7.5" r="1" /><circle cx="16.5" cy="10.5" r="1" /></svg>
							{/if}
							{item.label}
						</a>
					</li>
				{/each}
			</ul>
		</nav>

		<main class="min-w-0 p-4 sm:p-6 lg:p-6">
			<div
				class="min-w-0 rounded-xl p-3 sm:p-4 lg:p-5"
				style="background: rgb(var(--color-foreground)); border: 2px solid rgb(var(--color-border)); box-shadow: var(--shadow-small);"
			>
				{@render children()}
			</div>
		</main>
	</div>
</div>
