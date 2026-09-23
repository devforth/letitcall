<script lang="ts">
	import Icon from '@iconify/svelte';
	import checkIcon from '@iconify-icons/tabler/check';
	import copyIcon from '@iconify-icons/tabler/copy';
	import loaderIcon from '@iconify-icons/tabler/loader-2';
	import IconButton from '$lib/components/ui/IconButton.svelte';

	let {
		value,
		label,
		class: className = ''
	}: {
		value: string;
		label: string;
		class?: string;
	} = $props();

	let state = $state<'idle' | 'copying' | 'copied'>('idle');
	let resetTimer: ReturnType<typeof setTimeout> | undefined;

	$effect(() => {
		value;
		state = 'idle';
	});

	async function copy() {
		if (resetTimer) clearTimeout(resetTimer);
		state = 'copying';
		await navigator.clipboard.writeText(value);
		state = 'copied';
		resetTimer = setTimeout(() => (state = 'idle'), 2000);
	}
</script>

<IconButton
	class={`copy-button copy-button--${state} ${className}`}
	filled
	label={state === 'idle' ? label : state === 'copying' ? 'Copying…' : 'Copied'}
	onclick={copy}
>
	{#if state === 'copied'}
		<Icon icon={checkIcon} width="22" height="22" />
	{:else if state === 'copying'}
		<Icon icon={loaderIcon} width="16" height="16" class="animate-spin" />
	{:else}
		<Icon icon={copyIcon} width="22" height="22" />
	{/if}
</IconButton>
