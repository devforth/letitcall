<script lang="ts">
	import { onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import keyIcon from '@iconify-icons/at-icons/key';
	import bookIcon from '@iconify-icons/tabler/book';
	import bracesIcon from '@iconify-icons/tabler/braces';
	import checkIcon from '@iconify-icons/tabler/check';
	import codeIcon from '@iconify-icons/tabler/code';
	import copyIcon from '@iconify-icons/tabler/copy';
	import externalLinkIcon from '@iconify-icons/tabler/external-link';
	import linkIcon from '@iconify-icons/tabler/link';
	import trashIcon from '@iconify-icons/tabler/trash';
	import sortIcon from '@iconify-icons/tabler/arrows-sort';
	import { callApi } from '$lib/api';
	import { nextSort } from '$lib/sort';
	import type { APIIntegration, APITokenSummary } from '$lib/types';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import ConfirmationDialog from '$lib/components/ui/ConfirmationDialog.svelte';
	import CopyButton from '$lib/components/ui/CopyButton.svelte';
	import Dialog from '$lib/components/ui/Dialog.svelte';
	import IconButton from '$lib/components/ui/IconButton.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import SearchableSelect from '$lib/components/ui/SearchableSelect.svelte';
	import SortButton from '$lib/components/ui/SortButton.svelte';
	import TimedActions from '$lib/components/ui/TimedActions.svelte';

	let integration = $state<APIIntegration | null>(null);
	let name = $state('');
	let generatedToken = $state('');
	let generatedTokenName = $state('');
	let tokenDialogOpen = $state(false);
	let tokenSort = $state('createdAt-descending');
	const tokenSortKey = $derived(tokenSort.split('-')[0] as 'name' | 'createdAt');
	const tokenSortDirection = $derived(tokenSort.split('-')[1] as 'ascending' | 'descending');
	const tokenSortOptions = [
		{ value: 'createdAt-descending', label: 'Newest first' },
		{ value: 'createdAt-ascending', label: 'Oldest first' },
		{ value: 'name-ascending', label: 'Name, A-Z' },
		{ value: 'name-descending', label: 'Name, Z-A' }
	];
	let loading = $state(true);
	let creating = $state(false);
	let revoking = $state(false);
	let tokenToRevoke = $state<APITokenSummary | null>(null);
	const dateTimeFormat = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' });
	const sortedTokens = $derived([...(integration?.tokens ?? [])].sort(compareTokens));

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
			generatedTokenName = response.apiToken.name;
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

	function formatDateTime(value: string) {
		return dateTimeFormat.format(new Date(value));
	}

	function compareTokens(first: APITokenSummary, second: APITokenSummary) {
		const comparison =
			tokenSortKey === 'name'
				? first.name.localeCompare(second.name)
				: first.createdAt.localeCompare(second.createdAt);
		return tokenSortDirection === 'ascending' ? comparison : -comparison;
	}

	function toggleTokenSort(key: 'name' | 'createdAt') {
		tokenSort = nextSort(tokenSort, key);
	}

	function dismissGeneratedToken() {
		tokenDialogOpen = false;
		name = '';
		generatedToken = '';
		generatedTokenName = '';
	}
</script>

<PageTitle title="API Integration" />

<section aria-labelledby="api-integration-title" class="flex flex-col gap-6">
	<div class="mb-2">
		<PageHeader
			id="api-integration-title"
			title="API Integration"
			description="Connect lead-generation and scheduling systems to this installation."
			icon={codeIcon}
		/>
	</div>

	{#if loading}
		<p class="outlined-block p-6 text-sm">Loading API integration…</p>
	{:else if integration}
		{@const baseURL = integration.baseURL}
		{@const swaggerURL = integration.swaggerURL}
		{@const openAPIURL = integration.openAPIURL}
		<div class="grid gap-8">
			<section class="connection-section" aria-label="Connection">
				<div class="connection-cards">
					<div class="connection-card base-url-card">
						<CopyButton class="base-url-copy-button" value={baseURL} label="Copy base URL" />
						<div class="base-url-title">
							<span class="connection-title-row">
								<span class="connection-title-icon"><Icon icon={linkIcon} width="16" height="16" /></span>
								<span class="connection-card-title">Connection</span>
							</span>
							<span class="connection-card-description">Base API</span>
						</div>
						<code class="connection-url">{baseURL}</code>
					</div>
					<div class="connection-card documentation-card">
						<IconButton
							class="documentation-open-button"
							filled
							label="Open Swagger documentation"
							onclick={() => window.open(swaggerURL, '_blank', 'noopener,noreferrer')}
						>
							<Icon icon={externalLinkIcon} width="22" height="22" />
						</IconButton>
						<span>
							<span class="connection-title-row">
								<span class="connection-title-icon"><Icon icon={bookIcon} width="16" height="16" /></span>
								<span class="connection-card-title">Swagger documentation</span>
							</span>
							<span class="connection-card-description">Interactive reference</span>
						</span>
					</div>
					<div class="connection-card documentation-card">
						<IconButton
							class="documentation-open-button"
							filled
							label="Open OpenAPI JSON"
							onclick={() => window.open(openAPIURL, '_blank', 'noopener,noreferrer')}
						>
							<Icon icon={externalLinkIcon} width="22" height="22" />
						</IconButton>
						<span>
							<span class="connection-title-row">
								<span class="connection-title-icon"><Icon icon={bracesIcon} width="16" height="16" /></span>
								<span class="connection-card-title">OpenAPI JSON</span>
							</span>
							<span class="connection-card-description">Raw schema</span>
						</span>
					</div>
				</div>
			</section>

			<section class="tokens-section grid gap-5" aria-labelledby="tokens-title">
				<div class="token-table-wrap overflow-hidden rounded-lg">
						<div class="overflow-x-auto">
							<table class="token-table w-full min-w-[36rem] text-left text-sm">
								<thead>
									<tr>
										<th class="p-4" colspan="3">
											<div class="flex flex-wrap items-center justify-between gap-4">
												<div class="text-left">
													<h2 id="tokens-title" class="text-base font-semibold">
														Personal access tokens
														{#if integration.tokens.length > 0}
															<span style="color: rgb(var(--color-text) / 0.65);">· {integration.tokens.length}</span>
														{/if}
													</h2>
													<p class="mt-1 text-sm font-normal" style="color: rgb(var(--color-text) / 0.65);">Use a token as a bearer credential</p>
												</div>
												<Button
													rounded
													class="primary-action-button"
													onclick={() => (tokenDialogOpen = true)}
												>
													<span class="flex items-center gap-2">
														<Icon icon={keyIcon} width="20" height="20" />
														Generate token
													</span>
												</Button>
											</div>
											{#if integration.tokens.length > 0}
												<div class="token-sort-select mt-4">
													<SearchableSelect id="token-sort" label="Sort by" icon={sortIcon} options={tokenSortOptions} bind:value={tokenSort} clearable={false} required />
												</div>
											{/if}
										</th>
									</tr>
									{#if integration.tokens.length > 0}
										<tr>
											<th aria-sort={tokenSortKey === 'name' ? tokenSortDirection : 'none'} class="px-4 py-3">
											<SortButton label="Name" active={tokenSortKey === 'name'} direction={tokenSortDirection} onclick={() => toggleTokenSort('name')} />
											</th>
											<th aria-sort={tokenSortKey === 'createdAt' ? tokenSortDirection : 'none'} class="px-4 py-3">
											<SortButton label="Created" active={tokenSortKey === 'createdAt'} direction={tokenSortDirection} onclick={() => toggleTokenSort('createdAt')} />
											</th>
											<th class="px-4 py-3"><span class="sr-only">Actions</span></th>
										</tr>
									{/if}
								</thead>
								<tbody>
									{#each sortedTokens as token (token.id)}
										<tr data-timed-actions-row>
											<td class="px-4 py-3 font-medium">{token.name}</td>
											<td class="px-4 py-3"><time datetime={token.createdAt}>{formatDateTime(token.createdAt)}</time></td>
											<td class="px-4 py-3 text-right">
												<TimedActions
													label={`Show actions for ${token.name}`}
													controlsId={`token-actions-${token.id}`}
													size="2.75rem"
													actionsVisibleOnSmallScreens
												>
													<div class="token-actions">
														<IconButton class="revoke-token-button" filled label="Revoke token" onclick={() => (tokenToRevoke = token)}>
															<Icon icon={trashIcon} width="20" height="20" />
														</IconButton>
													</div>
												</TimedActions>
											</td>
										</tr>
									{:else}
										<tr>
											<td class="no-tokens-cell" colspan="3">No tokens have been generated</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					</div>
			</section>
		</div>
	{/if}
</section>

<Dialog
	open={tokenDialogOpen}
	wide
	bare
	sheet
	label={generatedToken ? `Token “${generatedTokenName}” created` : 'Generate token'}
	oncancel={dismissGeneratedToken}
>
	<div class="generated-token-dialog" role="status">
		{#if generatedToken}
			<div class="generated-token-dialog-body">
				<h2 class="generated-token-dialog-title">
					<Icon icon={checkIcon} width="22" height="22" aria-hidden="true" />
					Token “{generatedTokenName}” created
				</h2>
				<p class="generated-token-dialog-description">This is the only time the secret is shown — store it somewhere safe before closing</p>
				<div class="generated-token-value-row">
					<code class="generated-token-value">{generatedToken}</code>
					<CopyButton class="generated-token-copy-button" value={generatedToken} label="Copy token" />
				</div>
			</div>
			<div class="generated-token-dialog-footer">
				<Button rounded class="modal-action-button" onclick={dismissGeneratedToken}>
					I saved it, close
				</Button>
			</div>
		{:else}
			<form onsubmit={createToken}>
				<div class="generated-token-dialog-body grid gap-4">
					<div>
						<h2 class="generated-token-dialog-title">Generate token</h2>
						<p class="generated-token-dialog-description">Choose a name so you can recognize this token later</p>
					</div>
					<Input id="token-name" label="Token name" bind:value={name} icon="none" required />
				</div>
				<div class="generated-token-dialog-footer gap-3">
					<Button rounded variant="primary-outline" class="modal-action-button" type="button" onclick={dismissGeneratedToken}>Cancel</Button>
					<Button rounded class="modal-action-button" type="submit" disabled={creating}>
						{creating ? 'Generating…' : 'Generate'}
					</Button>
				</div>
			</form>
		{/if}
	</div>
</Dialog>

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

<style>
	.outlined-block {
		border: 0;
		border-radius: 8px;
		background: rgb(var(--color-background));
		box-shadow: 0 0 0 1px var(--color-border);
	}

	.token-table-wrap {
		background: rgb(var(--color-background));
		box-shadow: 0 0 0 1px var(--color-border);
	}

	.token-table {
		border-collapse: collapse;
	}

	.token-table thead tr:first-child {
		background: linear-gradient(to bottom, transparent 75%, color-mix(in srgb, var(--color-border) 30%, transparent));
	}

	.token-table th {
		border-bottom: 1px solid var(--color-border);
		color: rgb(var(--color-text));
		font-size: 0.75rem;
		font-weight: 400;
		letter-spacing: 0.025em;
	}

	.token-table tbody tr {
		border-bottom: 1px solid var(--color-border);
	}

	.token-table tbody tr:last-child {
		border-bottom: 0;
	}

	.tokens-section {
		container: tokens / inline-size;
	}

	.token-sort-select {
		display: none;
	}

	/* Below the table's 36rem minimum it would scroll, so the header and each token become cards. */
	@container tokens (max-width: 36rem) {
		.token-table-wrap {
			overflow: visible;
			border-radius: 0;
			background: none;
			box-shadow: none;
		}

		.token-table,
		.token-table thead,
		.token-table tbody {
			display: grid;
			gap: 1.25rem;
			min-width: 0;
		}

		.token-table thead tr:first-child,
		.token-table tbody tr {
			display: block;
			border: 0;
			border-radius: 0.5rem;
			/* Inset: the table's scroll wrapper clips anything drawn outside the row. */
			box-shadow: inset 0 0 0 1px var(--color-border);
		}

		.token-table thead tr:first-child {
			background:
				linear-gradient(to bottom, transparent 75%, color-mix(in srgb, var(--color-border) 30%, transparent)),
				rgb(var(--color-background));
		}

		.token-table thead tr:first-child th {
			display: block;
			border-bottom: 0;
		}

		.token-table thead tr:nth-child(2) {
			display: none;
		}

		.token-table tbody tr {
			display: grid;
			gap: 0.25rem;
			padding: 1rem 1.25rem 1.25rem;
			background: rgb(var(--color-background));
		}

		.token-table tbody td {
			padding: 0;
		}

		.token-sort-select {
			display: block;
		}
	}

	.no-tokens-cell {
		padding: 2.5rem 1rem;
		color: rgb(var(--color-text) / 0.65);
		font-size: 0.875rem;
		text-align: center;
	}

	.connection-section {
		container: connection / inline-size;
	}

	.connection-cards {
		display: grid;
		/* Wide enough for the longest title beside the corner button; narrower cards wrap. */
		grid-template-columns: repeat(auto-fit, minmax(min(19rem, 100%), 1fr));
		gap: 1rem;
	}

	.connection-card {
		display: flex;
		min-height: 8rem;
		flex-direction: column;
		align-items: flex-start;
		justify-content: space-between;
		border: 1px solid var(--color-border);
		border-radius: 0.5rem;
		padding: 1.25rem;
		background: rgb(var(--color-background));
		color: rgb(var(--color-text));
	}

	.connection-url,
	.generated-token-value {
		overflow: hidden;
		border-radius: 0.5rem;
		padding: 0.375rem 0.5rem;
		background: rgb(var(--color-text) / 0.08);
		color: rgb(var(--color-text) / 0.85);
		font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
		font-size: 0.875rem;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.connection-url {
		display: block;
		width: fit-content;
		max-width: 100%;
	}

	.base-url-card {
		position: relative;
		justify-content: flex-start;
		gap: 0.75rem;
	}

	:global(.base-url-copy-button) {
		position: absolute;
		top: 1.25rem;
		right: 1.25rem;
	}

	:global(.base-url-copy-button),
	:global(.generated-token-copy-button) {
		width: 2.75rem !important;
		height: 2.75rem !important;
	}

	.documentation-card {
		position: relative;
		justify-content: flex-start;
		gap: 0.75rem;
	}

	:global(.documentation-open-button) {
		position: absolute;
		top: 1.25rem;
		right: 1.25rem;
		width: 2.75rem !important;
		height: 2.75rem !important;
	}

	.connection-card-title,
	.connection-card-description {
		display: block;
	}

	.connection-title-row {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
	}

	.connection-title-icon {
		display: inline-flex;
		color: rgb(var(--color-text));
	}

	.connection-card-title {
		font-size: 0.875rem;
		font-weight: 600;
		white-space: nowrap;
	}

	.connection-card-description {
		margin-top: 0.125rem;
		color: rgb(var(--color-text) / 0.65);
		font-size: 0.8125rem;
	}

	.generated-token-dialog {
		overflow: hidden;
		border: 1px solid var(--color-border);
		border-radius: 0.75rem;
		background: rgb(var(--color-background));
		box-shadow: var(--shadow);
	}

	@media (max-width: 39.999rem) {
		.generated-token-dialog {
			border-radius: 1rem 1rem 0 0;
		}
	}

	.generated-token-dialog-body {
		padding: 1.5rem;
	}

	.generated-token-dialog-title {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin: 0;
		color: rgb(var(--color-text));
		font-size: 1.25rem;
		font-weight: 500;
	}

	.generated-token-dialog-title :global(svg) {
		color: rgb(var(--success));
	}

	.generated-token-dialog-description {
		margin: 0.5rem 0 0;
		color: rgb(var(--color-text) / 0.65);
		font-size: 0.875rem;
	}

	.generated-token-value-row {
		display: flex;
		align-items: stretch;
		gap: 0.75rem;
		margin-top: 1.25rem;
	}

	.generated-token-value {
		display: flex;
		min-width: 0;
		flex: 1;
		align-items: center;
	}

	.generated-token-dialog-footer {
		display: flex;
		justify-content: flex-end;
		border-top: 1px solid var(--color-border);
		padding: 1rem 1.5rem;
	}

	@media (max-width: 40rem) {
		.generated-token-dialog-title {
			font-size: 1.125rem;
		}

		.generated-token-value-row {
			flex-direction: column;
		}

		.generated-token-value {
			font-size: 0.875rem;
		}

		.generated-token-dialog-footer {
			padding: 1rem 1.5rem;
		}
	}

	:global(.revoke-token-button) {
		width: 2.75rem !important;
		height: 2.75rem !important;
	}

	/* One card per row (two 19rem cards no longer fit): merge them into a single container. */
	@container connection (max-width: 39rem) {
		.connection-cards {
			gap: 0;
			overflow: hidden;
			border: 1px solid var(--color-border);
			border-radius: 0.5rem;
		}

		.connection-card {
			min-height: 0;
			border: 0;
			border-radius: 0;
		}

		.connection-card + .connection-card {
			border-top: 1px solid var(--color-border);
		}
	}
</style>
