<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { Snippet } from 'svelte';
	import { callApi, appPath } from '$lib/api';
	import type { SessionUser } from '$lib/types';
	import Button from '$lib/components/ui/Button.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';

	let {
		user,
		children
	}: {
		user: SessionUser;
		children: Snippet;
	} = $props();

	let loggingOut = $state(false);

	async function logout() {
		loggingOut = true;
		try {
			await callApi('/api/auth/logout', { method: 'POST' });
		} finally {
			await goto(appPath('/auth/login'), { replaceState: true });
		}
	}
</script>

<div class="min-h-screen bg-white text-black">
	<header class="border-b border-black">
		<div class="mx-auto flex max-w-7xl flex-wrap items-center justify-between gap-4 px-4 py-4 sm:px-6 lg:px-8">
			<a class="text-lg font-bold tracking-tight" href={appPath('/')}>Let It Call</a>
			<div class="flex items-center gap-4">
				<span class="hidden text-sm sm:inline">{user.email}</span>
				<ThemeToggle />
				<Button variant="secondary" disabled={loggingOut} onclick={logout}>
					{loggingOut ? 'Signing out…' : 'Sign out'}
				</Button>
			</div>
		</div>
	</header>

	<div class="mx-auto grid max-w-7xl md:grid-cols-[14rem_1fr]">
		<nav class="border-b border-black p-4 md:min-h-[calc(100vh-77px)] md:border-r md:border-b-0 sm:p-6">
			<ul class="flex gap-2 md:grid">
				<li>
					<a
						class={`block border border-black px-4 py-3 text-sm font-medium ${page.url.pathname.startsWith(appPath('/scheduling')) ? 'bg-black text-white' : 'bg-white text-black'}`}
						href={appPath('/scheduling')}>Scheduling</a
					>
				</li>
				<li>
					<a
						class={`block border border-black px-4 py-3 text-sm font-medium ${page.url.pathname === appPath('/') ? 'bg-black text-white' : 'bg-white text-black'}`}
						href={appPath('/')}>Dashboard</a
					>
				</li>
				<li>
					<a
						class={`block border border-black px-4 py-3 text-sm font-medium ${page.url.pathname.startsWith(appPath('/users')) ? 'bg-black text-white' : 'bg-white text-black'}`}
						href={appPath('/users')}>Users</a
					>
				</li>
			</ul>
		</nav>

		<main class="min-w-0 p-4 sm:p-6 lg:p-10">
			{@render children()}
		</main>
	</div>
</div>
