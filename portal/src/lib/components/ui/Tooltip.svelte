<script lang="ts">
	import infoIcon from '@iconify-icons/tabler/info-circle';
	import Icon from '@iconify/svelte';

	let { text }: { text: string } = $props();
</script>

<span class="tooltip">
	<button class="trigger" type="button" aria-label={text} title={text}>
		<Icon icon={infoIcon} width="16" height="16" />
	</button>
	<span class="content" role="tooltip">{text}</span>
</span>

<style>
	.tooltip {
		position: relative;
		display: inline-flex;
	}

	.trigger {
		display: inline-grid;
		place-items: center;
		padding: 0;
		border: 0;
		background: transparent;
		color: rgb(var(--color-text) / 0.65);
		cursor: help;
	}

	.trigger:focus-visible {
		outline: 2px solid rgb(var(--color-primary));
		outline-offset: 2px;
	}

	.content {
		position: absolute;
		z-index: 20;
		left: 50%;
		bottom: calc(100% + 8px);
		width: max-content;
		max-width: 15rem;
		transform: translateX(-50%);
		padding: 0.5rem 0.625rem;
		border: 1px solid rgb(var(--color-border));
		border-radius: 8px;
		background: rgb(var(--color-foreground));
		box-shadow: var(--shadow-small);
		color: rgb(var(--color-text));
		font-size: 0.75rem;
		font-weight: 400;
		line-height: 1.35;
		opacity: 0;
		visibility: hidden;
		pointer-events: none;
		transition: opacity 0.15s;
	}

	.tooltip:hover .content,
	.trigger:focus + .content {
		opacity: 1;
		visibility: visible;
	}
</style>
