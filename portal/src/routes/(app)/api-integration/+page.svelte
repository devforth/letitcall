<script lang="ts">
	import { onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import copyIcon from '@iconify-icons/tabler/copy';
	import externalLinkIcon from '@iconify-icons/tabler/external-link';
	import { callApi } from '$lib/api';
	import type { APIIntegration, APITokenSummary } from '$lib/types';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import ConfirmationDialog from '$lib/components/ui/ConfirmationDialog.svelte';
	import Input from '$lib/components/ui/Input.svelte';

	let integration = $state<APIIntegration | null>(null);
	let name = $state('');
	let generatedToken = $state('');
	let loading = $state(true);
	let creating = $state(false);
	let revoking = $state(false);
	let tokenToRevoke = $state<APITokenSummary | null>(null);

	onMount(async () => {
		try {
			integration = await callApi<APIIntegration>('/api/integration');
		} catch {
			// callApi reports the error globally.
		} finally {
			loading = false;
		}
	});

	async function createToken(event: SubmitEvent) {
		event.preventDefault();
		creating = true;
		try {
			const response = await callApi<{ apiToken: APITokenSummary; token: string }>(
				'/api/integration/tokens',
				{ method: 'POST', body: JSON.stringify({ name }) }
			);
			if (integration) integration.tokens = [response.apiToken, ...integration.tokens];
			generatedToken = response.token;
			name = '';
		} catch {
			// callApi reports the error globally.
		} finally {
			creating = false;
		}
	}

	async function revokeToken() {
		if (!tokenToRevoke || !integration) return;
		revoking = true;
		try {
			await callApi(`/api/integration/tokens/${encodeURIComponent(tokenToRevoke.id)}`, {
				method: 'DELETE'
			});
			integration.tokens = integration.tokens.filter((token) => token.id !== tokenToRevoke?.id);
			tokenToRevoke = null;
		} catch {
			// callApi reports the error globally.
		} finally {
			revoking = false;
		}
	}
</script>

<PageTitle title="API Integration" />

<section aria-labelledby="api-integration-title">
	<div class="mb-6">
		<h1 id="api-integration-title" class="text-2xl font-semibold tracking-tight">API Integration</h1>
		<p class="mt-2 text-sm">Connect lead-generation and scheduling systems to this installation.</p>
	</div>

	{#if loading}
		<p class="border border-black p-6 text-sm">Loading API integration…</p>
	{:else if integration}
		<div class="grid max-w-4xl gap-8">
			<section class="grid gap-4 border border-black p-5" aria-labelledby="connection-title">
				<h2 id="connection-title" class="text-lg font-semibold">Connection</h2>
				<Input id="api-base-url" label="Base URL" value={integration.baseURL} readonly />
				<div class="flex flex-wrap gap-3 text-sm">
					<a
						class="inline-flex min-h-11 items-center gap-2 border border-black px-4 py-2 font-medium hover:bg-black hover:text-white"
						href={integration.swaggerURL}
						target="_blank"
						rel="noreferrer"
					>
						Swagger documentation <Icon icon={externalLinkIcon} class="size-4" />
					</a>
					<a
						class="inline-flex min-h-11 items-center gap-2 border border-black px-4 py-2 font-medium hover:bg-black hover:text-white"
						href={integration.openAPIURL}
						target="_blank"
						rel="noreferrer"
					>
						OpenAPI JSON <Icon icon={externalLinkIcon} class="size-4" />
					</a>
				</div>
			</section>

			<section class="grid gap-5 border border-black p-5" aria-labelledby="tokens-title">
				<div>
					<h2 id="tokens-title" class="text-lg font-semibold">Personal access tokens</h2>
					<p class="mt-2 text-sm">Use a token as a bearer credential. Each secret is shown only once.</p>
				</div>

				<form class="flex flex-col items-stretch gap-3 sm:flex-row sm:items-end" onsubmit={createToken}>
					<div class="grow">
						<Input id="token-name" label="Token name" bind:value={name} required />
					</div>
					<Button type="submit" disabled={creating}>{creating ? 'Generating…' : 'Generate token'}</Button>
				</form>

				{#if generatedToken}
					<div class="grid gap-3 border border-black p-4" role="status">
						<p class="text-sm font-medium">Copy this token now. It cannot be shown again.</p>
						<Input id="generated-api-token" label="New token" value={generatedToken} readonly />
						<div class="flex flex-wrap gap-2">
							<Button variant="secondary" onclick={() => navigator.clipboard.writeText(generatedToken)}>
								<Icon icon={copyIcon} class="mr-2 size-4" /> Copy token
							</Button>
							<Button variant="secondary" onclick={() => (generatedToken = '')}>Dismiss</Button>
						</div>
					</div>
				{/if}

				{#if integration.tokens.length === 0}
					<p class="border border-black p-4 text-sm">No tokens have been generated.</p>
				{:else}
					<ul class="grid gap-3">
						{#each integration.tokens as token (token.id)}
							<li class="flex flex-wrap items-center justify-between gap-4 border border-black p-4">
								<div>
									<p class="font-medium">{token.name}</p>
									<p class="mt-1 text-sm">Created {new Date(token.createdAt).toLocaleString()}</p>
								</div>
								<Button variant="danger" onclick={() => (tokenToRevoke = token)}>Revoke</Button>
							</li>
						{/each}
					</ul>
				{/if}
			</section>
		</div>
	{/if}
</section>

<ConfirmationDialog
	open={tokenToRevoke !== null}
	title="Revoke API token?"
	description={`Revoke ${tokenToRevoke?.name ?? 'this token'}? Systems using it will immediately lose access.`}
	confirmLabel="Revoke token"
	confirmingLabel="Revoking…"
	confirming={revoking}
	onconfirm={revokeToken}
	oncancel={() => (tokenToRevoke = null)}
/>
