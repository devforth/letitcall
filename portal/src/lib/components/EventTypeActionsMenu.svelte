<script lang="ts">
	import Icon from '@iconify/svelte';
	import dotsVerticalIcon from '@iconify-icons/tabler/dots-vertical';

	let {
		name,
		deleting = false,
		ondelete
	}: {
		name: string;
		deleting?: boolean;
		ondelete: () => void;
	} = $props();

	let open = $state(false);

	function deleteEventType() {
		open = false;
		ondelete();
	}
</script>

<details class="relative" bind:open>
	<summary
		class="event-type-actions-trigger grid size-10 cursor-pointer list-none place-items-center rounded-[10px] transition focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2"
		aria-label={`Actions for ${name}`}
		title="More actions"
	>
		<Icon icon={dotsVerticalIcon} width="22" height="22" />
	</summary>
	<div class="absolute right-0 z-10 mt-2 w-48 rounded-lg border-2 p-2 shadow-[var(--shadow-small)]" style="background: rgb(var(--color-foreground)); border-color: rgb(var(--color-border));">
		<button
			type="button"
			class="delete-action min-h-11 w-full rounded-md px-3 py-2 text-left text-sm font-semibold disabled:cursor-not-allowed disabled:opacity-40"
			disabled={deleting}
			onclick={deleteEventType}
		>
			{deleting ? 'Deleting…' : 'Delete event type'}
		</button>
	</div>
</details>

<style>
	.event-type-actions-trigger {
		color: rgb(var(--color-text));
		--tw-ring-color: rgb(var(--color-primary));
	}

	.event-type-actions-trigger:hover {
		background: rgb(var(--error) / 0.14);
		color: rgb(var(--error));
	}

	.delete-action {
		background: transparent;
		color: rgb(var(--error));
		cursor: pointer;
	}

	.delete-action:hover:not(:disabled) {
		background: rgb(var(--error) / 0.12);
	}
</style>
