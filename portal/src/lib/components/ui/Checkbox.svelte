<script lang="ts">
	import Icon from '@iconify/svelte';
	import checkIcon from '@iconify-icons/tabler/check';

	let {
		id,
		label,
		checked = $bindable(false),
		disabled = false
	}: { id: string; label: string; checked?: boolean; disabled?: boolean } = $props();
</script>

<label class="checkbox" class:disabled for={id}>
	<input
		{id}
		type="checkbox"
		bind:checked
		{disabled}
		class="control"
	/>
	<span class="box" aria-hidden="true">
		<Icon icon={checkIcon} width="15" height="15" class="check" />
	</span>
	<span class="label">{label}</span>
</label>

<style>
	.checkbox {
		display: inline-flex;
		min-height: 2.75rem;
		align-items: center;
		gap: 0.625rem;
		color: rgb(var(--color-text));
		font-size: 0.875rem;
		cursor: pointer;
	}

	.control {
		position: absolute;
		width: 1px;
		height: 1px;
		margin: -1px;
		overflow: hidden;
		clip: rect(0, 0, 0, 0);
		white-space: nowrap;
	}

	.box {
		display: grid;
		width: 1.25rem;
		height: 1.25rem;
		margin: 3px;
		flex: none;
		place-items: center;
		border-radius: 5px;
		background: rgb(var(--color-foreground));
		box-shadow: inset 0 0 0 2px rgb(var(--color-border));
		transition: background 0.18s, box-shadow 0.18s;
	}

	.box :global(.check) {
		color: rgb(var(--color-primary));
		opacity: 0;
		transform: scale(0.65);
		transition: opacity 0.15s, transform 0.15s;
	}

	.control:checked + .box {
		background: rgb(var(--color-primary) / 0.08);
		box-shadow: inset 0 0 0 2px rgb(var(--color-primary));
	}

	.control:checked + .box :global(.check) {
		opacity: 1;
		transform: scale(1);
	}

	.control:focus-visible + .box {
		box-shadow:
			inset 0 0 0 2px rgb(var(--color-primary)),
			0 0 0 3px rgb(var(--color-primary) / 0.25);
	}

	.checkbox:hover:not(.disabled) .box {
		box-shadow: inset 0 0 0 2px rgb(var(--color-primary));
	}

	.checkbox.disabled {
		cursor: not-allowed;
		opacity: 0.4;
	}

	@media (prefers-reduced-motion: reduce) {
		.box,
		.box :global(.check) {
			transition: none;
		}
	}
</style>
