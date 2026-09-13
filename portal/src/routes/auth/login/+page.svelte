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
	let flashlightX = $state(-160);
	let flashlightY = $state(-160);
	let flashlightStarted = $state(false);

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
		} catch (cause) {
			error = cause instanceof Error ? cause.message : 'Unable to load login settings';
		}
	});

	onMount(() => {
		window.addEventListener('pointermove', moveFlashlight);
		window.addEventListener('blur', hideFlashlight);
		return () => {
			window.removeEventListener('pointermove', moveFlashlight);
			window.removeEventListener('blur', hideFlashlight);
		};
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

	function moveFlashlight(event: PointerEvent) {
		flashlightStarted = true;
		flashlightX = event.clientX;
		flashlightY = event.clientY;
	}

	function hideFlashlight() {
		flashlightX = -160;
		flashlightY = -160;
	}
</script>

<style>
	:global(.login-bg) {
		background-color: rgb(var(--color-foreground));
		background-image: radial-gradient(circle, rgb(var(--color-text) / 0.08) 1.5px, transparent 1.5px);
		background-size: 18px 18px;
	}

	.background-dot-layer {
		background-image: radial-gradient(circle, rgb(var(--color-primary) / 0.5) 1.5px, transparent 1.5px);
		background-size: 18px 18px;
		mask-image: radial-gradient(circle 9rem at var(--flashlight-x) var(--flashlight-y), black 0%, black 35%, rgb(0 0 0 / 0.65) 56%, transparent 100%);
		-webkit-mask-image: radial-gradient(circle 9rem at var(--flashlight-x) var(--flashlight-y), black 0%, black 35%, rgb(0 0 0 / 0.65) 56%, transparent 100%);
	}

	.background-dot-layer.initial {
		mask-image:
			radial-gradient(circle 9rem at center, black 0%, black 35%, rgb(0 0 0 / 0.65) 56%, transparent 100%),
			radial-gradient(circle 9rem at center, black 0%, black 35%, rgb(0 0 0 / 0.65) 56%, transparent 100%);
		mask-position: 8% 12%, 82% 72%;
		mask-repeat: no-repeat;
		mask-size: 18rem 18rem, 18rem 18rem;
		-webkit-mask-image:
			radial-gradient(circle 9rem at center, black 0%, black 35%, rgb(0 0 0 / 0.65) 56%, transparent 100%),
			radial-gradient(circle 9rem at center, black 0%, black 35%, rgb(0 0 0 / 0.65) 56%, transparent 100%);
		-webkit-mask-position: 8% 12%, 82% 72%;
		-webkit-mask-repeat: no-repeat;
		-webkit-mask-size: 18rem 18rem, 18rem 18rem;
		animation: initial-flashlight 12s ease-in-out infinite;
	}

	.background-dot-core {
		background-image: radial-gradient(circle, rgb(var(--color-primary) / 0.5) 2.5px, transparent 2.5px);
		background-size: 18px 18px;
		mask-image: radial-gradient(circle 4rem at var(--flashlight-x) var(--flashlight-y), black 0%, black 35%, rgb(0 0 0 / 0.65) 56%, transparent 100%);
		-webkit-mask-image: radial-gradient(circle 4rem at var(--flashlight-x) var(--flashlight-y), black 0%, black 35%, rgb(0 0 0 / 0.65) 56%, transparent 100%);
	}

	.background-dot-core.initial {
		mask-image:
			radial-gradient(circle 4rem at center, black 0%, black 35%, rgb(0 0 0 / 0.65) 56%, transparent 100%),
			radial-gradient(circle 4rem at center, black 0%, black 35%, rgb(0 0 0 / 0.65) 56%, transparent 100%);
		mask-position: 8% 12%, 82% 72%;
		mask-repeat: no-repeat;
		mask-size: 18rem 18rem, 18rem 18rem;
		-webkit-mask-image:
			radial-gradient(circle 4rem at center, black 0%, black 35%, rgb(0 0 0 / 0.65) 56%, transparent 100%),
			radial-gradient(circle 4rem at center, black 0%, black 35%, rgb(0 0 0 / 0.65) 56%, transparent 100%);
		-webkit-mask-position: 8% 12%, 82% 72%;
		-webkit-mask-repeat: no-repeat;
		-webkit-mask-size: 18rem 18rem, 18rem 18rem;
		animation: initial-flashlight 12s ease-in-out infinite;
	}

	.login-form-panel {
		background: rgb(var(--color-foreground));
		box-shadow: 0 0 0 1px rgb(var(--color-border));
	}

	.login-theme-toggle {
		background: rgb(var(--color-foreground));
	}

	@keyframes initial-flashlight {
		0%, 100% {
			mask-position: 8% 12%, 82% 72%;
			-webkit-mask-position: 8% 12%, 82% 72%;
		}

		30% {
			mask-position: 72% 20%, 18% 82%;
			-webkit-mask-position: 72% 20%, 18% 82%;
		}

		58% {
			mask-position: 84% 78%, 28% 16%;
			-webkit-mask-position: 84% 78%, 28% 16%;
		}

		82% {
			mask-position: 26% 68%, 70% 42%;
			-webkit-mask-position: 26% 68%, 70% 42%;
		}
	}

	/* Label styling */
	:global(label span) {
		font-weight: 600 !important;
	}

	/* Dark mode overrides */
	:global(html.dark) .bg-red-50 {
		background-color: rgba(127, 29, 29, 0.2);
	}

	:global(html.dark) .border-red-200 {
		border-color: rgb(127, 29, 29);
	}

</style>

<PageTitle title="Sign in" />

<div
	class="relative min-h-screen overflow-hidden login-bg"
	style={`--flashlight-x: ${flashlightX}px; --flashlight-y: ${flashlightY}px;`}
>
	<div class:initial={!flashlightStarted} class="background-dot-layer pointer-events-none absolute inset-0" aria-hidden="true"></div>
	<div class:initial={!flashlightStarted} class="background-dot-core pointer-events-none absolute inset-0" aria-hidden="true"></div>
	<div class="login-theme-toggle fixed right-4 top-4 z-20 flex items-center gap-2 px-3 py-2">
		<span class="text-xs opacity-45">Theme</span>
		<ThemeToggle compact />
	</div>

	<main class="relative z-10 grid min-h-screen place-items-center p-4">
		<section class="w-full max-w-md" aria-label="Sign in">
			<div class="login-form-panel p-8 sm:p-10">
				<div class="mb-14 flex flex-col items-center text-center">
					<div class="flex items-center justify-center gap-4">
						<BrandLogo class="size-12 object-cover" />
						<p class="text-3xl font-bold leading-none text-primary">{branding.name.toUpperCase()}</p>
					</div>
					<p class="mt-3 text-2xl font-normal leading-tight">Scheduling Admin Panel</p>
				</div>

				{#if error}
					<p class="mb-5 border border-red-200 bg-red-50 p-3 text-sm rounded-lg" role="alert">{error}</p>
				{/if}

				<form class="grid gap-5" onsubmit={login}>
					<Input id="email" label="Email or username" icon="user" bind:value={email} required autocomplete="username" />
					<Input
						id="password"
						label="Password"
						type="password"
						bind:value={password}
						required
						autocomplete="current-password"
					/>
					<Button type="submit" fullWidth class="mt-8 lg-pd" disabled={submitting}>
						{submitting ? 'Signing in…' : 'Sign in'}
					</Button>
				</form>

				{#if googleEnabled}
					<div class="my-4 flex items-center gap-3" aria-hidden="true">
						<div class="h-px flex-1 bg-border"></div>
						<span class="text-sm font-medium opacity-50">or</span>
						<div class="h-px flex-1 bg-border"></div>
					</div>
					<Button variant="secondary" fullWidth class="lg-pd" onclick={googleLogin}>
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
