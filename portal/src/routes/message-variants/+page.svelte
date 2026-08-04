<script lang="ts">
	const variants = [
		{
			number: 1,
			title: 'Privacy-first',
			recommended: true,
			tags: ['No personal data', 'No management link'],
			message: `Scheduled through LetItCall.

Open LetItCall to view or manage this booking.`
		},
		{
			number: 2,
			title: 'Minimal operational',
			tags: ['No personal data', 'No management link'],
			message: `This booking was created through LetItCall.

Manage the booking securely in LetItCall.`
		},
		{
			number: 3,
			title: 'Formal',
			tags: ['Concise', 'No management link'],
			message: `Booking confirmed.

For booking details or changes, open the booking in LetItCall.`
		},
		{
			number: 4,
			title: 'System-generated',
			tags: ['Formal tone', 'No management link'],
			message: `This is a system-generated booking confirmation from LetItCall.

Use LetItCall to review or update the booking.`
		},
		{
			number: 5,
			title: 'Calendar-focused',
			tags: ['Clear actions', 'No management link'],
			message: `Scheduled via LetItCall.

Attendance can be managed using the calendar controls. Other changes must be made in LetItCall.`
		},
		{
			number: 6,
			title: 'With notes',
			tags: ['Includes invitee note', 'No management link'],
			message: `Scheduled through LetItCall.

Invitee note:
Please prepare the quarterly account summary.

Manage this booking in LetItCall.`
		},
		{
			number: 7,
			title: 'Unverified notes',
			tags: ['Labels user content', 'No management link'],
			message: `Booking confirmed through LetItCall.

Invitee-provided note (unverified):
Please prepare the quarterly account summary.

Open LetItCall to manage the booking.`
		},
		{
			number: 8,
			title: 'Single private link',
			tags: ['Direct action', 'Forwarding risk'],
			warning: true,
			message: `Scheduled through LetItCall.

Manage booking:
https://booking.example.com/manage/••••••

This private link allows booking changes. Do not forward it.`
		},
		{
			number: 9,
			title: 'Authorized-user wording',
			tags: ['Access-aware', 'No management link'],
			message: `Booking confirmed.

Authorized users can review, reschedule, or cancel this booking in LetItCall.`
		},
		{
			number: 10,
			title: 'Compliance-style',
			tags: ['Policy reminder', 'No management link'],
			message: `LetItCall booking confirmation.

Verify booking details and perform any changes through LetItCall. Do not include confidential information in calendar notes.`
		}
	];
</script>

<svelte:head>
	<title>Calendar message variants</title>
</svelte:head>

<main class="variant-page">
	<header class="page-header">
		<div>
			<p class="eyebrow">Temporary design page</p>
			<h1>Calendar description — 10 enterprise-safe variants</h1>
			<p class="intro">
				Each option is shown in the part of the calendar invitation that recipients can read and forward.
			</p>
		</div>
		<div class="recommendation">
			<span class="recommendation-number">1</span>
			<span><strong>Recommended</strong>Privacy-first and least repetitive</span>
		</div>
	</header>

	<section class="variant-grid" aria-label="Calendar description variants">
		{#each variants as variant (variant.number)}
			<article class:recommended={variant.recommended} class="variant-card">
				<header class="card-header">
					<span class="variant-number">{variant.number}</span>
					<div class="variant-title">
						<h2>{variant.title}</h2>
						<div class="tags">
							{#each variant.tags as tag}
								<span class:warning={variant.warning && tag === 'Forwarding risk'}>{tag}</span>
							{/each}
						</div>
					</div>
					{#if variant.recommended}<span class="recommended-label">Recommended</span>{/if}
				</header>

				<div class="calendar-preview">
					<div class="event-summary">
						<span class="calendar-mark" aria-hidden="true">3</span>
						<div>
							<strong>test event</strong>
							<span>Monday, August 3 · 3:30 PM–4:00 PM</span>
						</div>
					</div>
					<p class="message">{variant.message}</p>
				</div>
			</article>
		{/each}
	</section>
</main>

<style>
	:global(body) {
		background: #f4f4f1;
		color: #202124;
	}

	.variant-page {
		width: min(1440px, calc(100% - 40px));
		margin: 0 auto;
		padding: 48px 0 80px;
	}

	.page-header {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 32px;
		padding-bottom: 28px;
		border-bottom: 1px solid #cfcfca;
	}

	.eyebrow {
		margin: 0 0 8px;
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: #686863;
	}

	h1 {
		max-width: 780px;
		margin: 0;
		font-size: clamp(1.75rem, 3vw, 2.5rem);
		line-height: 1.1;
		letter-spacing: -0.03em;
	}

	.intro {
		max-width: 760px;
		margin: 12px 0 0;
		font-size: 0.95rem;
		line-height: 1.5;
		color: #62625d;
	}

	.recommendation {
		display: flex;
		align-items: center;
		gap: 12px;
		min-width: 280px;
		padding: 14px 16px;
		border: 1px solid #202124;
		background: #fff;
		font-size: 0.8rem;
		line-height: 1.35;
	}

	.recommendation strong,
	.recommendation span:last-child {
		display: block;
	}

	.recommendation-number {
		display: grid;
		width: 34px;
		height: 34px;
		flex: 0 0 auto;
		place-items: center;
		border-radius: 50%;
		background: #202124;
		color: #fff;
		font-weight: 700;
	}

	.variant-grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 24px;
		margin-top: 32px;
	}

	.variant-card {
		display: flex;
		min-width: 0;
		flex-direction: column;
		border: 1px solid #d3d3cf;
		background: #fff;
	}

	.variant-card.recommended {
		border: 2px solid #202124;
	}

	.card-header {
		display: flex;
		align-items: flex-start;
		gap: 12px;
		min-height: 82px;
		padding: 18px 20px;
		border-bottom: 1px solid #dededa;
	}

	.variant-number {
		display: grid;
		width: 30px;
		height: 30px;
		flex: 0 0 auto;
		place-items: center;
		border: 1px solid #202124;
		border-radius: 50%;
		font-size: 0.8rem;
		font-weight: 700;
	}

	.recommended .variant-number {
		background: #202124;
		color: #fff;
	}

	.variant-title {
		min-width: 0;
		flex: 1;
	}

	.variant-title h2 {
		margin: 2px 0 8px;
		font-size: 1rem;
		line-height: 1.2;
	}

	.tags {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}

	.tags span,
	.recommended-label {
		padding: 3px 7px;
		border: 1px solid #c8c8c3;
		border-radius: 999px;
		font-size: 0.65rem;
		line-height: 1.2;
		color: #5c5c57;
	}

	.tags span.warning {
		border-style: dashed;
		color: #202124;
		font-weight: 700;
	}

	.recommended-label {
		border-color: #202124;
		background: #202124;
		color: #fff;
		font-weight: 700;
	}

	.calendar-preview {
		flex: 1;
		padding: 22px 24px 28px;
		background: #fafafa;
	}

	.event-summary {
		display: flex;
		align-items: center;
		gap: 12px;
		padding-bottom: 18px;
		border-bottom: 1px solid #dededa;
	}

	.calendar-mark {
		display: grid;
		width: 32px;
		height: 32px;
		flex: 0 0 auto;
		place-items: center;
		border: 1px solid #74746f;
		font-size: 0.8rem;
		font-weight: 700;
	}

	.event-summary strong,
	.event-summary span {
		display: block;
	}

	.event-summary strong {
		margin-bottom: 3px;
		font-size: 0.88rem;
	}

	.event-summary div > span {
		font-size: 0.75rem;
		color: #72726d;
	}

	.message {
		min-height: 120px;
		margin: 18px 0 0;
		font-family: Arial, Helvetica, sans-serif;
		font-size: 0.88rem;
		line-height: 1.5;
		white-space: pre-line;
		overflow-wrap: anywhere;
		color: #3f3f3b;
	}

	@media (max-width: 860px) {
		.variant-page {
			width: min(100% - 24px, 680px);
			padding-top: 28px;
		}

		.page-header {
			align-items: stretch;
			flex-direction: column;
		}

		.recommendation {
			min-width: 0;
		}

		.variant-grid {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 480px) {
		.card-header {
			flex-wrap: wrap;
		}

		.recommended-label {
			margin-left: 42px;
		}

		.calendar-preview {
			padding-inline: 18px;
		}
	}
</style>
