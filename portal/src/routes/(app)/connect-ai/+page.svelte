<script lang="ts">
	import { onMount } from 'svelte';
	import Icon from '@iconify/svelte';
	import plugConnectedIcon from '@iconify-icons/tabler/plug-connected';
	import robotIcon from '@iconify-icons/tabler/robot';
	import trashIcon from '@iconify-icons/tabler/trash';
	import { callApi } from '$lib/api';
	import { branding } from '$lib/stores/branding.svelte';
	import { browserTimeFormatter } from '$lib/time-format';
	import type { MCPConnection } from '$lib/types';
	import PageHeader from '$lib/components/PageHeader.svelte';
	import PageTitle from '$lib/components/PageTitle.svelte';
	import CodeSnippet from '$lib/components/ui/CodeSnippet.svelte';
	import ConfirmationDialog from '$lib/components/ui/ConfirmationDialog.svelte';
	import IconButton from '$lib/components/ui/IconButton.svelte';
	import SegmentedControl from '$lib/components/ui/SegmentedControl.svelte';
	import TimedActions from '$lib/components/ui/TimedActions.svelte';

	let connections = $state<MCPConnection[]>([]);
	let mcpURL = $state('');
	let loading = $state(true);
	let client = $state('claude-code');
	let connectionToRevoke = $state<MCPConnection | null>(null);
	let revoking = $state(false);

	const blockStyle =
		'background: rgb(var(--color-background)); box-shadow: 0 0 0 1px var(--color-border);';
	const dateTimeFormat = browserTimeFormatter({ dateStyle: 'medium', timeStyle: 'short' });
	const clients = [
		{ value: 'claude-code', label: 'Claude Code' },
		{ value: 'codex', label: 'Codex' }
	];
	// The name the agent registers this server under; a brand of non-Latin letters only leaves it empty.
	const serverName = $derived(
		branding.name.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '') || 'letitcall'
	);
	const steps = $derived(
		client === 'claude-code'
			? [
					{ hint: 'Run in your terminal', snippet: `claude mcp add --transport http ${serverName} ${mcpURL}` },
					{ hint: `Then run in Claude Code, pick ${serverName} and choose Authenticate`, snippet: '/mcp' }
				]
			: [
					{
						hint: `Run in your terminal. Codex opens ${branding.name} in your browser to sign in; if it does not, run codex mcp login ${serverName}`,
						snippet: `codex mcp add ${serverName} --url ${mcpURL}`
					}
				]
	);

	onMount(load);

	// Agents connect in another tab, so the list refreshes when this one is focused again.
	async function load() {
		try {
			const response = await callApi<{ mcpURL: string; connections: MCPConnection[] }>('/api/mcp/connections');
			mcpURL = response.mcpURL;
			connections = response.connections;
		} catch {
			// callApi reports the error globally.
		} finally {
			loading = false;
		}
	}

	async function revoke() {
		const connection = connectionToRevoke!;
		revoking = true;
		try {
			await callApi(`/api/mcp/connections/${encodeURIComponent(connection.id)}`, { method: 'DELETE' });
			connections = connections.filter((candidate) => candidate.id !== connection.id);
			connectionToRevoke = null;
		} catch {
			// callApi reports the error globally.
		} finally {
			revoking = false;
		}
	}
</script>

<svelte:window onfocus={load} />

<PageTitle title="Connect AI" />

<section aria-labelledby="connect-ai-title" class="flex flex-col gap-6">
	<div class="mb-2">
		<PageHeader
			id="connect-ai-title"
			title="Connect AI"
			description={`Connect an AI agent to ${branding.name} over MCP. It acts on your behalf with your permissions.`}
			icon={robotIcon}
		/>
	</div>

	{#if loading}
		<p class="p-8 text-sm" style="color: rgb(var(--color-text) / 0.65);">Loading AI connections…</p>
	{:else}
		<section class="grid gap-4 rounded-lg p-4 sm:p-5" style={blockStyle} aria-labelledby="agent-setup-title">
			<div class="flex flex-wrap items-center justify-between gap-3">
				<h2 id="agent-setup-title" class="font-semibold">Connect an agent</h2>
				<SegmentedControl options={clients} value={client} label="AI agent" onchange={(value) => (client = value)} />
			</div>
			{#each steps as step (step.snippet)}
				<div class="grid gap-2">
					<p class="text-sm" style="color: rgb(var(--color-text) / 0.65);">{step.hint}</p>
					<CodeSnippet value={step.snippet} label="Copy command" />
				</div>
			{/each}
			<div class="grid gap-2">
				<p class="text-sm" style="color: rgb(var(--color-text) / 0.65);">Other agents: add this remote MCP server and sign in with OAuth</p>
				<CodeSnippet value={mcpURL} label="Copy MCP server URL" />
			</div>
		</section>

		<section class="grid gap-5" aria-labelledby="connections-title">
			<h2 id="connections-title" class="font-semibold">
				Connected agents
				{#if connections.length > 0}
					<span style="color: rgb(var(--color-text) / 0.65);">· {connections.length}</span>
				{/if}
			</h2>
			{#each connections as connection (connection.id)}
				<article data-timed-actions-row class="relative flex items-center gap-4 rounded-lg p-4 sm:p-5" style={blockStyle}>
					<Icon icon={plugConnectedIcon} width="24" height="24" class="shrink-0" />
					<div class="min-w-0 flex-1">
						<h3 class="truncate font-semibold">{connection.clientName}</h3>
						<p class="truncate text-sm" style="color: rgb(var(--color-text) / 0.65);">
							{new URL(connection.clientId).host} · connected {dateTimeFormat.format(new Date(connection.createdAt))}
						</p>
					</div>
					<TimedActions label={`Show actions for ${connection.clientName}`} controlsId={`connection-actions-${connection.id}`} actionsVisibleOnSmallScreens>
						<IconButton filled tone="danger" label={`Disconnect ${connection.clientName}`} onclick={() => (connectionToRevoke = connection)}>
							<Icon icon={trashIcon} width="20" height="20" />
						</IconButton>
					</TimedActions>
				</article>
			{:else}
				<p class="rounded-lg px-4 py-10 text-center text-sm" style="{blockStyle} color: rgb(var(--color-text) / 0.65);">No agents are connected</p>
			{/each}
		</section>
	{/if}
</section>

<ConfirmationDialog
	open={connectionToRevoke !== null}
	title="Disconnect AI agent?"
	description={`${connectionToRevoke?.clientName ?? 'This agent'} immediately loses access to ${branding.name}.`}
	confirmLabel="Disconnect"
	confirmingLabel="Disconnecting…"
	confirming={revoking}
	onconfirm={revoke}
	oncancel={() => (connectionToRevoke = null)}
/>
