<script lang="ts">
	import { beforeNavigate, goto } from '$app/navigation';
	import ConfirmationDialog from '$lib/components/ui/ConfirmationDialog.svelte';

	let { changed, title, description }: { changed: boolean; title: string; description: string } = $props();

	let target = $state<URL | null>(null);
	let leaving = false;

	beforeNavigate((navigation) => {
		if (leaving || !changed) return;
		navigation.cancel();
		if (!navigation.willUnload) target = navigation.to!.url;
	});

	export function leave(url: string | URL) {
		leaving = true;
		return goto(url);
	}
</script>

{#if target}
	<ConfirmationDialog
		open
		{title}
		{description}
		confirmLabel="Leave"
		cancelLabel="Stay"
		onconfirm={() => void leave(target!)}
		oncancel={() => (target = null)}
	/>
{/if}
