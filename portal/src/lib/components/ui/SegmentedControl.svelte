<script lang="ts">
	import Icon, { type IconifyIcon } from '@iconify/svelte';
	let {
		options,
		value,
		label,
		onchange
	}: {
		options: { value: string; label: string; icon?: IconifyIcon; suffix?: string }[];
		value: string;
		label: string;
		onchange: (value: string) => void;
	} = $props();
</script>

<div class="segmented-control" role="group" aria-label={label}>
	{#each options as option (option.value)}
		<button
			type="button"
			class="segmented-control-option"
			class:on={value === option.value}
			aria-pressed={value === option.value}
			onclick={() => onchange(option.value)}
		>
			{#if option.icon}<Icon icon={option.icon} width="16" height="16" />{/if}
			<span>{option.label}</span>
			{#if option.suffix}<span class="segmented-control-suffix">{option.suffix}</span>{/if}
		</button>
	{/each}
</div>

<style>
	.segmented-control {
		display: inline-flex;
		flex-shrink: 0;
		align-items: center;
		gap: 0.25rem;
		border: 1px solid var(--color-border);
		border-radius: 999px;
		padding: 0.25rem;
		background: rgb(var(--color-primary));
	}

	.segmented-control-option {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		border: 0;
		border-radius: 999px;
		padding: 0.4rem 0.9rem;
		background: transparent;
		color: rgb(var(--color-background));
		font-size: 0.8125rem;
		font-weight: 400;
		cursor: pointer;
		transition: color 0.15s, background 0.15s;
	}

	.segmented-control-suffix {
		opacity: 0.45;
		font-variant-numeric: tabular-nums;
	}

	.segmented-control-option:hover:not(.on) {
		background: rgb(var(--color-background));
		color: rgb(var(--color-primary));
		box-shadow: 0 1px 3px rgb(0 0 0 / 0.12);
	}

	.segmented-control-option.on {
		background: rgb(var(--color-background));
		color: rgb(var(--color-text));
		box-shadow: 0 0 0 1px var(--color-border), 0 1px 3px rgb(0 0 0 / 0.12);
	}
</style>
