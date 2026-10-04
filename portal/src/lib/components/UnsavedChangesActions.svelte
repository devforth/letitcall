<script lang="ts">
	import Icon from '@iconify/svelte';
	import checkIcon from '@iconify-icons/tabler/check';
	import refreshIcon from '@iconify-icons/tabler/refresh';
	import Button from '$lib/components/ui/Button.svelte';
	import StickyActions from '$lib/components/StickyActions.svelte';

	let {
		saving,
		submitLabel,
		savingLabel,
		onrevert
	}: {
		saving: boolean;
		submitLabel: string;
		savingLabel: string;
		onrevert: () => void;
	} = $props();
</script>

<StickyActions>
	<Button variant="primary-outline" class="outlined-action-button" onclick={onrevert} disabled={saving}>
		<span class="flex items-center gap-2">
			<Icon icon={refreshIcon} width="20" height="20" />
			Revert
		</span>
	</Button>
	<Button type="submit" rounded class="primary-action-button" disabled={saving}>
		<span class="flex items-center gap-2">
			<Icon icon={checkIcon} width="20" height="20" />
			{saving ? savingLabel : submitLabel}
		</span>
	</Button>
</StickyActions>
