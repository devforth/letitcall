<script lang="ts">
	import { page } from '$app/state';
	import Icon from '@iconify/svelte';
	import robotIcon from '@iconify-icons/tabler/robot';
	import { onMount } from 'svelte';
	import { callApi, signInAndReturn } from '$lib/api';
	import Button from '$lib/components/ui/Button.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import { branding } from '$lib/stores/branding.svelte';
	import type { MCPAuthorization } from '$lib/types';

	const request = page.url.searchParams.get('request') ?? '';
	let authorization = $state<MCPAuthorization | null>(null);
	let error = $state('');
	let resolving = $state(false);
	// A page asking for a double click could swap itself for this one between the clicks, landing the second on Allow.
	let armed = $state(false);

	onMount(async () => {
		try {
			authorization = await callApi<MCPAuthorization>(
				`/api/mcp/oauth/authorization?request=${encodeURIComponent(request)}`,
				undefined,
				false
			);
			setTimeout(() => (armed = true), 500);
		} catch (cause) {
			if (cause instanceof Error && cause.message === 'authentication required') {
				await signInAndReturn();
				return;
			}
			error = cause instanceof Error ? cause.message : 'Unable to load the connection request';
		}
	});

	async function resolve(approved: boolean) {
		resolving = true;
		try {
			const response = await callApi<{ redirectUrl: string }>('/api/mcp/oauth/authorization', {
				method: 'POST',
				body: JSON.stringify({ request, approved })
			});
			window.location.assign(response.redirectUrl);
		} catch {
			// callApi reports the error globally.
			resolving = false;
		}
	}
</script>

<PageTitle title="Connect AI" />

<main class="grid min-h-screen place-items-center px-4 py-12">
	<section
		class="w-full max-w-md rounded-2xl border-2 border-border p-8 text-center sm:p-10"
		style="background: rgb(var(--color-background)); box-shadow: var(--shadow);"
		aria-live="polite"
	>
		<Icon icon={robotIcon} width="32" height="32" class="mx-auto" />
		{#if authorization}
			<h1 class="mt-5 text-2xl font-semibold tracking-tight">Connect {authorization.clientName}?</h1>
			<p class="mt-3 text-sm">
				<b>{authorization.clientName}</b> ({authorization.clientHost}) asks to act on your behalf in {branding.name}: it
				will see and change everything you can.
			</p>
			<p class="mt-3 text-sm" style="color: rgb(var(--color-text) / 0.65);">
				{authorization.loopbackRedirect
					? `After you answer, your browser returns to an app running on this computer. Allow only if you have just started connecting ${authorization.clientName} yourself.`
					: `After you answer, your browser returns to ${authorization.redirectHost}.`}
			</p>
			<div class="mt-6 grid grid-cols-2 gap-3">
				<Button variant="outline" fullWidth disabled={resolving} onclick={() => resolve(false)}>Deny</Button>
				<Button fullWidth disabled={resolving || !armed} onclick={() => resolve(true)}>Allow</Button>
			</div>
		{:else if error}
			<h1 class="mt-5 text-2xl font-semibold tracking-tight">Cannot connect the AI agent</h1>
			<p class="mt-3 text-sm" role="alert">{error}</p>
		{:else}
			<h1 class="mt-5 text-2xl font-semibold tracking-tight">Connect AI</h1>
			<p class="mt-3 text-sm">Loading the connection request…</p>
		{/if}
	</section>
</main>
