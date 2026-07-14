<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import googleIcon from '@iconify-icons/logos/google-icon';
	import Icon from '@iconify/svelte';
	import { onMount } from 'svelte';
	import { appPath, callApi } from '$lib/api';
	import Button from '$lib/components/ui/Button.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';

	let error = $state('');

	onMount(async () => {
		try {
			await callApi('/api/auth/google/callback', {
				method: 'POST',
				body: JSON.stringify({
					state: page.url.searchParams.get('state') ?? '',
					code: page.url.searchParams.get('code') ?? '',
					error: page.url.searchParams.get('error') ?? ''
				})
			});
			await goto(appPath('/'), { replaceState: true });
		} catch (cause) {
			error = cause instanceof Error ? cause.message : 'Unable to complete Google sign-in';
		}
	});
</script>

<PageTitle title="Google sign-in" />

<main class="grid min-h-screen place-items-center px-4 py-12">
	<section
		class="w-full max-w-md rounded-2xl border-2 border-border p-8 text-center sm:p-10"
		style="background: rgb(var(--color-foreground)); box-shadow: var(--shadow);"
		aria-live="polite"
	>
		<Icon icon={googleIcon} width="32" height="32" class="mx-auto" />
		{#if error}
			<h1 class="mt-5 text-2xl font-semibold tracking-tight">Google sign-in failed</h1>
			<p class="mt-3 text-sm" role="alert">{error}</p>
			<div class="mt-6">
				<Button variant="secondary" fullWidth class="lg-pd" onclick={() => goto(appPath('/auth/login'))}>
					Back to sign in
				</Button>
			</div>
		{:else}
			<h1 class="mt-5 text-2xl font-semibold tracking-tight">Signing in with Google</h1>
			<p class="mt-3 text-sm">Finishing your sign-in…</p>
		{/if}
	</section>
</main>
