<script lang="ts">
	import Button from '$lib/components/ui/Button.svelte';
	import Dialog from '$lib/components/ui/Dialog.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import type { ManagedUser, UserDeletionImpact } from '$lib/types';

	let {
		open,
		user,
		impact,
		candidates,
		confirming = false,
		onconfirm,
		oncancel
	}: {
		open: boolean;
		user: ManagedUser;
		impact: Extract<UserDeletionImpact, { requiresReassignment: true }>;
		candidates: ManagedUser[];
		confirming?: boolean;
		onconfirm: (newHostEmail: string) => void;
		oncancel: () => void;
	} = $props();

	let newHostEmail = $state('');
	const options = $derived(
		candidates.map((candidate) => ({
			value: candidate.email,
			label: `${candidate.fullName ? `${candidate.fullName} — ` : ''}${candidate.email}${candidate.googleConnected ? ' · Google Calendar connected' : ' · Email only'}`
		}))
	);

	$effect(() => {
		if (open && !newHostEmail && candidates.length > 0) newHostEmail = candidates[0].email;
	});

	function localDate(value: string) {
		return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value));
	}

	function cancel() {
		if (!confirming) oncancel();
	}
</script>

<Dialog {open} wide label="Reassign upcoming meetings" oncancel={cancel}>
	<h2 class="text-xl font-semibold tracking-tight">Reassign upcoming meetings?</h2>
	<p class="mt-3 text-sm leading-6">
		{user.email} is the only required host for {impact.futureBookingCount} upcoming
		{impact.futureBookingCount === 1 ? 'booking' : 'bookings'}{impact.earliestBookingAt === impact.latestBookingAt
			? `, scheduled for ${localDate(impact.earliestBookingAt)}.`
			: `, from ${localDate(impact.earliestBookingAt)} to ${localDate(impact.latestBookingAt)}.`}
	</p>
	<div class="mt-5 border-2 border-black p-4 text-sm leading-6" role="note">
		The selected user will replace this host on the affected event types and bookings. Existing Google Calendar
		events will be removed from {user.email}'s calendar and added to the selected user's calendar when Google
		Calendar is connected.
	</div>
	<div class="mt-5">
		<Select id="replacement-host" label="New required host" {options} bind:value={newHostEmail} disabled={confirming} />
	</div>
	<div class="mt-6 flex flex-wrap justify-end gap-2">
		<Button variant="secondary" class="dialog-cancel" disabled={confirming} onclick={cancel}>Cancel</Button>
		<Button variant="danger" class="dialog-confirm" disabled={confirming || !newHostEmail} onclick={() => onconfirm(newHostEmail)}>
			{confirming ? 'Reassigning and deleting…' : 'Reassign and delete user'}
		</Button>
	</div>
</Dialog>
