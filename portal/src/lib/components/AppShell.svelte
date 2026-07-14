<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { onMount, type Snippet } from 'svelte';
	import Icon from '@iconify/svelte';
	import calendarIcon from '@iconify-icons/tabler/calendar-filled';
	import clockIcon from '@iconify-icons/tabler/clock-filled';
	import functionIcon from '@iconify-icons/tabler/function-filled';
	import moonIcon from '@iconify-icons/tabler/moon';
	import paintIcon from '@iconify-icons/tabler/paint-filled';
	import timelineIcon from '@iconify-icons/tabler/timeline-event-filled';
	import xIcon from '@iconify-icons/tabler/x';
	import { callApi, appPath, avatarURL } from '$lib/api';
	import addUserIcon from '$lib/icons/add-user';
	import type { SessionUser } from '$lib/types';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import BrandLogo from '$lib/components/BrandLogo.svelte';
	import MenuToggleIcon from '$lib/components/ui/MenuToggleIcon.svelte';
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
	let navOpen = $state(false);
	let navCollapsed = $state(false);
	let isMobile = $state(false);

	onMount(() => {
		const mq = window.matchMedia('(max-width: 767px)');
		isMobile = mq.matches;

		const onChange = (e: MediaQueryListEvent) => {
			isMobile = e.matches;
			navOpen = false;
			navCollapsed = false;
		};
		mq.addEventListener('change', onChange);
		return () => mq.removeEventListener('change', onChange);
	});

	const initials = $derived.by(() => {
		const parts = user.fullName?.trim().split(/\s+/).filter(Boolean) ?? [];
		if (parts.length >= 2) return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
		if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
		return user.email.slice(0, 2).toUpperCase();
	});

	const navItems = [
		{ label: 'Bookings', href: '/', exact: true, icon: calendarIcon },
		{ label: 'Scheduling', href: '/scheduling', exact: false, icon: clockIcon },
		{ label: 'Users', href: '/users', exact: false, icon: addUserIcon },
		{ label: 'Branding', href: '/branding', exact: false, icon: paintIcon },
		{ label: 'API Integration', href: '/api-integration', exact: false, icon: functionIcon },
		{ label: 'Audit log', href: '/audit-log', exact: false, icon: timelineIcon }
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

	function toggleSidebar() {
		if (isMobile) {
			navOpen = !navOpen;
			return;
		}
		navCollapsed = !navCollapsed;
	}

	function collapseOnMobile() {
		if (isMobile) navOpen = false;
	}
</script>

<svelte:window
	onclick={closeMenu}
	onkeydown={(e) => {
		if (e.key === 'Escape') {
			closeMenu();
			collapseOnMobile();
		}
	}}
/>

<div
	class="min-h-screen transition-[padding-left] duration-300 ease-out motion-reduce:transition-none md:pl-[var(--sidebar-w)]"
	style="--sidebar-w: {navCollapsed ? '4.5rem' : '16rem'};"
>
	<header
		class="relative z-20 border-b-2 md:h-[66px]"
		style="background: rgb(var(--color-foreground)); border-color: rgb(var(--color-border)); box-shadow: 0 0.75rem 1rem -0.75rem rgb(0 0 0 / 0.2);"
	>
		<div
			class="mx-auto flex w-full max-w-[76rem] flex-wrap items-center justify-between gap-4 px-4 py-2 sm:px-6 lg:px-8"
		>
			<div class="flex min-w-0 items-center gap-3">
				<button
					type="button"
					class="menu-toggle-button flex size-10 shrink-0 items-center justify-center rounded-lg md:hidden"
					aria-label={isMobile
						? navOpen
							? 'Hide navigation'
							: 'Show navigation'
						: navCollapsed
							? 'Expand navigation'
							: 'Collapse navigation'}
					aria-expanded={isMobile ? navOpen : !navCollapsed}
					aria-controls="app-sidebar"
					onclick={toggleSidebar}
				>
					<MenuToggleIcon open={navOpen} />
				</button>
				<a class="flex min-w-0 items-center gap-3 tracking-tight" href={appPath('/')}>
					<BrandLogo class="size-10 shrink-0 rounded-lg border-2 border-black object-cover" />
					<span class="flex min-w-0 flex-col leading-tight">
						<span class="truncate text-lg font-bold" style="color: rgb(var(--color-primary));">{branding.name}</span>
						<span class="hidden text-sm font-bold sm:block">Scheduling Admin Panel</span>
					</span>
				</a>
			</div>
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
								<Icon icon={moonIcon} width="16" height="16" class="shrink-0" style="color: rgb(var(--color-muted-foreground));" />
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

	{#if navOpen && isMobile}
		<button
			type="button"
			class="fixed inset-0 z-30 cursor-default bg-black/20"
			aria-label="Close navigation"
			onclick={collapseOnMobile}
		></button>
	{/if}

	<div
		aria-hidden="true"
		class="pointer-events-none fixed bottom-0 top-[66px] z-30 hidden w-4 transition-[left] duration-300 ease-out motion-reduce:transition-none md:block"
		style="left: var(--sidebar-w); background: linear-gradient(90deg, rgb(0 0 0 / 0.06), transparent);"
	></div>

	<div class="contents">
		<nav
			id="app-sidebar"
			class={`fixed inset-y-0 left-0 z-40 w-72 overflow-y-auto border-r-2 p-4 transition-[transform,width] duration-300 ease-out motion-reduce:transition-none ${navOpen ? 'translate-x-0' : '-translate-x-full'} md:flex md:min-h-screen md:w-[var(--sidebar-w)] md:min-w-0 md:flex-col md:p-0 md:translate-x-0 md:overflow-hidden`}
			style="background: rgb(var(--color-foreground)); border-color: rgb(var(--color-border));"
			aria-label="Primary navigation"
			inert={isMobile && !navOpen}
		>
			<div class="mb-8 flex items-center justify-between md:hidden">
				<span class="text-sm font-semibold" style="color: rgb(var(--color-text));">Navigation</span>
				<button
					type="button"
					class="flex size-10 items-center justify-center rounded-lg"
					style="border: 2px solid rgb(var(--color-border)); color: rgb(var(--color-text));"
					aria-label="Close navigation"
					onclick={collapseOnMobile}
				>
					<Icon icon={xIcon} width="20" height="20" />
				</button>
			</div>

			<div
				class="hidden h-[66px] items-center border-b-2 px-4 md:flex"
				style="border-color: rgb(var(--color-border));"
			>
				<button
					type="button"
					class="menu-toggle-button flex size-10 shrink-0 items-center justify-center rounded-lg"
					aria-label={navCollapsed ? 'Expand navigation' : 'Collapse navigation'}
					aria-expanded={!navCollapsed}
					aria-controls="app-sidebar"
					onclick={toggleSidebar}
				>
					<MenuToggleIcon open={!navCollapsed} />
				</button>
				{#if !navCollapsed}
					<div class="ml-auto">
						<ThemeToggle />
					</div>
				{/if}
			</div>

			<div class="w-full md:px-4 md:pt-4">
				<ul class="flex flex-col gap-3">
					{#each navItems as item (item.href)}
						{@const active = isActive(item)}
						<li>
							<a
								class={`nav-link flex items-center rounded-lg text-sm font-medium ${navCollapsed ? 'mx-auto size-10 justify-center p-0' : 'gap-3 px-3 py-2.5'}`}
								href={appPath(item.href)}
								aria-current={active ? 'page' : undefined}
								aria-label={navCollapsed ? item.label : undefined}
								title={navCollapsed ? item.label : undefined}
								onclick={collapseOnMobile}
							>
								<Icon icon={item.icon} width="22" height="22" class="shrink-0" />
								<span class={navCollapsed ? 'hidden' : 'whitespace-nowrap'}>{item.label}</span>
							</a>
						</li>
					{/each}
				</ul>
			</div>

		</nav>

		<main class="min-w-0 p-4 sm:p-6 lg:p-8">
			<div
				class="mx-auto w-full min-w-0 max-w-6xl rounded-xl p-3 sm:p-4 lg:p-5"
				style="background: rgb(var(--color-foreground)); border: 2px solid rgb(var(--color-border)); box-shadow: var(--shadow-small);"
			>
				{@render children()}
			</div>
		</main>
	</div>
</div>

<style>
	.menu-toggle-button {
		border: 0;
		background: transparent;
		color: rgb(var(--color-text));
		cursor: pointer;
		transition: background 0.18s;
	}

	.menu-toggle-button:hover {
		background: rgb(var(--color-muted-background));
	}

	.menu-toggle-button:focus-visible {
		outline: 2px solid rgb(var(--color-text));
		outline-offset: 2px;
	}

	.nav-link {
		color: rgb(var(--color-text));
		transition: background 0.18s, color 0.18s;
	}

	.nav-link:hover:not([aria-current='page']) {
		background: rgb(var(--color-primary) / 0.08);
		color: rgb(var(--color-primary));
	}

	.nav-link[aria-current='page'] {
		background: transparent;
		color: rgb(var(--color-primary));
		font-weight: 700;
	}

	.nav-link:focus-visible {
		outline: 2px solid rgb(var(--color-text));
		outline-offset: 2px;
	}
</style>
