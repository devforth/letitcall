<script lang="ts">
	import Icon, { type IconifyIcon } from '@iconify/svelte';
	let {
		options,
		value,
		label,
		iconOnly = false,
		onchange
	}: {
		options: { value: string; label: string; icon?: IconifyIcon; suffix?: string }[];
		value: string;
		label: string;
		iconOnly?: boolean;
		onchange: (value: string) => void;
	} = $props();

	const selectedIndex = $derived(Math.max(options.findIndex((option) => option.value === value), 0));
	const controlPadding = $derived(iconOnly ? 0.1875 : 0.25);
	const controlGap = 0.25;
	const indicatorWidth = $derived(
		`calc(${100 / options.length}% - ${(2 * controlPadding + (options.length - 1) * controlGap) / options.length}rem)`
	);
	const indicatorLeft = $derived(
		`calc(${(selectedIndex * 100) / options.length}% + ${controlPadding * (1 - (2 * selectedIndex) / options.length) + (selectedIndex * controlGap) / options.length}rem)`
	);
</script>

<div
	class="segmented-control"
	class:icon-only={iconOnly}
	role="group"
	aria-label={label}
	style:grid-template-columns={`repeat(${options.length}, minmax(0, 1fr))`}
>
	<span class="segmented-control-indicator" style:left={indicatorLeft} style:width={indicatorWidth}></span>
	{#each options as option (option.value)}
		<button
			type="button"
			class="segmented-control-option"
			class:on={value === option.value}
			class:icon-only={iconOnly}
			aria-pressed={value === option.value}
			aria-label={iconOnly ? option.label : undefined}
			onclick={() => onchange(option.value)}
		>
			{#if option.icon}<Icon icon={option.icon} width="16" height="16" />{/if}
			{#if !iconOnly}<span>{option.label}</span>{/if}
			{#if option.suffix}<span class="segmented-control-suffix">{option.suffix}</span>{/if}
		</button>
	{/each}
</div>

<style>
	.segmented-control {
		display: inline-grid;
		flex-shrink: 0;
		align-items: center;
		position: relative;
		gap: 0.25rem;
		border: 1px solid var(--color-border);
		border-radius: 999px;
		padding: 0.25rem;
		background: rgb(var(--color-primary));
	}

	.segmented-control-option {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		position: relative;
		z-index: 1;
		gap: 0.35rem;
		border: 0;
		border-radius: 999px;
		padding: 0.4rem 0.9rem;
		background: transparent;
		color: rgb(var(--color-background));
		font-size: 0.8125rem;
		font-weight: 400;
		cursor: pointer;
		white-space: nowrap;
		transition: color 0.2s;
	}

	.segmented-control-indicator {
		position: absolute;
		top: 0.25rem;
		bottom: 0.25rem;
		border-radius: 999px;
		background: rgb(var(--color-background));
		box-shadow: 0 0 0 1px var(--color-border), 0 1px 3px rgb(0 0 0 / 0.12);
		transition: left 0.25s ease, width 0.25s ease;
	}

	.segmented-control-option.icon-only {
		justify-content: center;
		width: 1.25rem;
		height: 1.25rem;
		padding: 0;
	}

	.segmented-control-option.icon-only :global(svg) {
		width: 0.875rem;
		height: 0.875rem;
	}

	.segmented-control.icon-only {
		padding: 0.1875rem;
	}

	.segmented-control.icon-only .segmented-control-indicator {
		top: 0.1875rem;
		bottom: 0.1875rem;
	}

	.segmented-control-suffix {
		opacity: 0.45;
		font-variant-numeric: tabular-nums;
	}

	.segmented-control-option:hover:not(.on) {
		background: rgb(var(--color-background));
		color: rgb(var(--color-primary));
		font-weight: 500;
		letter-spacing: -0.012em;
		box-shadow: 0 1px 3px rgb(0 0 0 / 0.12);
	}

	.segmented-control-option.on {
		color: rgb(var(--color-text));
		font-weight: 500;
		letter-spacing: -0.012em;
	}

	:global(html.dark) .segmented-control-option {
		font-weight: 500;
		letter-spacing: -0.012em;
	}

	:global(html.dark) .segmented-control-option.on,
	:global(html.dark) .segmented-control-option:hover:not(.on) {
		font-weight: 400;
		letter-spacing: normal;
	}
</style>
