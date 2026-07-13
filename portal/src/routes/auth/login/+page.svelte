<script lang="ts">
	import { goto } from '$app/navigation';
	import googleIcon from '@iconify-icons/logos/google-icon';
	import Icon from '@iconify/svelte';
	import { onMount } from 'svelte';
	import { callApi, appPath, getPublicConfig, getSession } from '$lib/api';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import BrandLogo from '$lib/components/BrandLogo.svelte';
	import { branding } from '$lib/stores/branding.svelte';

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
			const config = await getPublicConfig();
			googleEnabled = config.googleLoginEnabled;
			branding.name = config.brandName;
			branding.logoPath = config.logoPath;
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

<style>
	:global(.login-bg) {
		background-color: rgb(var(--color-background));
	}

	/* Input styling */
	:global(input) {
		border-color: #e5e5e5 !important;
		border-radius: 10px !important;
		background-color: white !important;
		transition: all 0.2s !important;
	}

	:global(input:focus) {
		outline: none !important;
		border-color: rgb(var(--color-secondary)) !important;
		box-shadow: 0 0 0 3px rgb(var(--color-secondary) / 0.15) !important;
	}

	/* Label styling */
	:global(label span) {
		font-weight: 600 !important;
	}

	/* Button styling */
	:global(button) {
		border-radius: 10px !important;
		transition: all 0.3s !important;
		font-weight: 700 !important;
		cursor: pointer !important;
	}

	/* Secondary button styling */
	:global(form button[type="submit"]) {
		background: rgb(var(--color-primary)) !important;
		border: 2px solid color-mix(in srgb, rgb(var(--color-primary)), black 10%) !important;
		color: rgb(var(--color-contrast-text)) !important;
		padding: 0.875rem !important;
		min-height: auto !important;
		font-size: 1rem !important;
		box-shadow: var(--shadow-small);
	}

	:global(form button[type="submit"]:hover:not(:disabled)) {
		background: color-mix(in srgb, rgb(var(--color-primary)), white 10%) !important;
		border-color: color-mix(in srgb, rgb(var(--color-primary)), white 10%) !important;
	}

	:global(form button[type="submit"]:active:not(:disabled)) {
		background: color-mix(in srgb, rgb(var(--color-primary)), black 8%) !important;
		border-color: color-mix(in srgb, rgb(var(--color-primary)), black 15%) !important;
		box-shadow: none;
	}

	/* Secondary button styling */
	:global(button[type="button"]) {
		border: 2px solid rgb(var(--color-primary)) !important;
		background-color: rgb(var(--color-foreground)) !important;
		color: rgb(var(--color-text)) !important;
		padding: 0.875rem !important;
		min-height: auto !important;
		font-size: 1rem !important;
		box-shadow: var(--shadow-small);
	}

	:global(button[type="button"]:hover:not(:disabled)) {
		background-color: rgb(var(--color-primary) / 0.1) !important;
		border-color: rgb(var(--color-primary)) !important;
	}

	:global(button[type="button"]:active:not(:disabled)) {
		background-color: rgb(var(--color-primary) / 0.2) !important;
		border-color: color-mix(in srgb, rgb(var(--color-primary)), black 10%) !important;
		box-shadow: none;
	}

	/* Dark mode overrides */
	:global(html.dark input) {
		background-color: rgba(255, 255, 255, 0.05) !important;
		border-color: rgb(var(--color-primary) / 0.3) !important;
	}

	:global(html.dark input:focus) {
		background-color: rgb(var(--color-secondary) / 0.1) !important;
		border-color: rgb(var(--color-secondary)) !important;
		box-shadow: 0 0 15px rgb(var(--color-secondary) / 0.3) !important;
	}

	:global(html.dark) section {
		border-color: rgb(var(--color-primary) / 0.4);
	}

	:global(html.dark) .bg-red-50 {
		background-color: rgba(127, 29, 29, 0.2);
	}

	:global(html.dark) .border-red-200 {
		border-color: rgb(127, 29, 29);
	}

	:global(html.dark) .bg-gray-300 {
		background-color: rgb(75, 85, 99);
	}

</style>

<PageTitle title="Sign in" />

<div class="relative min-h-screen overflow-hidden login-bg">
	<!-- Gradient background - positioned to start from form midpoint -->
	<div class="absolute top-0 right-0 bottom-0 w-3/5"></div>

	<!-- Floating background elements -->
	<div class="absolute top-0 right-12 w-96 h-96 bg-white/10 rounded-full blur-3xl -mr-48 -mt-48"></div>
	<div class="absolute bottom-0 right-1/4 w-96 h-96 bg-white/10 rounded-full blur-3xl"></div>

	<div class="fixed right-4 top-4 z-50">
		<ThemeToggle />
	</div>

	<main class="relative z-10 grid min-h-screen grid-cols-1 lg:grid-cols-4 items-center gap-8 lg:gap-0">
		<!-- Left side content -->
		<div class="hidden lg:flex lg:col-span-2 flex-col justify-center px-12 xl:px-20">
			<h2 class="mb-8 text-3xl xl:text-4xl font-bold leading-tight whitespace-nowrap">
				<span class="text-primary">{branding.name.toUpperCase()}</span>
				<span class="mx-1">-</span>
				<span>Schedule Better</span>
			</h2>
			<p class="text-lg mb-8 max-w-lg">{branding.name} makes team scheduling simple and transparent. Share availability once. Get things booked faster.</p>
			<div class="space-y-4">
				<div class="flex items-start gap-3">
					<span class="text-primary text-2xl">✓</span>
					<span>Simple availability sharing</span>
				</div>
				<div class="flex items-start gap-3">
					<span class="text-primary text-2xl">✓</span>
					<span>Instant scheduling</span>
				</div>
				<div class="flex items-start gap-3">
					<span class="text-primary text-2xl">✓</span>
					<span>Team coordination</span>
				</div>
			</div>
		</div>

		<!-- Right side form -->
		<section class="w-full max-w-md mx-auto lg:mx-0 lg:col-span-2 px-4 lg:pl-8 xl:pl-12" aria-labelledby="login-title">
			<div class="p-8 sm:p-10 rounded-2xl border-2 border-border" style="background: rgb(var(--color-foreground)); box-shadow: var(--shadow);">
				<div class="mb-8">
					<BrandLogo class="mb-6 size-20 border border-black object-cover" />
					<h1 id="login-title" class="text-secondary text-3xl font-normal tracking-tight">Welcome Back</h1>
					<p class="mt-2 text-sm">Sign in to manage your team's schedule</p>
				</div>

				{#if error}
					<p class="mb-5 border border-red-200 bg-red-50 p-3 text-sm rounded-lg" role="alert">{error}</p>
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
					<Button type="submit" fullWidth class="mt-2" disabled={submitting}>
						{submitting ? 'Signing in…' : 'Sign in'}
					</Button>
				</form>

				{#if googleEnabled}
					<div class="my-4 flex items-center gap-3" aria-hidden="true">
						<div class="h-px flex-1 bg-gray-300"></div>
						<span class="text-sm font-medium opacity-50">or</span>
						<div class="h-px flex-1 bg-gray-300"></div>
					</div>
					<Button variant="secondary" fullWidth class="!p-[11px]" onclick={googleLogin}>
						<span class="flex items-center gap-2">
							<Icon class="self-center" icon={googleIcon} width="28" height="28" />
							<span class="flex flex-col items-start leading-tight">
								<span class="-mt-1 text-base leading-[1.2]">Continue with Google</span>
								<span class="text-sm font-normal leading-none opacity-50">to get access to calendar</span>
							</span>
						</span>
					</Button>
				{/if}
			</div>
		</section>
	</main>
</div>
