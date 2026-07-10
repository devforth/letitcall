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
		class="grid size-11 cursor-pointer list-none place-items-center border border-black bg-white text-black transition hover:bg-black hover:text-white focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2"
		aria-label={`Actions for ${name}`}
		title="More actions"
	>
		<Icon icon={dotsVerticalIcon} width="22" height="22" />
	</summary>
	<div class="absolute right-0 z-10 mt-2 w-48 border border-black bg-white p-2 shadow-[4px_4px_0_0_#000]">
		<button
			type="button"
			class="min-h-11 w-full px-3 py-2 text-left text-sm font-medium underline transition hover:bg-black hover:text-white disabled:cursor-not-allowed disabled:opacity-40"
			disabled={deleting}
			onclick={deleteEventType}
		>
			{deleting ? 'Deleting…' : 'Delete event type'}
		</button>
	</div>
</details>
