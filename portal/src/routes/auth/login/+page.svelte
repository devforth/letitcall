<script lang="ts">
	import { goto } from '$app/navigation';
	import googleIcon from '@iconify-icons/logos/google-icon';
	import Icon from '@iconify/svelte';
	import { onMount } from 'svelte';
	import { callApi, appPath, getPublicConfig, getSession } from '$lib/api';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';

	let email = $state('');
	let password = $state('');
	let googleEnabled = $state(false);
	let submitting = $state(false);
	let error = $state('');

	onMount(async () => {
		try {
			await getSession(false);
			await goto(appPath('/'), { replaceState: true });
			return;
		} catch {
			// Anonymous visitors should remain on the login page.
		}

		try {
			googleEnabled = (await getPublicConfig()).googleLoginEnabled;
		} catch (cause) {
			error = cause instanceof Error ? cause.message : 'Unable to load login settings';
		}
	});

	async function login(event: SubmitEvent) {
		event.preventDefault();
		submitting = true;
		error = '';

		try {
			await callApi('/api/auth/login', {
				method: 'POST',
				body: JSON.stringify({ email, password })
			});
			await goto(appPath('/'), { replaceState: true });
		} catch (cause) {
			error = cause instanceof Error ? cause.message : 'Unable to sign in';
		} finally {
			submitting = false;
		}
	}

	function googleLogin() {
		window.location.assign(appPath('/api/auth/google/start'));
	}
</script>

<svelte:head><title>Sign in · Let It Call</title></svelte:head>

<main class="grid min-h-screen place-items-center px-4 py-12">
	<section class="w-full max-w-md border border-black p-6 sm:p-8" aria-labelledby="login-title">
		<div class="mb-8">
			<p class="mb-2 text-sm font-medium">Let It Call</p>
			<h1 id="login-title" class="text-2xl font-semibold tracking-tight">Sign in</h1>
			<p class="mt-2 text-sm">Use the credentials created by an existing user.</p>
		</div>

		{#if error}
			<p class="mb-5 border border-black p-3 text-sm" role="alert">{error}</p>
		{/if}

		<form class="grid gap-5" onsubmit={login}>
			<Input id="email" label="Email or username" bind:value={email} required autocomplete="username" />
			<Input
				id="password"
				label="Password"
				type="password"
				bind:value={password}
				required
				autocomplete="current-password"
			/>
			<Button type="submit" fullWidth disabled={submitting}>
				{submitting ? 'Signing in…' : 'Sign in'}
			</Button>
		</form>

		{#if googleEnabled}
			<div class="my-6 flex items-center gap-3" aria-hidden="true">
				<div class="h-px flex-1 bg-black"></div>
				<span class="text-xs uppercase">or</span>
				<div class="h-px flex-1 bg-black"></div>
			</div>
			<Button variant="secondary" fullWidth onclick={googleLogin}>
				<span class="flex items-center gap-2">
					<Icon icon={googleIcon} width="20" height="20" />
					Continue with Google and allow calendar access
				</span>
			</Button>
		{/if}
	</section>
</main>
