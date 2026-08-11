<script lang="ts">
	import Icon from '@iconify/svelte';
	import notesIcon from '@iconify-icons/tabler/align-left';
	import pencilIcon from '@iconify-icons/mdi/edit';
	import userIcon from '@iconify-icons/tabler/user';
	import usersIcon from '@iconify-icons/tabler/users';

	let {
		dateLabel,
		timeLabel,
		timezone,
		attendeeName,
		attendeeEmail,
		guestEmails,
		notes,
		attendeeLabel = 'Attendee (you)',
		editable = false,
		onChangeSchedule,
		onChangeDetails
	}: {
		dateLabel: string;
		timeLabel: string;
		timezone: string;
		attendeeName: string;
		attendeeEmail: string;
		guestEmails: string[];
		notes?: string;
		attendeeLabel?: string;
		editable?: boolean;
		onChangeSchedule?: () => void;
		onChangeDetails?: () => void;
	} = $props();
</script>

<div class="booking-details" class:booking-details-static={!editable}>
	<section class="booking-schedule">
		<p class="booking-date">{dateLabel}</p>
		<p class="booking-time">{timeLabel}</p>
		<p class="booking-muted">{timezone}</p>
		{#if editable}
			<button type="button" class="booking-change" aria-label="Change schedule" title="Change schedule" onclick={onChangeSchedule}>
				<Icon icon={pencilIcon} width="22" height="22" />
			</button>
		{/if}
	</section>
	<div class="booking-contact-details">
		<section>
			<p class="booking-label"><span class="booking-label-icon" aria-hidden="true"><Icon icon={userIcon} width="16" height="16" /></span>{attendeeLabel}</p>
			<p class="booking-value">{attendeeName} · {attendeeEmail}</p>
		</section>
		{#if guestEmails.length > 0}
			<section>
				<p class="booking-label"><span class="booking-label-icon" aria-hidden="true"><Icon icon={usersIcon} width="16" height="16" /></span>Guests · {guestEmails.length}</p>
				<ul class="booking-list">
					{#each guestEmails as email (email)}
						<li>{email}</li>
					{/each}
				</ul>
			</section>
		{/if}
		{#if notes}
			<section>
				<p class="booking-label"><span class="booking-label-icon" aria-hidden="true"><Icon icon={notesIcon} width="16" height="16" /></span>Notes</p>
				<p class="booking-notes">{notes}</p>
			</section>
		{/if}
		{#if editable}
			<button type="button" class="booking-change" aria-label="Change contact information" title="Change contact information" onclick={onChangeDetails}>
				<Icon icon={pencilIcon} width="22" height="22" />
			</button>
		{/if}
	</div>
</div>

<style>
	.booking-details {
		display: grid;
		grid-template-columns: minmax(13rem, 0.8fr) minmax(0, 1.2fr);
		overflow: hidden;
		margin-top: 1rem;
		border: 1px solid rgb(var(--color-border));
		border-radius: 1rem;
	}

	.booking-schedule {
		position: relative;
		padding: 1.25rem 1.25rem 5rem;
		border-right: 1px solid rgb(var(--color-border));
	}

	.booking-details-static .booking-schedule,
	.booking-details-static .booking-contact-details {
		padding-bottom: 1.25rem;
	}

	.booking-date {
		font-size: 1.5rem;
		font-weight: 600;
		line-height: 1.25;
	}

	.booking-time {
		margin-top: 0.25rem;
		font-size: 1rem;
		font-weight: 600;
		line-height: 1.5rem;
	}

	.booking-muted {
		font-size: 0.8125rem;
		overflow-wrap: anywhere;
		color: rgb(var(--color-text) / 0.62);
	}

	.booking-contact-details {
		position: relative;
		display: grid;
		align-content: start;
		gap: 1.5rem;
		padding: 1.25rem 1.25rem 5rem;
	}

	.booking-contact-details section {
		padding: 0;
	}

	.booking-label {
		display: flex;
		align-items: center;
		gap: 0.375rem;
		font-size: 0.75rem;
		font-weight: 400;
		color: rgb(var(--color-text) / 0.6);
	}

	.booking-label-icon {
		display: grid;
		place-items: center;
		color: inherit;
	}

	.booking-value,
	.booking-list,
	.booking-notes {
		margin: 0.25rem 0 0 1.4rem;
		font-size: 1rem;
		font-weight: 400;
		line-height: 1.5rem;
		color: rgb(var(--color-text));
		overflow-wrap: anywhere;
	}

	.booking-list {
		display: grid;
		gap: 0.125rem;
		padding: 0;
		list-style: none;
	}

	.booking-notes {
		line-height: 1.5;
		white-space: pre-wrap;
	}

	.booking-change {
		display: grid;
		position: absolute;
		right: 1.25rem;
		bottom: 1.25rem;
		place-items: center;
		width: 3rem;
		height: 3rem;
		border: 0;
		border-radius: 0.875rem;
		background: rgb(var(--color-primary) / 0.12);
		padding: 0;
		color: rgb(var(--color-primary));
		cursor: pointer;
		opacity: 0;
		transition: background 0.2s ease, opacity 0.2s ease;
	}

	.booking-change:hover {
		background: rgb(var(--color-primary) / 0.2);
	}

	.booking-schedule:hover .booking-change,
	.booking-schedule:focus-within .booking-change,
	.booking-contact-details:hover .booking-change,
	.booking-contact-details:focus-within .booking-change {
		opacity: 1;
	}

	@media (hover: none) {
		.booking-change {
			opacity: 1;
		}
	}

	@media (max-width: 640px) {
		.booking-details {
			grid-template-columns: minmax(0, 1fr);
		}

		.booking-schedule {
			border-right: 0;
			border-bottom: 1px solid rgb(var(--color-border));
		}
	}
</style>
