<script lang="ts">
	import Icon from '@iconify/svelte';
	import removeIcon from '@iconify-icons/pajamas/remove';
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
		invalidEmails = []
	}: {
		idPrefix: string;
		emails: string[];
		limit?: number | null;
		legend?: string | null;
		addLabel?: string;
		invalidEmails?: boolean[];
	} = $props();

	const boldIcon = (icon: { body: string }) => ({
		...icon,
		body: icon.body.replace('stroke-width="2"', 'stroke-width="3"')
	});
	const boldPlusIcon = boldIcon(plusIcon);

	const canAdd = $derived(limit === null || emails.length < limit);

	function addGuest() {
		emails.push('');
	}

	function removeGuest(index: number) {
		emails.splice(index, 1);
	}
</script>

<fieldset class="grid gap-4">
	{#if legend}
		<legend class="text-sm font-medium">{legend}</legend>
	{/if}
	<div class="grid gap-3">
		{#if emails.length > 0}
			{#each emails as _, index (index)}
				<div class="guest-email-row grid grid-cols-[minmax(0,1fr)_auto] items-start gap-3 sm:grid-cols-[minmax(0,50%)_auto]">
					<Input
						id={`${idPrefix}-${index}`}
						label={`Guest email ${index + 1}`}
						type="email"
						icon="email"
						bind:value={emails[index]}
						required
						invalid={invalidEmails[index] ?? false}
						noAutofill
					/>
					<IconButton tone="danger" label={`Remove guest ${index + 1}`} onclick={() => removeGuest(index)}>
						<Icon icon={removeIcon} width="20" height="20" />
					</IconButton>
				</div>
			{/each}
		{/if}
		{#if canAdd}
			<div>
				<Button class="guest-add-button gap-2" onclick={addGuest}>
					<Icon icon={boldPlusIcon} width="20" height="20" />
					{addLabel}
				</Button>
			</div>
		{/if}
	</div>
</fieldset>

<style>
	.guest-email-row :global(.icon-button) {
		width: 44px;
		height: 44px;
		border: 0;
		border-radius: 10px;
	}

	.guest-email-row :global(.tone-danger:not(:disabled)) {
		background: rgb(var(--error) / 0.14);
		color: rgb(var(--error));
	}
</style>
