<script lang="ts">
	import Icon from '@iconify/svelte';
	import checkIcon from '@iconify-icons/tabler/check';
	import xIcon from '@iconify-icons/tabler/x';
	import plusIcon from '@iconify-icons/tabler/plus';
	import Button from '$lib/components/ui/Button.svelte';
	import IconButton from '$lib/components/ui/IconButton.svelte';
	import Input from '$lib/components/ui/Input.svelte';

	let {
		idPrefix,
		emails = $bindable(),
		limit = null,
		legend = 'Additional guests',
		addLabel = 'Add guest',
		pending = $bindable(false)
	}: {
		idPrefix: string;
		emails: string[];
		limit?: number | null;
		// Pass null where a surrounding section heading already names the group.
		legend?: string | null;
		addLabel?: string;
		/** True while a typed guest is still waiting to be confirmed or discarded. */
		pending?: boolean;
	} = $props();

	// Icons are drawn a little bolder so they read at the small button size.
	const boldIcon = (icon: { body: string }) => ({
		...icon,
		body: icon.body.replace('stroke-width="2"', 'stroke-width="3"')
	});
	const boldCheckIcon = boldIcon(checkIcon);
	const boldXIcon = boldIcon(xIcon);
	const boldPlusIcon = boldIcon(plusIcon);

	let adding = $state(false);
	let draft = $state('');
	let error = $state('');
	let draftRow = $state<HTMLDivElement>();

	const canAdd = $derived(limit === null || emails.length < limit);

	$effect(() => {
		if (adding) draftRow?.querySelector('input')?.focus();
	});

	// Lets the surrounding form block its own submit until this draft is resolved.
	$effect(() => {
		pending = adding && draft.trim().length > 0;
	});

	function startAdding() {
		draft = '';
		error = '';
		adding = true;
	}

	function cancelAdding() {
		draft = '';
		error = '';
		adding = false;
	}

	function confirmGuest() {
		const email = draft.trim();
		if (!email) {
			error = 'Enter a guest email';
			return;
		}
		if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
			error = 'Enter a valid email address';
			return;
		}
		if (emails.some((existing) => existing.toLowerCase() === email.toLowerCase())) {
			error = 'This guest is already added';
			return;
		}
		emails.push(email);
		cancelAdding();
	}

	function removeGuest(index: number) {
		emails.splice(index, 1);
	}

	// The field can live inside a booking form, so Enter must confirm the guest
	// instead of submitting that form.
	function onKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			event.preventDefault();
			confirmGuest();
		} else if (event.key === 'Escape') {
			event.preventDefault();
			cancelAdding();
		}
	}
</script>

<fieldset class="grid gap-4">
	{#if legend}
		<legend class="text-sm font-medium">{legend}</legend>
	{/if}
	{#if emails.length > 0}
		<ul class="flex flex-wrap gap-2">
			{#each emails as email, index (`${idPrefix}-${email}`)}
				<li class="guest-chip">
					<span class="truncate">{email}</span>
					<button
						type="button"
						class="guest-chip-remove"
						aria-label={`Remove ${email}`}
						title={`Remove ${email}`}
						onclick={() => removeGuest(index)}
					>
						<Icon icon={boldXIcon} width="16" height="16" />
					</button>
				</li>
			{/each}
		</ul>
	{/if}
	{#if adding}
		<!-- Enter/Escape are handled on the row because Input doesn't forward key events. -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="guest-draft grid gap-2 sm:grid-cols-[1fr_auto] sm:gap-3 sm:items-start md:grid-cols-[minmax(0,50%)_auto]"
			bind:this={draftRow}
			onkeydown={onKeydown}
		>
			<!-- A text input with an email keyboard: `type="email"` is what Chrome
			     classifies as fillable, and the format is validated in confirmGuest anyway. -->
			<Input
				id={`${idPrefix}-draft`}
				label="Guest email"
				type="text"
				inputmode="email"
				icon="email"
				bind:value={draft}
				{error}
				noAutofill
				hint={pending ? 'Confirm this guest to continue' : ''}
			/>
			<div class="guest-draft-actions flex gap-2">
				<IconButton tone="primary" label="Add this guest" onclick={confirmGuest}>
					<Icon icon={boldCheckIcon} width="20" height="20" />
				</IconButton>
				<IconButton tone="danger" label="Discard this guest" onclick={cancelAdding}>
					<Icon icon={boldXIcon} width="20" height="20" />
				</IconButton>
			</div>
		</div>
	{:else if canAdd}
		<div>
			<Button class="guest-add-button gap-2" onclick={startAdding}>
				<Icon icon={boldPlusIcon} width="20" height="20" />
				{addLabel}
			</Button>
		</div>
	{/if}
</fieldset>

<style>
	/* Sized like the guest email input — 44px tall with the field's 10px corners. */
	.guest-chip {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		max-width: 100%;
		min-height: 44px;
		border: 1px solid rgb(var(--color-border));
		border-radius: 10px;
		background: rgb(var(--color-border) / 0.4);
		box-shadow: 0 1px 2px rgb(var(--color-shadow) / 0.06);
		padding: 0.375rem 0.5rem 0.375rem 0.875rem;
		font-size: 0.875rem;
		font-weight: 600;
		line-height: 1.25;
	}

	.guest-chip-remove {
		display: grid;
		place-items: center;
		flex-shrink: 0;
		width: 1.75rem;
		height: 1.75rem;
		border-radius: 8px;
		cursor: pointer;
		color: rgb(var(--color-text) / 0.7);
		transition:
			color 0.15s ease,
			background-color 0.15s ease;
	}

	.guest-chip:hover .guest-chip-remove {
		color: rgb(var(--color-text));
	}

	.guest-chip-remove:hover,
	.guest-chip-remove:focus-visible {
		background: rgb(var(--error) / 0.14);
		color: rgb(var(--error));
	}

	/* Square blocks sized a little past the 44px input height so they read as solid actions. */
	.guest-draft-actions :global(.icon-button) {
		width: 48px;
		height: 48px;
		border: 0;
		border-radius: 11px;
	}

	/* Pull up by half the overhang so they stay centered on the input, not on its error text. */
	@media (min-width: 640px) {
		.guest-draft-actions {
			margin-top: -2px;
		}
	}

	/* Confirm/discard keep their tinted look without hover, like table row actions. */
	.guest-draft-actions :global(.tone-primary:not(:disabled)) {
		background: rgb(var(--color-primary) / 0.14);
		color: rgb(var(--color-primary));
	}

	.guest-draft-actions :global(.tone-danger:not(:disabled)) {
		background: rgb(var(--error) / 0.14);
		color: rgb(var(--error));
	}

	@media (prefers-reduced-motion: reduce) {
		.guest-chip-remove {
			transition: none;
		}
	}
</style>
