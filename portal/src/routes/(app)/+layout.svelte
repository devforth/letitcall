<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { getSession } from '$lib/api';
	import AppShell from '$lib/components/AppShell.svelte';
	import type { SessionUser } from '$lib/types';

	let { children } = $props();
	let user = $state<SessionUser | null>(null);
	let error = $state('');

	onMount(async () => {
		try {
			user = (await getSession()).user;
		} catch (cause) {
			if (cause instanceof Error && cause.message !== 'authentication required') {
				error = cause.message;
				return;
			}
			await goto('/auth/login', { replaceState: true });
		}
	});
</script>

{#if user}
	<AppShell {user}>{@render children()}</AppShell>
{:else if error}
	<main class="grid min-h-screen place-items-center p-6">
		<div class="max-w-md border border-black p-6">
			<h1 class="text-xl font-semibold">Unable to load the application</h1>
			<p class="mt-3 text-sm">{error}</p>
		</div>
	</main>
{:else}
	<main class="grid min-h-screen place-items-center p-6">
		<p class="text-sm">Loading…</p>
	</main>
{/if}
